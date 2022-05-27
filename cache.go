package DawnCache

import (
	"DawnCache/lru"
	"sync"
)

// cache 单机并发缓存，带有互斥锁
type cache struct {
	mu         sync.Mutex // 互斥锁
	lru        *lru.Cache // LRU 缓存
	cacheBytes int64      // 缓存容量
}

// add 向缓存中添加键值对
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}

	c.lru.Add(key, value)
}

// get 查找
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 底层lru缓存为空，直接返回
	if c.lru == nil {
		return
	}
	// 在底层lru缓存中查找
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
