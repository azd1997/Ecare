package main

import (
	"fmt"

	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(printCmd)

	printCmd.AddCommand(printChainCmd)

	printCmd.AddCommand(printBlockCmd)

	printCmd.AddCommand(printTxCmd)

	printCmd.AddCommand(printAccountCmd)
}

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "print information of chain/block/tx/account/userid/...",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0\n", ecoin.NODE_VERSION)
	},
}

var printChainCmd = &cobra.Command{
	Use:   "chain",
	Short: "print information of chain, default the latest 5 block",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0\n", ecoin.NODE_VERSION)
	},
}

var printBlockCmd = &cobra.Command{
	Use:   "block",
	Short: "print information of block, please specify block id or hash",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0\n", ecoin.NODE_VERSION)
	},
}

var printTxCmd = &cobra.Command{
	Use:   "tx",
	Short: "print information of tx",
	Long: "print information of tx, please specify tx id, optional block id or hash (the block which contains the tx), optional time range. for quick search",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0\n", ecoin.NODE_VERSION)
	},
}

var printAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "print information of account",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0\n", ecoin.NODE_VERSION)
	},
}

