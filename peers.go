package DawnCache

import pb "DawnCache/dawncachepb"

// PeerPicker 选取节点的接口
type PeerPicker interface {
	// PickPeer 根据 key 选择相应的 PeerGetter 获取数据
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 远程获取数据的接口
type PeerGetter interface {
	// Get 根据 groupName 和 key 获取源数据
	Get(in *pb.Request, out *pb.Response) error
}
