package tx

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
func (args *GeneralArgs) Check() (err error) {
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
		return utils.WrapError("Args_Check", ErrWrongArguments)
	}

	// 检查 to 的有效性
	if err = args.To.IsValid(account.All, account.AllRole); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查 amount 有效性(余额是否足够)
	// 交给tx包调用者去做

	// TODO: 检查 description 格式

	return nil
}
