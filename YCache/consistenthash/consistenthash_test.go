/**
 * @Author：Robby
 * @Date：2022/1/8 20:13
 * @Function：
 **/

package consistenthash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	// 创建环形环，由于在测试用例中，无法知道字符串对应的真实hash，所有自定义一个hash函数
	hash := NewMap(3, func(key []byte) uint32 {
		// 简单的将真实cache节点的IP转换为int类型，作为hash函数
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 在hash环上添加真实cache节点，那么虚拟节点的hash为： 02/12/22、04/14/24、06/16/26 这9个数，这9个数排序后为：02、04、06、12、14、16、22、24、26
	hash.Add("6", "4", "2")

	// 如果查询key是 2、11、23、27，那么对于的虚拟cache的hash值分别是02、12、24、02，那么真实cache节点的信息为：2、2、4、2
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	// 运行hash一致性算法，比较结果
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// 对应的虚拟节点hash是 08, 18, 28
	hash.Add("8")

	// 那么27对应虚拟节点是28，也就是8
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// 搞个错误的试试，4对应的虚拟hash是6、那么故意写成2
	testCases["4"] = "2"
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

}