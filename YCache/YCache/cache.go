/**
 * @Author：Robby
 * @Date：2022/1/9 02:06
 * @Function：
 **/

package YCache

import (
	"seven-days-projects/YCache/YCache/lru"
	"sync"
)

// cacheInstance cache实例，在lru算法的基础上封装了mutex互斥锁，解决办法问题
type cacheInstance struct {
	mu         sync.Mutex
	cache      *lru.Cache
	cacheBytes int64  // cacheInstance最大占用内存
}

// Add 封装并发控制
func (c *cacheInstance) Add(key string, value *ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil { // Lazy Initialization 延时初始化
		c.cache = lru.NewCache(c.cacheBytes, nil)
	}
	// 添加记录, value必须实现Value接口的所有方法
	c.cache.Add(key, value)
}

func (c *cacheInstance) GetValue(key string) (value *ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache == nil {
		return
	}
	// 获取记录
	if v, ok := c.cache.GetValue(key); ok {
		return v.(*ByteView), ok
	}

	return
}
