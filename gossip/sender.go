package gossip

import (
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
)

type broadcast []byte

func (b broadcast) Invalidates(memberlist.Broadcast) bool {
	return false
}

func (b broadcast) Message() []byte {
	return b
}

func (b broadcast) Finished() {
	b = nil
}

type Sender struct {
	srv *Server
	dgm *delegateM

	queue *memberlist.TransmitLimitedQueue
}

func newSender(srv *Server, dgm *delegateM) *Sender {
	return &Sender{
		srv: srv,
		dgm: dgm,

		queue: &memberlist.TransmitLimitedQueue{
			NumNodes: func() int {
				return 1
			},
		},
	}
}

func (s *Sender) Self() *Node {
	if s == nil {
		return nil
	}
	return s.srv.Peer(s.srv.name)
}

func (s *Sender) Name() string {
	if s == nil {
		return ""
	}
	return s.srv.name
}

func (s *Sender) UpdateMetadata(timeout time.Duration) error {
	if s == nil {
		return nil
	}
	return s.srv.memberlist.UpdateNode(timeout)
}

func (s *Sender) Ping(name string) (time.Duration, error) {
	if s == nil {
		return 0, nil
	}
	s.srv.pmu.RLock()
	var peer = s.srv.peers[name]
	s.srv.pmu.RUnlock()
	if peer == nil {
		return 0, fmt.Errorf("no route to host")
	}
	var addr, _ = net.ResolveUDPAddr("udp", peer.Address())
	var rtt, err = s.srv.memberlist.Ping(peer.Name, addr)
	if err != nil {
		peer.RTT = rtt
	}
	return rtt, err
}

func (s *Sender) Peer(name string) *Node {
	if s == nil {
		return nil
	}
	return s.srv.Peer(name)
}

func (s *Sender) Peers() []*Node {
	if s == nil {
		return nil
	}
	return s.srv.Peers()
}

func (s *Sender) SendBestEffort(name string, msg []byte) {
	if s == nil {
		return
	}
	s.srv.pmu.RLock()
	var peer = s.srv.peers[name]
	s.srv.pmu.RUnlock()
	if peer != nil {
		s.srv.memberlist.SendBestEffort(peer.node, msg)
	}
}

func (s *Sender) SendReliable(name string, msg []byte) {
	if s == nil {
		return
	}
	s.srv.pmu.RLock()
	var peer = s.srv.peers[name]
	s.srv.pmu.RUnlock()
	if peer != nil {
		s.srv.memberlist.SendReliable(peer.node, msg)
	}
}

func (s *Sender) Broadcast(msg []byte) {
	if s == nil {
		return
	}
	s.queue.QueueBroadcast(broadcast(msg))
}

func (s *Sender) getBroadcasts(overhead, limit int) [][]byte {
	return s.queue.GetBroadcasts(overhead, limit)
}
