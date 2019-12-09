package transaction

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// ArbitrateArgs 新建交易函数newTxArbitrate()的传参
type ArbitrateArgs struct {
	ArbitratorAccount account.Account
	Arbitrator        account.UserId
	TargetTX          CommercialTX
	Arbitration       Arbitration
	Description       string
}

// Check 检查参数值是否合规
func (args *ArbitrateArgs) Check(argsFunc CheckArgsFunc) (err error) {

	// 检查Arbitrator
	arbitrator, err := args.ArbitratorAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if arbitrator != args.Arbitrator {
		return utils.WrapError("Args_Check", ErrUnmatchedSender)
	}

	// TargetTX 不能为空
	if args.TargetTX == nil {
		return utils.WrapError("Args_Check", ErrNilSourceTx)
	}

	// 仲裁结果码不检查了，在交易检查端检查


	// 根据传入的函数检查
	if err = argsFunc(args); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 参数有效
	return nil
}
