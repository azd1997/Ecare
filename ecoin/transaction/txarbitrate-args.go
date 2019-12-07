package tx

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// ArbitrateArgs 新建交易函数newTxArbitrate()的传参
type ArbitrateArgs struct {
	//	BaseArgs
	ArbitratorAccount       account.Account
	Arbitrator     account.UserId
	TargetTX    CommercialTX
	Description      string
}

// CheckArgsValue 检查参数值是否合规
func (args *ArbitrateArgs) Check() (err error) {

	// 检查Arbitrator
	arbitrator, err := args.ArbitratorAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if arbitrator != args.Arbitrator {
		return utils.WrapError("Args_Check", ErrUnmatchedTxReceiver)
	}

	// TargetTX 不能为空
	if args.TargetTX == nil {
		return utils.WrapError("Args_Check", err)
	}

	// 参数有效
	return nil
}
