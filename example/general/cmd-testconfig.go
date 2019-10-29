package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "display the config of client",
	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Printf("ipv4: %s; port: %d\n", E_config.Ipv4, E_config.Port)
		fmt.Printf("ipv4: %s; port: %d\n", E_Opts.Ipv4(), E_Opts.Port())
	},
}
