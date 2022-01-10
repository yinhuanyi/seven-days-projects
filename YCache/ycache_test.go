/**
 * @Author：Robby
 * @Date：2022/1/8 15:29
 * @Function：
 **/

package YCache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

// TestGetter 回调函数测试
func TestGetter(t *testing.T) {
	// 让函数满足接口，调用者必须实现回调函数
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	// f.Get调用的是自己
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// TestGet Group的get方法测试
func TestGet(t *testing.T) {
	// 做计数器
	loadCounts := make(map[string]int, len(db))

	// 创建cache miss的回调函数，因为NewGroup的参数是Getter接口，因此普通函数必须是GetterFunc类型
	fn := GetterFunc(func(key string) ([]byte, error) {
		log.Println("search key from DB", key)
		if v, ok := db[key]; ok { // key存在与DB中
			if _, ok := loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}

		return nil, fmt.Errorf("%s not exist", key)
	})

	// scores是命名空间，2 << 10 表示移位操作，这里是2的11次方，fn是GetterFunc类型的回调函数，因为NewGroup的参数是Getter接口，因此普通函数必须是GetterFunc类型
	g := NewGroup("scores", 2<<10, fn)

	// 遍历所有的key
	for k, v := range db {

		view, err := g.Get(k)
		if err != nil || view.String() != v {
			t.Fatal("failed to get value") // 数据从cache和db都没有获取到
		}else if loadCounts[k] == 1 {
			log.Printf("cache %s miss, get value from db\n", k)
		}else {
			fmt.Println(loadCounts[k])
			log.Printf("get %s from cache\n", k)
		}

	}
}