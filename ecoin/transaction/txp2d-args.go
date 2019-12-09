package transaction

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/storage"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// P2DArgs 新建交易函数newTxP2D()的传参
type P2DArgs struct {
	//	BaseArgs
	FromAccount    account.Account
	From           account.UserId
	To             account.UserId
	Amount         common.Coin
	Description    string
	PurchaseTarget storage.TargetData
}

// Check 检查参数值是否合规
func (args *P2DArgs) Check(argsFunc CheckArgsFunc) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.FromAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if args.From != fromID {
		return utils.WrapError("Args_Check", ErrWrongArgs)
	}

	// 检查 to 的有效性
	if err = args.To.IsValid(account.Single, account.Doctor); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查数据存储是否有效？需要像数据服务器发送查询请求
	// 非医院节点只能查询存在与否
	// 医院节点拥有数据读取权利，但只有数据所有者才能解密数据

	// 检查 purchaseTarget是否有效？
	if err = args.PurchaseTarget.IsOk(); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 根据传入的函数检查
	if err = argsFunc(args); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// 参数有效
	return nil
}
