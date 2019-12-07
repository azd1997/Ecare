package tx

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/common"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// CoinbaseArgs 新建交易函数newTxCoinbase()的传参
type CoinbaseArgs struct {
	To          account.UserId
	Amount      common.Coin
	Description string
}

// Check 检查参数值是否合规
func (args *CoinbaseArgs) Check() (err error) {
	// 检查 to 的有效性
	// coinbase交易只允许出块节点构建，而出块节点的roleNo 0~9
	if err = args.To.IsValid(account.A, account.AllRole); err != nil {
		return utils.WrapError("Args_Check", err)
	}


	// 检查 amount 有效性
	// TODO: 检查coinbase奖励是否合乎规则

	// TODO: 检查 description 格式，以及代码注入？

	return nil
}
