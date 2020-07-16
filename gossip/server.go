package gossip

import (
	"crypto/sha1"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hashicorp/memberlist"

	"github.com/cjey/gbase/context"
)

type Server struct {
	Config *memberlist.Config

	name string

	mu      sync.Mutex
	shutsig chan struct{}

	memberlist *memberlist.Memberlist
	sender     *Sender
	delegate   Delegate

	// key<bootstrap node addr> => value<online>
	bootstraps map[string]bool

	// key<node name> => value<node>
	peers map[string]*Node
	pmu   sync.RWMutex
}

// NewServer return a gossip server, enable lzw compression default
func NewServer(name, key string) *Server {
	// 若每个region部署3个节点，30个region共计N=90
	// 若每个zone部署3个节点，50个zone共计N=150
	// 所以，log(N+1)的值结果不会超过3
	var cfg = &memberlist.Config{
		Name: name,

		Transport:               nil, // use default
		BindAddr:                "",
		BindPort:                0, // random port, maybe multi instance by overseer
		AdvertiseAddr:           "",
		AdvertisePort:           0,
		ProtocolVersion:         memberlist.ProtocolVersionMax,
		TCPTimeout:              10 * time.Second,        // Timeout after 10 seconds
		IndirectChecks:          3,                       // Use 3 nodes for the indirect ping
		RetransmitMult:          4,                       // Retransmit a message 4 * log(N+1) nodes
		SuspicionMult:           4,                       // Suspect a node for 4 * log(N+1) * Interval
		SuspicionMaxTimeoutMult: 6,                       // For 10k nodes this will give a max timeout of 120 seconds
		PushPullInterval:        30 * time.Second,        // Low frequency
		ProbeTimeout:            1000 * time.Millisecond, // Reasonable RTT time for LAN
		ProbeInterval:           1 * time.Second,         // Failure check every second
		DisableTcpPings:         true,                    // Just UDP
		AwarenessMaxMultiplier:  8,                       // Probe interval backs off to 8 seconds

		GossipNodes:          3,                      // Gossip to 3 nodes
		GossipInterval:       200 * time.Millisecond, // Gossip more rapidly
		GossipToTheDeadTime:  30 * time.Second,       // Same as push/pull
		GossipVerifyIncoming: true,
		GossipVerifyOutgoing: true,

		EnableCompression: true, // Enable compression by default

		SecretKey: nil,
		Keyring:   nil,

		DNSConfigPath: "/etc/resolv.conf",

		HandoffQueueDepth: 1024,
		UDPBufferSize:     1400,
		CIDRsAllowed:      nil, // same as allow all
	}

	if key != "" {
		var hashed = sha1.Sum([]byte("magic=Chaos Is A Ladder&" + key))
		cfg.SecretKey = hashed[:16] // enable AES-128
	}

	return &Server{
		Config: cfg,

		shutsig:    make(chan struct{}),
		bootstraps: make(map[string]bool),
		peers:      make(map[string]*Node),
	}
}

func (s *Server) RegisterDelegate(d Delegate) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.memberlist != nil {
		select {
		case <-s.shutsig:
			return fmt.Errorf("already shutdown")
		default:
			return fmt.Errorf("already serving")
		}
	}
	s.delegate = d
	return nil
}

