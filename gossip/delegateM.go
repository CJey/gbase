package gossip

import (
	"time"

	"github.com/hashicorp/memberlist"
)

type delegateM struct {
	dg     Delegate
	srv    *Server
	sender *Sender
}

func newDelegateM(srv *Server) *delegateM {
	var d = &delegateM{
		dg:  srv.delegate,
		srv: srv,
	}
	d.sender = newSender(srv, d)
	return d
}

func (d *delegateM) start() {
	if d.dg != nil {
		d.dg.GossipStarted(d.sender)
	}
}

// <Delegate>，通常来说，如果不实现此代理则服务没有什么意义

// 本Node的有限大小元数据(默认512字节)，在广播alive时会提供给接收节点
// 本方法在节点刚启动时，以及主动调用server的UpdateNode方法时才被调用到
// 即server假定metadata不会改变，如果发生了改变，必须主动调用UpdateNode方法来触发一次cluster范围的metadata更新
func (d *delegateM) NodeMeta(limit int) []byte {
	if d.dg == nil {
		return nil
	}
	return d.dg.Metadata(limit)
}

// 当有用户数据过来时，会调用本方法
// msg如果要保存，必须做copy
// 数据可以是广播数据，也可以是某节点直接单播过来的数据
func (d *delegateM) NotifyMsg(msg []byte) {
	if d.dg == nil {
		return
	}

	d.dg.NotifyMessage(msg)
}

// 获取可广播的用户数据，server通过此方法获得和传播增量状态
// 数据总长不得超过limit，并且，每个slice必须开头预留overhead个字节
func (d *delegateM) GetBroadcasts(overhead, limit int) [][]byte {
	if d.dg == nil {
		return nil
	}
	return d.sender.getBroadcasts(overhead, limit)
}

// 获取本机的状态数据，用于和peer交换
// 新节点加入时，或者定期的pullpush发生后，此方法会被调用一次
// 本节点应当通过此方法返回自身的完整状态信息给到发起pullpush的节点
// 虽然可以快速收敛，但成本较高
// 一般不用，不予实现
func (d *delegateM) LocalState(join bool) []byte {
	return nil
}

// 合并peer的状态数据，用于合并其他peer的状态数据
// 当本节点加入集群时，或者定期的pullpush发生后，此方法会被调用一次
// server会将从其他节点获得的完整状态数据传递给此方法
// 此方法应当将得到的状态数据合并到自身的状态中
// 与LocalState配合使用
// 一般不用，不予实现
func (d *delegateM) MergeRemoteState(buf []byte, join bool) {
}

// <EventDelegate>，节点的事件通知

// Node加入后会触发此方法，服务create完成前，必定会触发一次本地node的Join
// 随后，有新节点加入时才会触发一次
func (d *delegateM) NotifyJoin(peer *memberlist.Node) {
	var node = newNode(peer)
	d.srv.nodeOnline(node)
	if d.dg == nil {
		return
	}
	d.dg.NotifyJoin(node)
	go d.sender.Ping(node.Name)
}

// 有Node离开时会触发一次此方法
// 主动离开会立即触发，本节点如果检测发现某节点无法ping通，也收不到alive广播，则自动认为该节点离开，并也会触发一次此方法
func (d *delegateM) NotifyLeave(peer *memberlist.Node) {
	var node = d.srv.Peer(peer.Name)
	if node == nil {
		return
	}
	d.srv.nodeOffline(node)
	if d.dg == nil {
		return
	}
	d.dg.NotifyLeave(node)
}

// Node有更新时会触发一次此方法（调用服务的Update方法，主要用途是更新节点的Metadata）
func (d *delegateM) NotifyUpdate(peer *memberlist.Node) {
	var node = d.srv.Peer(peer.Name)
	if node == nil {
		return
	}
	var rtt = node.RTT
	var newNode = newNode(peer)
	newNode.RTT = rtt
	d.srv.nodeUpdate(newNode)
	if d.dg == nil {
		return
	}
	d.dg.NotifyUpdate(newNode)
}

// <PingDelegate>
// 节点之间互ping，可以借此传播一些标志信息，以及探测节点之间的rtt
// 为了让rtt有意义，tcp ping和indirect ping都不会触发此委托
// 1. ping可以通过payload推送数据
// 2. ping会触发双向互ping，可用于某些关键控制数据的交换

// 主动发起方拿出自己的AckPayload，向目标发出一个ping消息
func (d *delegateM) AckPayload() []byte {
	if d.dg == nil {
		return nil
	}
	return d.dg.PingPayload()
}

// 被动方收到主动方推送过来的payload
// 随后，被动方也会向主动方再主动发起一个ping消息作为回应
func (d *delegateM) NotifyPingComplete(peer *memberlist.Node, rtt time.Duration, payload []byte) {
	var node = d.srv.Peer(peer.Name)
	if node == nil {
		return
	}
	d.srv.nodePing(node, rtt)
	if d.dg == nil {
		return
	}
	d.dg.NotifyPing(node, payload)
}

// <AliveDelegate>
// 节点发送alive广播后便触发此委托，服务create完成前，会触发一次本地node的Alive
// alive广播在生命周期内可能会发送多次，当有节点加入时，可能会连续广播多次
// 一般而言，可以不用关注alive广播，只通过Join和Leave事件来关注节点的情况即可

// 返回值非nil，则表示直接忽略该节点的alive广播，也等于是禁止该节点加入本server
// 特别的用法上(一般不用)，可以通过此机制，来在一个广播域中划分子域(类似同一个交换机下，跑多个网段)
// 一般不用，不予实现
func (d *delegateM) NotifyAlive(peer *memberlist.Node) error {
	return nil
}

// <MergeDelegate>

// 当有Node执行了tcp push/pull，发现有新节点加入需要决策是否允许merge node列表时，此委托便会被触发
// 如果返回值非nil，则表示不同意合并这些节点的信息，这样也不会触发后续的MergeRemoteState方法
// 只有在节点join前会调用一次，同意或者拒绝之后，将不再调用
// 注意，即使返回值非nil，也不表示该节点不在集群中，仅仅表示不接受该节点的full state交换
// 一般不用，不予实现
func (d *delegateM) NotifyMerge(peers []*memberlist.Node) error {
	return nil
}

// <ConflictDelegate>

// 如果两个不同的Node宣告的Name却相同，可以触发此委托
// 对于这两个Node本身，检测到后，会触发一次此委托，后续不再触发
// 而第三方节点在新冲突节点full state时，触发一次alive，并触发一次conflict，并且以本地已有的node为准
// 而当旧节点leave后，第三方节点经过几个周期的conflict后，将接纳新节点占用此Name
// 如果网络中已经存在Name冲突，新加入的节点只会认可先join到自己的那个
// 一般不用，不予实现
func (d *delegateM) NotifyConflict(existing, other *memberlist.Node) {
}
