package transaction

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// H2PArgs 新建交易函数newTxH2P()的传参
type H2PArgs struct {
	FromAccount account.Account
	From        account.UserId
	P2H         *TxP2H
	Response    []byte
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *H2PArgs) Check(argsFunc CheckArgsFunc) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	if err = args.From.IsValid(account.Single, account.Hospital); err != nil {
		return utils.WrapError("Args_Check", err)
	}
	fromID, err := args.FromAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// P2H不能为空
	if args.P2H == nil {
		return utils.WrapError("Args_Check", err)
	}

	//
	if args.From != fromID || args.From != args.P2H.To {
		return utils.WrapError("Args_Check", ErrUnmatchedSender)
	}

	// 根据传入的函数检查
	if err = argsFunc(args); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 参数有效
	return nil
}
