package DawnCache

import (
	"errors"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string // 一个组的命名空间，用于区分不同的缓存，如学生姓名、成绩可以放到不同的缓存中去
	getter    Getter // 当查找数据未命中时，调用该函数获取值
	mainCache cache  // 底层缓存
	peers     PeerPicker
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group) // 存储所有的缓存
)

// NewGroup 新建一个 *Group 缓存
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g // 添加到 map 中
	return g
}

// GetGroup 根据命名空间返回对应的 Group
func GetGroup(name string) *Group {
	if g, ok := groups[name]; ok {
		return g
	}
	return nil
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}
	// 可以在缓存中查询到，返回数据
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}
	// 从远程或者回调函数获取key对应的value
	return g.load(key)
}

// RegisterPeers 注册 PeerPicker
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		// RegisterPeers 不允许调用超过1次
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// load 从别处加载数据
func (g *Group) load(key string) (ByteView, error) {
	if g.peers != nil {
		// peers 不为空，可以从远程获取数据
		if peer, ok := g.peers.PickPeer(key); ok {
			// 从远程获取数据
			view, err := g.getFromPeer(peer, key)
			if err != nil {
				log.Println("[GeeCache] Failed to get from peer", err)
				return ByteView{}, err
			}
			return view, nil
		}
	}
	// 本地通过回调函数获取数据
	return g.getLocally(key)
}

// getFromPeer 从 peer 处获取数据
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	data, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: data}, nil
}

// getLocally 从本地，即调用回调函数获取 value
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value) // 将新获取到的数据放入缓存中
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
