/**
 * @Author：Robby
 * @Date：2022/1/9 10:25
 * @Function：
 **/

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	YCache2 "seven-days-projects/YCache/YCache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

// 创建实例
func createGroup() *YCache2.Group {
	return YCache2.NewGroup("scores", 2<<10, YCache2.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

// 启动cache通信的HTTP服务
func startCacheServer(addr string, addrs []string, group *YCache2.Group) {
	// 实例化HTTPPool
	peers := YCache2.NewHTTPPool(addr)
	// 设置cache IP与HTTP信息的对应关系
	peers.Set(addrs...)
	// 将HTTPPool绑定到group中
	group.RegisterPeers(peers)
	log.Println("YCache is running at", addr)
	// 启动服务 阻塞
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 在启动一个对外的HTTP api服务
func startAPIServer(apiAddr string, group *YCache2.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// 获取请求的key
			key := r.URL.Query().Get("key")
			// 查询当前节点的缓存记录，如果当前节点没有，请求其他节点，如果没有就从本地数据库加载
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			// 这里将数据拷贝了一份，避免指针的原因，高并发的时候，数据被修改
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend api server is running at", apiAddr)
	// apiAddr[7:]为：localhost:8001
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "YCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	// cache实例对应的IP地址
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	// 存储所有的cache节点IP
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}
	// 创建本地cache实例
	group := createGroup()
	// 启动api服务器
	if api {
		go startAPIServer(apiAddr, group)
	}
	// 启动cache通信服务器
	startCacheServer(addrMap[port], []string(addrs), group)
}