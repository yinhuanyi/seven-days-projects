/**
 * @Author：Robby
 * @Date：2022/1/9 02:12
 * @Function：
 **/

package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)


// Hash 定义hash函数的结构，接收data，返回一个uint32类型的hash值
type Hash func(data []byte) uint32

// Map 这是hash环的主要结构
type Map struct {
	hash     Hash // hash函数
	replicas int  // 一个cache节点对应的虚拟节点个数
	keys     []int // hash环的虚拟节点hash值
	hashMap  map[int]string // 虚拟节点与真实节点的映射表
}

// NewMap Map的构造函数
func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 默认的hash函数是crc32.ChecksumIEEE算法
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 批量cache节点到hash环上，这里的key是节点IP
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 计算虚拟节点的hash值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 将虚拟节点的hash值添加到哈希环上
			m.keys = append(m.keys, hash)
			// 添加虚拟节点hash值 -> 真实cache节点的IP
			m.hashMap[hash] = key
		}
	}
	// 让哈希环上的hash值从小到大排序
	sort.Ints(m.keys)
}

// Get 基于查询的key，获取真实节点的IP值
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	// 计算查询的key对应的hash值
	hash := int(m.hash([]byte(key)))

	// 二分搜索算法，获取最近虚拟节点的索引
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// 由于认为keys一个环形结构，idx==len(m.keys)，那么就是取第一个虚拟节点的hash值，如果idx<len(m.keys)，那么就直接从keys中取虚拟节点的hash值即可，最终从hashMap获取真实cache的IP值
	return m.hashMap[m.keys[idx%len(m.keys)]]
}