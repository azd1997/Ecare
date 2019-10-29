package main

import (
	"fmt"
	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	rootCmd.AddCommand(initChainCmd)

	initChainCmd.PersistentFlags().StringP("genesismsg", "m", "Eiger created the chain", "genesis msg")
}

var initChainCmd = &cobra.Command{
	Use:   "initchain",
	Short: "init chain in the local db. then you can use command startnode to continue the chain",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init chain starts ......")
		msg := fmt.Sprintf("%s | %s",
			time.Now().Format("2006/01/02 15:04:05"),
			cmd.PersistentFlags().Lookup("genesismsg").Value.String())
		_, err := ecoin.NewChain(E_Opts, msg)
		if err != nil {
			fmt.Println("init chain failed: ", err)
		} else {
			fmt.Println("init chain success!")
			fmt.Println("Tip: Now you can startnode to listen all requests in the ecare world!")
		}

	},
}