/**
 * @Author：Robby
 * @Date：2022/1/9 02:06
 * @Function：
 **/

package YCache

import "seven-days-projects/YCache/ycachepb"

// PeerPicker 用于实现基于用户传递的key，计算出请求的cache节点信息
type PeerPicker interface {
	// 基于key返回cache对应的HTTP客户端
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 用于实现HTTP客户端，从对应cache节点的名称空间中获取group，从而获取缓存记录
type PeerGetter interface {
	// 基于group、key的信息，实现HTTP的客户端，返回其他节点的缓存记录
	//Get(group string, key string) ([]byte, error)
	Get(in *ycachepb.Request, out *ycachepb.Response) error
}