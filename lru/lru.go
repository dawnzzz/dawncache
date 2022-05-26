package lru

import "container/list"

// Cache LRU 缓存
type Cache struct {
	maxBytes  int64                         // 缓存的最大容量, 0 表示缓存容量不受限制
	nBytes    int64                         // 当前缓存已占用的空间
	ll        *list.List                    // 用于 LRU 的双向链表
	cache     map[string]*list.Element      // 用于保存 key 与双向链表节点地址之间的映射关系
	OnEvicted func(key string, value Value) // 移除数据时执行
}

// entry 保存在双向链表中的条目
type entry struct {
	key   string
	value Value
}

// Value 缓存中value的类型，只要能求出占用空间即可
type Value interface {
	Len() int
}

func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get 获取键值key对应的value
func (c *Cache) Get(key string) (value Value, ok bool) {
	if elem, ok := c.cache[key]; ok {
		// lru 命中
		value = elem.Value.(*entry).value
		// 将节点移动到链头
		c.ll.MoveToFront(elem)
		return value, ok
	}
	// 未命中
	return nil, false
}

// DeleteOldest 删除最近最久未使用的节点 即链表尾部的节点
func (c *Cache) DeleteOldest() {
	elem := c.ll.Back() // 得到尾部节点
	if elem != nil {
		kv := elem.Value.(*entry)
		// 删除尾部节点
		c.ll.Remove(elem)
		// 从cache中删除
		delete(c.cache, kv.key)
		// 减少占用空间
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			// 删除数据时执行
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add 在 lru 缓存中添加数据
func (c *Cache) Add(key string, value Value) {
	if elem, ok := c.cache[key]; ok {
		// 如果缓存里已有键值为key的数据，则更新value
		c.ll.MoveToFront(elem)
		kv := elem.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value

	} else {
		// 没有键值为key的数据，新增一个节点，插入链表头部
		elem := c.ll.PushFront(&entry{key: key, value: value})
		c.nBytes += int64(len(key)) + int64(value.Len())
		c.cache[key] = elem
	}
	// 超出最大空间，删除最久未使用节点
	for c.maxBytes != 0 && c.nBytes > c.maxBytes {
		c.DeleteOldest()
	}
}

// Len 获取已经添加了多少条数据
func (c *Cache) Len() int {
	return c.ll.Len()
}
