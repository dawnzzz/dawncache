package singleflight

import "sync"

// call 代表一次查询请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu      sync.Mutex       // 对 hashMap 的访问互斥
	hashMap map[string]*call // 保存 key 和请求的映射关系
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.hashMap == nil { // 延迟初始化
		g.hashMap = make(map[string]*call)
	}
	if c, ok := g.hashMap[key]; ok {
		// 已在 hashMap 中记录，等待结果即可
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// 没有在 hashMap 中记录
	// 新建 call 在 hashMap 中记录
	c := new(call)
	c.wg.Add(1)
	g.hashMap[key] = c
	g.mu.Unlock()

	// 远程请求数据
	c.val, c.err = fn()
	c.wg.Done() // 得到数据

	g.mu.Lock()
	delete(g.hashMap, key)
	g.mu.Unlock()

	return c.val, c.err
}
