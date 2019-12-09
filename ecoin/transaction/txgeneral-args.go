package transaction

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// GeneralArgs 新建交易函数newTxGeneral()的传参
// 不用指针是为了避免误操作将账户信息进行修改，尽管是可以恢复的
type GeneralArgs struct {
	FromAccount account.Account
	From        account.UserId
	To          account.UserId
	Amount      common.Coin
	Description string
}

// Check 检查参数值是否合规
func (args *GeneralArgs) Check(argsFunc CheckArgsFunc) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromId
	err = args.From.IsValid(account.All, account.AllRole)
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	fromId, err := args.FromAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if args.From != fromId {
		return utils.WrapError("Args_Check", ErrWrongArgs)
	}

	// 检查 to 的有效性
	if err = args.To.IsValid(account.All, 0); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 根据传入的函数检查
	if err = argsFunc(args); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	return nil
}
