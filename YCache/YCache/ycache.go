/**
 * @Author：Robby
 * @Date：2022/1/9 02:06
 * @Function：
 **/

package YCache

import (
	"fmt"
	"log"
	"seven-days-projects/YCache/YCache/singleflight"
	"seven-days-projects/YCache/YCache/ycachepb"
	"sync"
)

// Getter 当cache miss的时候，从哪里获取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 函数类型, 这是一个接口函数，实现接口的同时调用自己，接口型函数只能应用于接口内部只定义了一个方法的情况，让用户传递进来的函数作为回调函数使用，可以将其他函数转换接口定义的函数
type GetterFunc func(key string) ([]byte, error)

// Get 实现Getter接口的Get方法
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group 可以认为是缓存的命名空间
type Group struct {
	name      string        // 空间名称
	getter    Getter        // 获取外部数据的接口，待用户传递进来，实现cache miss的回调函数
	mainCache cacheInstance // cache实例

	// 新增成员变量
	peers PeerPicker // PeerPicker接口的实现体是HTTPPool

	loader *singleflight.Group // 这里是singleflight的Group
}

// groups是一个全局变量，那么在HTTP请求中可以获取到这个groups变量
var (
	mu     sync.RWMutex              // 读写锁，读取不加锁
	groups = make(map[string]*Group) // 创建一个map，用于存放group实例与命名空间的对应关系
)

// NewGroup Group构造函数
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	// 初始化group
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cacheInstance{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	// 添加到命名空间中
	groups[name] = g
	return g
}

// GetGroup 基于名称获取cache的实例
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// populateCache 缓存查询到的数据
func (g *Group) populateCache(key string, value *ByteView) {
	g.mainCache.Add(key, value)
}

// getLocally 从本地获取数据
func (g *Group) getLocally(key string) (*ByteView, error) {
	bytes, err := g.getter.Get(key) // 执行用户传递的回调函数
	if err != nil {
		return &ByteView{}, err

	}
	// 封装数据为ByteView类型
	value := &ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// getFromPeer 获取PeerGetter的实现体httpGetter，基于名称空间和key，获取其他节点的缓存信息
func (g *Group) getFromPeer(peer PeerGetter, key string) (*ByteView, error) {
	// 构建请求对象
	req := &ycachepb.Request{
		Group: g.name,
		Key:   key,
	}
	// 构建响应对象
	res := &ycachepb.Response{}
	// 请求其他节点的缓存数据
	err := peer.Get(req, res)
	if err != nil {
		return &ByteView{}, err
	}
	return &ByteView{b: res.Value}, nil


	//// 调用httpGetter的Get方法获取缓存记录
	//bytes, err := peer.Get(g.name, key)
	//// 如果HTTP请求失败了，那么返回空数据
	//if err != nil {
	//	return &ByteView{}, err
	//}
	//return &ByteView{b: bytes}, nil
}

// load 可以做一些数据组装操作
func (g *Group) load(key string) (value *ByteView, err error) {
	// g.peers != nil 表示需要从其他节点请求数据

	// 使用singleflight的Do方法包裹这段请求逻辑
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			// 基于key获取HTTP请求信息，这个peer就是httpGetter
			if peer, ok := g.peers.PickPeer(key); ok {
				// 将httpGetter实例作为参数传递到getFromPeer方法中，在getFromPeer获取其他节点的缓存记录
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				// 如果请求失败了，打印日志，会执行从本地获取数据的操作
				log.Println("[YCache] Failed to get from peer", err)
			}
		}
		// 从本地获取数据
		return g.getLocally(key)
	})
	if err == nil {
		// 由于返回的是接口，需要断言一下
		return view.(*ByteView), nil
	}

	return
}



// Get Group的get方法
func (g *Group) Get(key string) (*ByteView, error) {
	if key == "" {
		return &ByteView{}, fmt.Errorf("key is required")
	}
	// 如果缓存存在，直接返回
	if v, ok := g.mainCache.GetValue(key); ok {
		log.Println("[YCache] hit")
		return v, nil
	}
	// 如果缓存不存在，调用load方法
	return g.load(key)
}


// 新增方法

// RegisterPeers 将HTTPPool绑定到Group中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}