package gossip

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/hashicorp/memberlist"
)

type Node struct {
	node *memberlist.Node

	Name string
	Addr net.IP
	Port uint16
	Meta []byte
	RTT  time.Duration
}

func newNode(node *memberlist.Node) *Node {
	return &Node{
		node: node,

		Name: node.Name,
		Addr: node.Addr,
		Port: node.Port,
		Meta: node.Meta,
		RTT:  -1,
	}
}

func (n *Node) Address() string {
	return net.JoinHostPort(n.Addr.String(), strconv.Itoa(int(n.Port)))
}

func (n *Node) String() string {
	if n.Name != "" {
		return fmt.Sprintf("%s (%s)", n.Name, n.Address())
	}
	return n.Address()
}

type Delegate interface {
	// GossipStarted give the sender to delegate
	// announce gossip server started
	GossipStarted(*Sender)

	// Metadata get local node meta data
	Metadata(limit int) []byte
	// NotifyJoin notify node join or node online
	NotifyJoin(*Node)
	// NotifyJoin notify node leave or node offline
	NotifyLeave(*Node)
	// NotifyUpdate notify node's metadata updated
	NotifyUpdate(*Node)

	// PingPayload given the payload push to other node while pinging.
	// And payload will be compress by lzw default, then aes encrypt optionally,
	// the final size must less than udp packet max size limit
	PingPayload() []byte
	// NotifyPing notify that other node pushed it's payload to me while pinging
	NotifyPing(other *Node, payload []byte)

	// NotifyMessage notify received message
	NotifyMessage(msg []byte)
}