func (s *Server) Serve(ctx context.Context, bindstr string, bootstrapsstr ...string) error {
	s.mu.Lock()
	var locked = true
	var unlock = func() {
		if locked {
			locked = false
			s.mu.Unlock()
		}
	}
	defer unlock()

	if s.memberlist != nil {
		select {
		case <-s.shutsig:
			return fmt.Errorf("already shutdown")
		default:
			return fmt.Errorf("already serving")
		}
	}

	// check all address, make bootstraps uniq, and remove bind from bootstraps if it has
	bind, bootstraps, err := UniqTCPAddr(bindstr, bootstrapsstr)
	if err != nil {
		return err
	}
	for _, addr := range bootstraps {
		s.bootstraps[addr.String()] = false
	}

	// overwrite config
	var cfg = s.Config
	cfg.BindAddr = bind.IP.String()
	cfg.BindPort = bind.Port
	cfg.AdvertiseAddr = cfg.BindAddr
	cfg.AdvertisePort = cfg.BindPort
	cfg.PushPullInterval = 0 // just disable full state sync

	if cfg.Logger == nil {
		var l = newLogWriter(ctx, _LOG_LEVEL_WARN)
		cfg.Logger = log.New(l, "", 0)
	}

	// set delegate
	var dgm = newDelegateM(s)
	cfg.Delegate = dgm
	cfg.Events = dgm
	cfg.Ping = dgm

	// create server use retry policy
	var ecnt = 0
	for {
		var srv, err = memberlist.Create(cfg)
		if err == nil {
			s.name = cfg.Name
			s.memberlist = srv
			break
		}
		ecnt++
		if ecnt >= 10 { // total retry 10 times, total wait 45s
			return fmt.Errorf("gossip server serve failed after creating memberlist 10 times, %w", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.shutsig:
			return nil
		case <-time.After(time.Duration(ecnt) * time.Second):
		}
	}
	unlock()

	dgm.start()

	if len(s.bootstraps) > 0 {
		go s.keepBootstrapsOnline()
	}

	go func() {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			defer s.mu.Unlock()
			select {
			case <-s.shutsig:
			default:
				close(s.shutsig)
			}
		case <-s.shutsig:
		}
	}()

	// wait shutdown signal
	<-s.shutsig
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-s.shutsig:
		return nil
	default:
	}

	defer func() {
		close(s.shutsig)
	}()

	var timeout = 5 * time.Second
	if d, ok := ctx.Deadline(); ok {
		timeout = time.Until(d)
	}

	// first, broadcast Leave message
	if err := s.memberlist.Leave(timeout); err != nil {
		return err
	}
	// then shutdown
	return s.memberlist.Shutdown()
}

func (s *Server) nodeOnline(node *Node) {
	var addr = node.Address()

	if _, ok := s.bootstraps[addr]; ok {
		// bootstrap node online
		s.bootstraps[addr] = true
	}

	if node.Name == s.name {
		node.RTT = 0
	}

	s.pmu.Lock()
	s.peers[node.Name] = node
	s.pmu.Unlock()
}

func (s *Server) nodeOffline(node *Node) {
	var addr = node.Address()

	if _, ok := s.bootstraps[addr]; ok {
		// bootstrap node offline
		s.bootstraps[addr] = false
	}

	s.pmu.Lock()
	delete(s.peers, node.Name)
	s.pmu.Unlock()
}

func (s *Server) nodeUpdate(node *Node) {
	s.pmu.Lock()
	s.peers[node.Name] = node
	s.pmu.Unlock()
}

func (s *Server) nodePing(node *Node, rtt time.Duration) {
	s.pmu.Lock()
	if peer := s.peers[node.Name]; peer != nil {
		peer.RTT = rtt
	}
	s.pmu.Unlock()
}

func (s *Server) Peer(name string) *Node {
	s.pmu.RLock()
	defer s.pmu.RUnlock()
	return s.peers[name]
}

func (s *Server) Peers() []*Node {
	s.pmu.RLock()
	defer s.pmu.RUnlock()
	var peers = make([]*Node, 0, len(s.peers))
	for _, p := range s.peers {
		peers = append(peers, p)
	}
	return peers
}

func (s *Server) keepBootstrapsOnline() {
	for {
		var offlines = make([]string, 0)
		for addr, online := range s.bootstraps {
			if !online {
				offlines = append(offlines, addr)
			}
		}
		if len(offlines) > 0 {
			s.memberlist.Join(offlines)
		}

		select {
		case <-s.shutsig:
			return
		case <-time.After(5 * time.Second):
		}
	}
}
