package tx

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// TxP2RArgs 新建交易函数newTxP2R()的传参
type P2RArgs struct {
	//	BaseArgs
	FromAccount account.Account
	From        account.UserId
	R2P         *TxR2P
	Response    []byte
	Description string
}

// Check 检查参数值是否合规
func (args *P2RArgs) Check() (err error) {

	// 检查FromID
	fromID, err := args.FromAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if args.From != fromID {
		return utils.WrapError("Args_Check", ErrWrongArgs)
	}
	if err = args.From.IsValid(account.Single, account.Patient); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// R2P不能为空
	if args.R2P == nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查r2p内to是否和此时的from对应，都是本机拥有的账户
	if fromID != args.R2P.To {
		return utils.WrapError("Args_Check", ErrUnmatchedSender)

	}

	// 参数有效
	return nil
}
