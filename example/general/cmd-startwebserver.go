package main

import (
	"fmt"

	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startWebServerCmd)
}

var startWebServerCmd = &cobra.Command{
	Use:   "startserver",
	Short: "start the web server for visualized interface",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0  by%s\n", ecoin.NODE_VERSION, AUTHOR)
		fmt.Println("web server starting......")

		// TODO： webserver的逻辑：1. 获取gsm， 启动节点； 2. 开辟协程作为web服务器监听，并将gsm作为参数传入；
		// TODO: 启动webserver； 前端读取某个位置的账户文件，将账户数据传入后台，后台据此新建gsm启动节点

		// 使用该命令后启动web监听，web页面点击启动节点再启动node

		fmt.Printf("web server start success! access http://%s for visualized interface\n", serverAddr)
	},
}
