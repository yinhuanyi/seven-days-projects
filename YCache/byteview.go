/**
 * @Author：Robby
 * @Date：2022/1/9 02:05
 * @Function：
 **/

package YCache

// ByteView 对记录的value，封装自定义数据类型
type ByteView struct {
	b []byte
}

func (v *ByteView) Len() int {
	return len(v.b)
}

func (v *ByteView) String() string {
	return string(v.b)
}

// 拷贝一份ByteView的b数组
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// ByteSlice 的b是只读的，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}