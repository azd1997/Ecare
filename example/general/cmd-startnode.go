package main

import (
	"fmt"
	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startnodeCmd)

	// config参数必须显式地初始化于initConfig之前。所以这句话应该挪至main.go中initConfig前。这导致rootCmd.AddCommand也得显式的挪过去
	// startnodeCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.json", "config file path")

	// 前面说的是无意义的。
	// cobra的PersistentFlag(持久flag)可以被子命令使用，子命令无需再添加
/*	(base) eiger@eiger-ThinkPad-X1-Carbon-3rd:~/gopath-default/src/github.com/azd1997/Ecare/example/general$ ./main --config ./cccc
	2019/10/16 03:22:49 cfgFile:  ./cccc
	2019/10/16 03:22:49 读取配置失败:  open ./cccc: no such file or directory
	(base) eiger@eiger-ThinkPad-X1-Carbon-3rd:~/gopath-default/src/github.com/azd1997/Ecare/example/general$ ./main startnode --config ./cccc
	2019/10/16 03:23:05 cfgFile:  ./cccc
	2019/10/16 03:23:05 读取配置失败:  open ./cccc: no such file or directory*/

}

var startnodeCmd = &cobra.Command{
	Use:   "startnode",
	Short: "start node server",
	Run: func(cmd *cobra.Command, args []string) {
		err := ecoin.StartNode(E_Opts)
		if err != nil {
			fmt.Println("start node server failed: ", err)
		} else {
			fmt.Println("start node server success!")
		}
	},
}
