package tx

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/storage"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// R2PArgs 新建交易函数newTxR2P()的传参
type R2PArgs struct {
	//	BaseArgs
	FromAccount    *account.Account
	From           account.UserId
	To             account.UserId
	Amount         common.Coin
	P2R            *TxP2R
	PurchaseTarget storage.TargetData
	Response       []byte
	TxComplete     bool
	Description    string
}

// Check 检查参数值是否合规
func (args *R2PArgs) Check() (err error) {

	// 检查FromID
	fromID, err := args.FromAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if args.From != fromID {
		return utils.WrapError("Args_Check", ErrWrongArguments)
	}
	if err = args.From.IsValid(account.Single, account.ResearchInstitution); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查 to 的有效性
	if err = args.To.IsValid(account.Single, account.Patient); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查to与P2R的from是否匹配
	if args.P2R != nil {
		if args.To != args.P2R.From {
			return utils.WrapError("Args_Check", ErrUnmatchedTxReceiver)
		}
		if args.From != args.P2R.R2P.From {
			return utils.WrapError("Args_Check", ErrUnmatchedTxSender)
		}
	}

	// 检查 purchaseTarget是否有效？
	if err = args.PurchaseTarget.IsOk(); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 参数有效
	return nil
}
