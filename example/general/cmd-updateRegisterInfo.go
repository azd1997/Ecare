package main

import (
	"fmt"
	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(updateRegisterInfoCmd)
	updateRegisterInfoCmd.Flags().String("name", "", "register name")
	updateRegisterInfoCmd.Flags().String("phone", "", "register phone")
	updateRegisterInfoCmd.Flags().String("institution", "", "register institution")
}

var updateRegisterInfoCmd = &cobra.Command{
	Use:   "updateinfo",
	Short: "update register info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0  by%s\n", ecoin.NODE_VERSION, AUTHOR)

		// TODO： 实现思路：参数传入，更新opts并保存到config.json，得到新的opts后构建更改信息的交易，广播出去。
		//  （账户本身和注册信息没什么关联，所以注册信息想怎么改就怎么改，当然可以设置一个规则，或者说引用ecoin的程序制定了规则（一个判断函数），将这个规则传入）
		// TODO: 注意一旦本地有账户文件，config里的注册信息其实没什么用。只有没有注册文件时，才会根据config里的注册信息去创建账户
	},
}



