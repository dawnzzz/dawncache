package DawnCache

import (
	"errors"
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

// load 从别处加载数据
func (g *Group) load(key string) (ByteView, error) {
	// 暂时全部调用回调函数加载key对应的value
	// 从远程调用之后实现
	return g.getLocally(key)
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
