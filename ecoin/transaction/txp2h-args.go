package tx

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/storage"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// P2HArgs 新建交易函数newTxP2H()的传参
type P2HArgs struct {
	//	BaseArgs
	FromAccount    account.Account
	From           account.UserId
	To             account.UserId
	Amount         common.Coin
	Description    string
	PurchaseTarget storage.TargetData
	PurchaseType   uint8
}

// CheckArgsValue 检查参数值是否合规
func (args *P2HArgs) Check() (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

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

	// 检查 to 的有效性
	if err = args.To.IsValid(account.Single, account.Hospital); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查 purchaseTarget是否有效？
	if err = args.PurchaseTarget.IsOk(); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查 purchaseType
	if args.PurchaseType != ECG_DIAG_AUTO && args.PurchaseType != ECG_DIAG_DOCTOR {
		return ErrUnknownPurchaseType
	}

	// 参数有效
	return nil
}
