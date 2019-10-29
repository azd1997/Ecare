package types

// 实际使用的Role需在区块链各个版本的客户端软件配置文件中指定。这里只定义基本的格式

// 关于币，基本考虑与人民币等值，需用人民币换购并在系统内流通
// 增值，可以参考银行存款，将增长部分发给积极地用户
// 服务费，暂不设置

// 假设系统启动之初，作价一千万，

// Role 账户角色，作为权限控制
type Role struct {
	no      uint8   // 编号，从0开始。role0为创始者，编号不可改，别名可以自定义
	alias   string  // 名称
	initial Coin // 初始币量
	coinbaseReward Coin // 挖矿奖励交易。B类角色只能是0因为不能挖矿。A类可根据协议设置数值
	genesisReward Coin		// 创始者奖励量
	// ks和es组合可以描述很多种币的增长策略，默认值为ks=不设，es=不设，币量不自增
	enableKsEs bool
	ks []int // 系数值		-x^3+3x^2+x+1 + x^-1   [-1 3 1 1 1]
	es []int // 幂指数值 [3 2 1 0 -1]
}

// No 获取角色编号
func (r *Role) No() uint8 {
	return r.no
}

// Alias 获取角色的字符名称
func (r *Role) Alias() string {
	return r.alias
}

// InitialBalance 初始余额
func (r *Role) InitialBalance() Coin {
	return r.initial
}

// CoinbaseReward 出块奖励金额
func (r *Role) CoinbaseReward() Coin {
	return r.coinbaseReward
}

// GenesisReward 出块奖励金额
func (r *Role) GenesisReward() Coin {
	return r.genesisReward
}

// TODO: 转发节点出块有一定收益； 转发节点持续运行自增长（相当于利润）； 普通用户节点持续使用自增长（利润、bonus）
