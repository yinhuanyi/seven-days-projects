/**
 * @Author：Robby
 * @Date：2022/1/9 02:12
 * @Function：
 **/

package lru

import "container/list"

// 双向链表节点数据类型
type entry struct {
	key   string // 这里的key于map中的key是同一个key
	value Value
}

// Value 类型需要实现Len方法，返回链表节点entry的大小
type Value interface {
	Len() int
}

// Cache 缓存节点数据结构
type Cache struct {
	maxBytes int64                          // 节点最大内存
	nbytes   int64                          // 节点当前内存
	ll       *list.List                     // 双向链表
	cache    map[string]*list.Element       // map
	OnEvicted func(key string, value Value) //当链表数据被删除的回调函数
}


// NewCache cache构造函数
func NewCache(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes, // 初始化缓存最大内存占用
		ll:        list.New(), // 链表初始化
		cache:     make(map[string]*list.Element), // map初始化
		OnEvicted: onEvicted,
	}
}

// GetValue 获取cache节点数据
func (c *Cache) GetValue(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok { // 查询map中是否有对应的key，如果存在获取value，也就是链表的元素
		c.ll.MoveToFront(ele)    // 将当前元素移动到队首，为lru算法做铺垫，那么队尾的就是最近最少使用的节点，优先删除
		kv := ele.Value.(*entry) // 获取链表节点数据
		return kv.value, true
	}
	return
}

// RemoveOldest 获取到队尾节点，从链表中删除，这也是
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 获取链表最后一个节点
	if ele != nil {
		c.ll.Remove(ele)                                       // 删除链表最后一个节点
		kv := ele.Value.(*entry)                               // 获取节点的key
		delete(c.cache, kv.key)                                // 删除map中的key
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 重新计算链表的内存大小
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value) // 执行回调函数
		}
	}
}

// Add 新增/修改
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok { // 如果key在map中存在，表示更新cache节点数据
		c.ll.MoveToFront(ele)                                  // 移动当前节点链表队首
		kv := ele.Value.(*entry)                               // 获取当前节点的entry
		c.nbytes += int64(value.Len()) - int64(kv.value.Len()) // 重新计算链表的内存大小
		kv.value = value
	} else { // 如果key在map中不存在，表示添加节点数据
		ele = c.ll.PushFront(&entry{key, value})         // 在链表头部插入新节点
		c.cache[key] = ele                               // 添加map映射
		c.nbytes += int64(len(key)) + int64(value.Len()) // 重新计算链表的内存大小
	}
	// 添加节点的时候需要判断是否超过了cache最大内存，如果超过，需要删除最近最少使用的节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len 获取cache中数据的长度
func (c *Cache) Len() int {
	return c.ll.Len()
}