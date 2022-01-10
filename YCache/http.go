/**
 * @Author：Robby
 * @Date：2022/1/9 02:06
 * @Function：
 **/

package YCache

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"seven-days-projects/YCache/consistenthash"
	"seven-days-projects/YCache/ycachepb"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_cache/"
	defaultReplicas = 50
)


// HTTPPool 存放当前节点运行的socket信息
type HTTPPool struct {
	self     string
	basePath string

	// 下面是新增成员变量，用于客户端实现
	mu sync.Mutex // 添加互斥锁
	peers *consistenthash.Map // 添加一致性hash算法实例
	httpGetters map[string]*httpGetter // 映射真实的cache实例信息与HTTP客户端的对应关系
}

// NewHTTPPool 构造函数
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 封装请求日志输出，当有请求进入到server，在ServeHTTP方法中会被调用
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP http的handler, 约定的访问路径为/<basepath>/<groupname>/<key>，实现Handler接口的ServeHTTP方法
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 验证路由前缀是否以/api开头/
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	// 打印日志，包括请求的方法和路径，例如 GET /api/scores/Tom
	p.Log("%s %s", r.Method, r.URL.Path)

	// 获取basePath的字符串长度，也就是字符数，/api/字符数是5
	basePathLengh := len(p.basePath)
	// 获取请求路径中去掉/api/的路径， 就是scores/Tom
	backendPath := r.URL.Path[basePathLengh:]
	// 以/为分隔符，切割两段
	parts := strings.SplitN(backendPath, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 获取命名空间
	groupName := parts[0]
	// 获取key名称
	key := parts[1]

	// 获取group对象
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	// cache中获取key
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//// 返回缓存数据，application/octet-stream是二级制流
	//w.Header().Set("Content-Type", "application/octet-stream")
	//w.Write(view.ByteSlice())
	// 通过 proto.Marshal函数对响应的数据进行序列化
	body, err := proto.Marshal(&ycachepb.Response{Value: view.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}


// 下面是HTTP客户端实现

// 表示HTTP的请求信息，例如：例如 http://locahost:8080/api/，
type httpGetter struct {
	baseURL string
}

// Get 实现PeerGetter接口的方法，构建HTTP客户端
// Get方法的实现也要改
func (h *httpGetter) Get(in *ycachepb.Request, out *ycachepb.Response) error {

	// 拼凑url：http://locahost:8080/api/scores/Tom
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.GetGroup()), // 将url的字符串转移，类似于url路径的编码
		url.QueryEscape(in.GetKey()),
	)
	// HTTP客户端请求cache的IP地址
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}
	// 获取请求数据，转换为[]byte类型，此时请求数据一定是记录的值
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	// 将bytes基于protobuf解码，解码的数据写入到out
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	//return bytes, nil
	return nil
}

// 验证httpGetter是否实现了PeerGetter接口
var _ PeerGetter = &httpGetter{}
// 下面这种验证方式，是将nil转换为httpGetter类型，再赋值给接口
//var _ PeerGetter = (*httpGetter)(nil)

// Set 实现了基于cache节点信息创建节点与HTTP请求信息的对应关系
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 获取一致性hash算法实例
	p.peers = consistenthash.NewMap(defaultReplicas, nil)
	// 添加cache节点信息
	p.peers.Add(peers...)
	// 初始化cache节点与URL的对应关系，例如 {"127.0.0.1": "httpClient1", "127.0.0.2": "httpClient2",}
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	// 遍历cache节点，创建cache节点与HTTP客户端映射关系，因为httpGetter实现了HTTP客户端+url
	for _, peer := range peers {
		// peer + p.basePath 为 127.0.0.1/api/
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer 实现PeerPicker接口PickPeer方法，根据具体的 key，选择cache节点，返回节点对应的 HTTP 客户端
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 基于p.peers一致性hash算法获取节点信息，如果节点不是自己，也不为空，那么获取到HTTP客户端
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Get data from %s", peer)
		// 基于cache节点信息获取到客户端
		return p.httpGetters[peer], true
	}

	// 如果是自己，那么返回nil
	return nil, false
}

// 验证HTTPPool实现了PeerPicker接口
var _ PeerPicker = &HTTPPool{}