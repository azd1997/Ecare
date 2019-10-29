package main

import (
	"fmt"

	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version of client",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Ecare v%d.0.0  by%s\n", ecoin.NODE_VERSION, AUTHOR)
	},
}
