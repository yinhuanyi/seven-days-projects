/**
 * @Author：Robby
 * @Date：2022/1/9 02:12
 * @Function：
 **/

package singleflight

import "sync"

// call 代表正在进行中，或已经结束的请求。使用 sync.WaitGroup 锁避免重入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 是 singleflight 的主数据结构，管理不同 key 的请求(call)
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 的作用就是，针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 如果当前key请求已近存在
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait() // 如果请求正在进行中，则等待
		return c.val, c.err
	}
	// 如果是第一次请求key
	c := new(call)
	c.wg.Add(1) // 让其他的请求等待
	g.m[key] = c // 将call写入到m中
	g.mu.Unlock()

	// 获取fn执行的结果，写入到call对象中
	c.val, c.err = fn()
	c.wg.Done() // 第一次请求结束，后续的请求可以直接从c对象中获取到第一次请求结果

	g.mu.Lock()
	delete(g.m, key) // 从map中删除对应的key，以便后续请求可以查询执行fn函数
	g.mu.Unlock()

	return c.val, c.err
}