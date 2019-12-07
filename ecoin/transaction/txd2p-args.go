package tx

import (
	"github.com/azd1997/Ecare/ecoin/account"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// D2PArgs 新建交易函数newTxD2P()的传参
type D2PArgs struct {
	FromAccount        account.Account
	From      account.UserId
	P2D         *TxP2D
	Response    []byte
	Description string
}

// Check 检查参数值是否合规
func (args *D2PArgs) Check() (err error) {

	// 检查From
	if err = args.From.IsValid(account.Single, account.Patient); err != nil {
		return utils.WrapError("Args_Check", err)
	}

	// TODO： 检查p2d: 由于是已经被检查过的交易，只要查找未完成交易池中是否有就行了，没有就是无效的。
	// 交易验证时这部分工作由上层去做，参数校验时只调用也应该在上层维护一个未完成交易池（只与自己相关），
	// 然后在里边找这个P2D然后传入。因此这里也没什么必要检查P2D的存在性
	// 但根据P2D内容可以检查交易双方的对应关系
	// P2D不能为空
	if args.P2D == nil {
		return utils.WrapError("Args_Check", err)
	}

	// 检查p2D内to是否和此时的from对应，都是本机拥有的账户
	selfId, err := args.FromAccount.UserId()
	if err != nil {
		return utils.WrapError("Args_Check", err)
	}
	if selfId != args.P2D.To || selfId != args.From {
		return utils.WrapError("Args_Check", ErrUnmatchedSender)
	}


	// 参数有效
	return nil
}
