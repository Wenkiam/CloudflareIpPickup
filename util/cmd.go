package util

import (
	"fmt"
)

func ReadFromCmdLine(name string) (value string) {
	fmt.Println("请输入" + name)
	fmt.Scanln(&value)
	return
}
