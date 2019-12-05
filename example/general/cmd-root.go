package main

import "github.com/spf13/cobra"

var (
	cfgFile string
)

const (
	AUTHOR = "Eiger"
)

var rootCmd = &cobra.Command{
	Use:                        "ecoin",
	Short:                      "An example ecoin client",
	Run: func(cmd *cobra.Command, args []string) {
		// Do something
	},
}

func init() {
	// 初始化配置加载。一定要在config pflag绑定之后，否则取不到值。可以使用
	cobra.OnInitialize(initConfig)
	// 来保证其在flag初始化之后

	// config pflag绑定
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config.json", "config file path")
}

func execute() error {
	return rootCmd.Execute()
}