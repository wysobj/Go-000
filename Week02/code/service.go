package main

/*
题:
我们在数据库操作的时候，比如dao层中当遇到一个sql.ErrNoRows的时候，是否应该Wrap这个error，抛给上层。为什么，应该怎么做请写出代码？

答:
这种情况下不应该Wrap这个存储层返回的error并向上返回，因为这样就需要service层甚至更上层感知存储层的错误定义，形成了跨层的耦合。
*/

import (
	"fmt"
	"week2/dao"
)

func main() {
	var id = 1

	obj, err := dao.Get(id)
	if err != nil {
		fmt.Printf("Query dao failed, %+v", err)
		return
	}

	fmt.Printf("Get dao success, %v.", *obj)
}
