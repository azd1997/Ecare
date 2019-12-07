package tx

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// H2PArgs 新建交易函数newTxH2P()的传参
type H2PArgs struct {
	FromAccount        account.Account
	From      account.UserId
	P2H    *TxP2H
	Response    []byte
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *H2PArgs) Check() (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	if err = args.From.IsValid(account.Single, account.Hospital); err != nil {
		return utils.WrapError("Args_Check", err)
	}
	fromID, err := args.FromAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if args.From != fromID || args.From != args.P2H.To {
		return utils.WrapError("Args_Check", ErrUnmatchedTxReceiver)
	}

	// 检查p2h是否在未完成交易池中

	// 检查p2h内to是否和此时的from对应，都是本机拥有的账户

	// TODO: 检查 response 有效性

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}
