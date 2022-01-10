/**
 * @Author：Robby
 * @Date：2022/1/9 11:40
 * @Function：
 **/

package main

import "fmt"

type Stu struct {
	Name string
}

func main() {
	fmt.Printf("%v\n", Stu{"Tom"}) // {Tom}
	fmt.Printf("%+v\n", Stu{"Tom"}) // {Name:Tom}
}