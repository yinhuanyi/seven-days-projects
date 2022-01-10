/**
 * @Author：Robby
 * @Date：2022/1/8 13:45
 * @Function：
 **/

package lru

import (
	"reflect"
	"testing"
)

// 类型别名作用是：实现Len方法
type String string

func (d String) Len() int {
	return len(d)
}

// TestGetValue 测试数据写入和获取
func TestGetValue(t *testing.T) {
	cache := NewCache(int64(0) , nil) // 初始化cache, int64(0)表示cache内存大小不限制
	cache.Add("key1", String("1234"))     // 添加数据
	if v, ok := cache.GetValue("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := cache.GetValue("key2"); !ok {
		t.Fatalf("cache miss key2 failed")
	}
}

// TestRemoveOldest 测试cache满了，是否会执行lru算法
func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	cache := NewCache(int64(cap), nil)
	cache.Add(k1, String(v1))
	cache.Add(k2, String(v2))
	cache.Add(k3, String(v3))
	// 当添加了k3，基于lru算法会移除k1
	if _, ok := cache.GetValue("key1"); ok || cache.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

// TestOnEvicted 测试删除记录，回调函数是否被调用
func TestOnEvicted(t *testing.T) {
	// keys保存cache淘汰的key
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	cache := NewCache(int64(10), callback)
	cache.Add("key1", String("123456"))
	cache.Add("k2", String("k2"))
	cache.Add("k3", String("k3"))
	cache.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}