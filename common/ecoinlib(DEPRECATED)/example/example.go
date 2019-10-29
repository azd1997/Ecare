package main

import (
	"fmt"
	ecoin "github.com/azd1997/Ecare/common/ecoinlib/api"
)

func main() {

	// 配置设置
	ecoin.Ecoin.DefaultOption().SetUserID("eiger")
	fmt.Println(ecoin.Ecoin.Opts.UserID(), ecoin.Ecoin.Opts.AccountFilePathTemp())

	// 测试Ecoin方法
	ecoin.Ecoin.TestMethod()
}