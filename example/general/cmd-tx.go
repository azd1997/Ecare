package main

import (
	"fmt"
	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

func init() {
	rootCmd.AddCommand(txCmd)
	txCmd.Flags().StringP("to", "t", "", "id of the one you transfer to")
	txCmd.Flags().UintP("amount", "a", 0, "transfer amount")
	txCmd.PersistentFlags().StringP("description", "d", "", "additional description")

	txCmd.AddCommand(txGeneralCmd)
	txGeneralCmd.Flags().StringP("to", "t", "", "id of the one you transfer to")
	txGeneralCmd.Flags().UintP("amount", "a", 0, "transfer amount")

	txCmd.AddCommand(txR2PCmd)
	txR2PCmd.Flags().StringP("to", "t", "", "id of the one you transfer to")
	txR2PCmd.Flags().UintP("amount", "a", 0, "transfer amount")
	txR2PCmd.Flags().String("p2r", "", "base64 string of p2r tx id")
	txR2PCmd.Flags().Bool("complete", false, "is whole tx complete?")
	txR2PCmd.Flags().UintSlice("target", []uint{0, 0, 0},
	"target data: arg0 for start time, arg1 for end time, arg2 for the last n records")

	txCmd.AddCommand(txP2RCmd)
	txP2RCmd.Flags().String("r2p", "", "base64 string of r2p tx id")
	// TODO： 尽管cobra里有BytesBase64等flag绑定方法。但是不是太熟悉，暂时不使用，先使用string然后自己转换
	txP2RCmd.Flags().String("response", "this is a response", "response of its source tx")

	txCmd.AddCommand(txP2HCmd)
	txP2HCmd.Flags().StringP("to", "t", "", "id of the one you transfer to")
	txP2HCmd.Flags().UintP("amount", "a", 0, "transfer amount")
	txP2HCmd.Flags().UintSlice("target", []uint{0, 0, 0},
		"target data: arg0 for start time, arg1 for end time, arg2 for the last n records")
	txP2HCmd.Flags().Uint8("type", 0, "type of diagnose")

	txCmd.AddCommand(txH2PCmd)
	txH2PCmd.Flags().String("p2h", "", "base64 string of p2h tx id")
	txH2PCmd.Flags().String("response", "this is a response", "response of its source tx")

	txCmd.AddCommand(txP2DCmd)
	txP2DCmd.Flags().StringP("to", "t", "", "id of the one you transfer to")
	txP2DCmd.Flags().UintP("amount", "a", 0, "transfer amount")
	txP2DCmd.Flags().UintSlice("target", []uint{0, 0, 0},
		"target data: arg0 for start time, arg1 for end time, arg2 for the last n records")

	txCmd.AddCommand(txD2PCmd)
	txD2PCmd.Flags().String("p2d", "", "base64 string of p2d tx id")
	txD2PCmd.Flags().String("response", "this is a response", "response of its source tx")


}

var txCmd = &cobra.Command{
	Use:   "tx",
	Short: "create a tx",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("default tx is txGeneral......\n")

		// 检查命令行参数是否已设置
		to := cmd.PersistentFlags().Lookup("to").Value.String()
		//fmt.Printf("tx-to: %s\n", to)
		amount, err := strconv.Atoi(cmd.PersistentFlags().Lookup("amount").Value.String())
		if err != nil {
			fmt.Println("amount: ", err)
		}
		//fmt.Printf("tx-amount: %d\n", amount)
		description := cmd.PersistentFlags().Lookup("description").Value.String()
		//fmt.Printf("tx-description: %s\n", description)

		if to == "" {
			cmd.Help()
			os.Exit(1)
		}

		// 构造交易并广播
		err = ecoin.NewTxGeneral(E_Opts, to, uint(amount), description)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}

var txGeneralCmd = &cobra.Command{
	Use:   "general",
	Short: "create a txGeneral",
	Run: func(cmd *cobra.Command, args []string) {
		// 解析参数
		to := cmd.Flag("to").Value.String()
		//fmt.Println("to: ", to)
		amount, err := cmd.Flags().GetUint("amount")
		if err != nil {
			fmt.Println("amount: ", err)
		}
		//fmt.Println("amount: ", amount)
		description := cmd.Flag("description").Value.String()

		if to == "" {
			cmd.Help()
			os.Exit(1)
		}

		// 构造交易并广播
		err = ecoin.NewTxGeneral(E_Opts, to, amount, description)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}

var txR2PCmd = &cobra.Command{
	Use:   "r2p",
	Short: "create a txR2P",
	Run: func(cmd *cobra.Command, args []string) {

		// ./main tx r2p -t=azd -a=70 --target 1,2,3 --p2r mocktx

		// 解析参数
		//这样会panic: to := cmd.PersistentFlags().Lookup("to").Value.String()
		// 子命令使用父亲、爷爷传下来的全局flag时是Flag使用，而不是PersistentFlag(只有绑定者可用)
		to := cmd.Flag("to").Value.String()
		//fmt.Println("to: ", to)
		amount, err := cmd.Flags().GetUint("amount")
		if err != nil {
			fmt.Println("amount: ", err)
		}
		//fmt.Println("amount: ", amount)
		description := cmd.Flag("description").Value.String()
		target, err := cmd.Flags().GetUintSlice("target")
		if err != nil {
			fmt.Println("target: ", err)
		}
		p2rID64 := cmd.Flag("p2r").Value.String()
		complete, err := cmd.Flags().GetBool("complete")
		if err != nil {
			fmt.Println("complete: ", err)
		}

		if to == "" || len(target) == 0 || len(target) > 3 {
			cmd.Help()
			os.Exit(1)
		}

		// 构造交易并广播
		err = ecoin.NewTxR2P(E_Opts, to, amount, description, target, p2rID64, complete)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}

// TODO: 其实也可以选择将这些交易参数预置在Option中，由调用时更新opts里的交易参数，再去使用。这样好处是写函数传参时简洁许多

var txP2RCmd = &cobra.Command{
	Use:   "p2r",
	Short: "create a txP2R",
	Run: func(cmd *cobra.Command, args []string) {
		// 解析参数
		description := cmd.Flag("description").Value.String()
		r2pID64 := cmd.Flag("r2p").Value.String()
		response := cmd.Flag("response").Value.String()

		// 构造交易并广播
		err := ecoin.NewTxP2R(E_Opts, r2pID64, []byte(response), description)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}

var txP2HCmd = &cobra.Command{
	Use:   "p2h",
	Short: "create a txP2H",
	Run: func(cmd *cobra.Command, args []string) {

		to := cmd.Flag("to").Value.String()
		amount, err := cmd.Flags().GetUint("amount")
		if err != nil {
			fmt.Println("amount: ", err)
		}
		typ, err := cmd.Flags().GetUint8("type")
		if err != nil {
			fmt.Println("type: ", err)
		}
		description := cmd.Flag("description").Value.String()
		target, err := cmd.Flags().GetUintSlice("target")
		if err != nil {
			fmt.Println("target: ", err)
		}
		// typ 目前只有0,1
		if to == "" || len(target) == 0 || len(target) > 3 || typ > 1 {
			cmd.Help()
			os.Exit(1)
		}

		// 构造交易并广播
		err = ecoin.NewTxP2H(E_Opts, to, amount, description, target, typ)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}

var txH2PCmd = &cobra.Command{
	Use:   "h2p",
	Short: "create a txH2P",
	Run: func(cmd *cobra.Command, args []string) {
		// 解析参数
		description := cmd.Flag("description").Value.String()
		p2hID64 := cmd.Flag("p2h").Value.String()
		response := cmd.Flag("response").Value.String()

		// 构造交易并广播
		err := ecoin.NewTxH2P(E_Opts, p2hID64, []byte(response), description)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}

var txP2DCmd = &cobra.Command{
	Use:   "p2d",
	Short: "create a txP2D",
	Run: func(cmd *cobra.Command, args []string) {

		to := cmd.Flag("to").Value.String()
		amount, err := cmd.Flags().GetUint("amount")
		if err != nil {
			fmt.Println("amount: ", err)
		}
		description := cmd.Flag("description").Value.String()
		target, err := cmd.Flags().GetUintSlice("target")
		if err != nil {
			fmt.Println("target: ", err)
		}
		// typ 目前只有0,1
		if to == "" || len(target) == 0 || len(target) > 3 {
			cmd.Help()
			os.Exit(1)
		}

		// 构造交易并广播
		err = ecoin.NewTxP2D(E_Opts, to, amount, description, target)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}

var txD2PCmd = &cobra.Command{
	Use:   "d2p",
	Short: "create a txD2P",
	Run: func(cmd *cobra.Command, args []string) {
		// 解析参数
		description := cmd.Flag("description").Value.String()
		p2dID64 := cmd.Flag("p2d").Value.String()
		response := cmd.Flag("response").Value.String()

		// 构造交易并广播
		err := ecoin.NewTxD2P(E_Opts, p2dID64, []byte(response), description)
		if err != nil {
			fmt.Println("交易事务构建失败: ", err)
		} else {
			fmt.Println("交易事务构建成功！")
		}
	},
}