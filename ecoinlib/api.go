package ecoin

import (
	"encoding/base64"

	"github.com/azd1997/Ecare/ecoinlib/log"
)

// NewChain 外部包调用，一劳永逸
func NewChain(opts *Option, genesisMsg string) (c *Chain, err error) {

	// 1. 获取account
	account, err := loadOrCreateAccount(opts.Port())
	if err != nil {
		return nil, WrapError("NewChain", err)
	}
	// 2. 构造Option
	opts = opts.SetAccount(*account)
	log.Success("主配置构建成功！ 准备创建区块链......")

	// 3. 构造TxCoinbaseArgs
	coinbaseArgs := &TxCoinbaseArgs{
		To:          *opts.UserID(),
		Amount:      10000,
		Description: genesisMsg,
	}

	// 4. 构造InitChainArgs
	args := &InitChainArgs{
		coinbase: coinbaseArgs,
		opts:     opts,
	}

	// 5. InitChain
	c, err = InitChain(args)
	if err != nil {
		return nil, WrapError("NewChain", err)
	}
	log.Info("区块链创世信息: %s", genesisMsg)
	log.Success("区块链创建成功！ 接下来启动节点即代表整个区块链网络启动")
	return c, nil
}

// StartNode 启动节点
func StartNode(opts *Option) (err error) {
	// 1~8. 从本地恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}
	// 9. 启动节点服务器
	gsm.StartServer()
	log.Success("启动节点服务器成功！")

	return nil
}

// newTXAndBroadcast 新建交易并广播
func newTXAndBroadcast(typ uint8, gsm *GlobalStateMachine, args ArgsOfNewTX) error {
	tx, err := newTransactionWithArgsCheck(TX_GENERAL, gsm, args)
	if err != nil {
		return err
	}
	log.Info("[%s] 交易构建成功！ txID: %s", TXTypes[typ], base64.StdEncoding.EncodeToString(tx.Id()))

	// 广播交易
	err = gsm.BroadcastTx(gsm.Addrs.L1Ipv4Honest(), tx)
	if err != nil {
		return err
	}
	log.Success("广播交易成功! ")

	return nil
}

// 新建通用交易。并广播
func NewTxGeneral(opts *Option, to string, amount uint, description string) error {
	// TODO: 这里不应该这么做
	//  希望的是：
	//  两种使用情况：
	//  1.需要转账的节点平时不启动节点（反正自己的全局状态虽然落后，但是关于自己本人的信息是没有问题的），需要转账时恢复gsm再构造交易广播给所有转发节点
	//  2.也可以平时一直是启动服务器的状态，然后刚好某个时间想转账。这种情况下怎么做？
	//  2-1. 一样是需要转账时另起一个gsm进行广播而后再退出，关闭这个进程（可是可以）
	//  2-2. 有没有进程间共享一个gsm的办法呢
	//  现阶段： 先按照方案一来做，后面再优化。

	// 注意！：只是发送消息而不监听的话是不会占用端口的，而是随机分配一个

	// 1. 恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}

	// 2. 构造交易参数
	args := &TxGeneralArgs{
		From:        gsm.Opts().Account(),
		FromID:      *gsm.Opts().UserID(),
		To:          *newUserID(to, gsm),
		Amount:      Coin(amount),
		Description: description,
	}
	log.Info("交易参数构建成功！ to: %s; amount: %d; description: %s", to, amount, description)

	// 3. 构建交易并广播
	err = newTXAndBroadcast(TX_GENERAL, gsm, args)
	if err != nil {
		return err
	}

	return nil
}

// 新建R2P交易。并广播
func NewTxR2P(opts *Option, to string, amount uint, description string,
	targetRaw []uint, p2rID64 string, complete bool, ) error {
	// TODO: p2rID64指的是txID进行base64编码得到的字符串,所以要进行解码。
	// 由于在UCTXP作为键时使用string([]byte)相对来说很方便，所以策略是，内部使用string([]byte)作键，
	// 到api.go这个最接近外部调用者的文件中，就对这个ID进行base64编码及解码
	// TODO: 因为假如说普通用户刚上线加载UCTXP，可能是落后的，这时用户看不到最新的回复交易，也就没办法构建自己的回复，所以只能等区块链同步到了。这个逻辑是合理的

	// 1. 恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}

	// 2. 构造交易参数
	// 来源交易可以从区块中获取也可以从UCTXP中获取。从UCTXP中更快！ 对于普通节点而言，也维护UCTXP（但只存自己相关的，首先是从区块链中得知，再存到UCTXP）
	p2rID, err := base64.StdEncoding.DecodeString(p2rID64)
	if err != nil {
		return err
	}
	sourceTx := gsm.UCTXP.GetTX(p2rID)
	if sourceTx == nil {	// tx == nil 说明不存在
		return ErrTXNotInUCTXP
	}
	p2rBytes, err := sourceTx.Serialize()
	if err != nil {
		return err
	}
	// TODO: 犹豫中。。。是否要修改TxR2P这些交易里的源交易字节这一项？改成源交易指针？或者源交易ID？
	args := &TxR2PArgs{
		From:        gsm.Opts().Account(),
		FromID:      *gsm.Opts().UserID(),
		To:          *newUserID(to, gsm),
		Amount:      Coin(amount),
		Description: description,
		PurchaseTarget: *newTargetData(targetRaw),
		P2RBytes:       p2rBytes,
		TxComplete:     complete,
		Storage:        &MosquittoBroker{Addr:opts.brokerAddr},
	}
	log.Info("交易参数构建成功！ to: %s; amount: %d; description: %s", to, amount, description)

	// 3. 构建交易并广播
	err = newTXAndBroadcast(TX_R2P, gsm, args)
	if err != nil {
		return err
	}

	return nil
}

// 新建p2r交易。并广播
func NewTxP2R(opts *Option, r2pID64 string, response []byte, description string) error {

	// 1. 恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}

	// 2. 构造交易参数
	// 来源交易可以从区块中获取也可以从UCTXP中获取。从UCTXP中更快！ 对于普通节点而言，也维护UCTXP（但只存自己相关的，首先是从区块链中得知，再存到UCTXP）
	r2pID, err := base64.StdEncoding.DecodeString(r2pID64)
	if err != nil {
		return err
	}
	sourceTx := gsm.UCTXP.GetTX(r2pID)
	if sourceTx == nil {	// tx == nil 说明不存在
		return ErrTXNotInUCTXP
	}
	r2pBytes, err := sourceTx.Serialize()
	if err != nil {
		return err
	}
	// TODO: 犹豫中。。。是否要修改TxR2P这些交易里的源交易字节这一项？改成源交易指针？或者源交易ID？
	args := &TxP2RArgs{
		From:        gsm.Opts().Account(),
		FromID:      *gsm.Opts().UserID(),
		R2PBytes:r2pBytes,
		Response:response,
		Description: description,
	}
	log.Info("交易参数构建成功！")

	// 3. 构建交易并广播
	err = newTXAndBroadcast(TX_P2R, gsm, args)
	if err != nil {
		return err
	}

	return nil
}

// 新建p2h交易。并广播
func NewTxP2H(opts *Option, to string, amount uint, description string, target []uint, typ uint8) error {

	// 1. 恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}

	// 2. 构造交易参数
	args := &TxP2HArgs{
		From:        gsm.Opts().Account(),
		FromID:      *gsm.Opts().UserID(),
		To:*newUserID(to, gsm),
		Amount:Coin(amount),
		Description: description,
		PurchaseTarget:*newTargetData(target),
		PurchaseType:typ,
		Storage:&MosquittoBroker{Addr:opts.brokerAddr},
	}
	log.Info("交易参数构建成功！ to: %s; amount: %d; description: %s", to, amount, description)

	// 3. 构建交易并广播
	err = newTXAndBroadcast(TX_P2H, gsm, args)
	if err != nil {
		return err
	}

	return nil
}

// NewTxH2P 新建交易并广播
func NewTxH2P(opts *Option, p2hID64 string, response []byte, description string) error {

	// 1. 恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}

	// 2. 构造交易参数
	// 来源交易可以从区块中获取也可以从UCTXP中获取。从UCTXP中更快！ 对于普通节点而言，也维护UCTXP（但只存自己相关的，首先是从区块链中得知，再存到UCTXP）
	p2hID, err := base64.StdEncoding.DecodeString(p2hID64)
	if err != nil {
		return err
	}
	sourceTx := gsm.UCTXP.GetTX(p2hID)
	if sourceTx == nil {	// tx == nil 说明不存在
		return ErrTXNotInUCTXP
	}
	p2hBytes, err := sourceTx.Serialize()
	if err != nil {
		return err
	}
	// TODO: 犹豫中。。。是否要修改TxR2P这些交易里的源交易字节这一项？改成源交易指针？或者源交易ID？
	args := &TxH2PArgs{
		From:        gsm.Opts().Account(),
		FromID:      *gsm.Opts().UserID(),
		P2HBytes:p2hBytes,
		Response:response,
		Description: description,
	}
	log.Info("交易参数构建成功！")

	// 3. 构建交易并广播
	err = newTXAndBroadcast(TX_H2P, gsm, args)
	if err != nil {
		return err
	}

	return nil
}

// 新建p2d交易。并广播
func NewTxP2D(opts *Option, to string, amount uint, description string, target []uint) error {

	// 1. 恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}

	// 2. 构造交易参数
	args := &TxP2HArgs{
		From:        gsm.Opts().Account(),
		FromID:      *gsm.Opts().UserID(),
		To:*newUserID(to, gsm),
		Amount:Coin(amount),
		Description: description,
		PurchaseTarget:*newTargetData(target),
		Storage:&MosquittoBroker{Addr:opts.brokerAddr},
	}
	log.Info("交易参数构建成功！ to: %s; amount: %d; description: %s", to, amount, description)

	// 3. 构建交易并广播
	err = newTXAndBroadcast(TX_P2D, gsm, args)
	if err != nil {
		return err
	}

	return nil
}

// NewTxD2P 新建交易并广播
func NewTxD2P(opts *Option, p2dID64 string, response []byte, description string) error {

	// 1. 恢复GSM
	gsm, err := newGSM(opts)
	if err != nil {
		return err
	}

	// 2. 构造交易参数
	// 来源交易可以从区块中获取也可以从UCTXP中获取。从UCTXP中更快！ 对于普通节点而言，也维护UCTXP（但只存自己相关的，首先是从区块链中得知，再存到UCTXP）
	p2dID, err := base64.StdEncoding.DecodeString(p2dID64)
	if err != nil {
		return err
	}
	sourceTx := gsm.UCTXP.GetTX(p2dID)
	if sourceTx == nil {	// tx == nil 说明不存在
		return ErrTXNotInUCTXP
	}
	p2dBytes, err := sourceTx.Serialize()
	if err != nil {
		return err
	}
	// TODO: 犹豫中。。。是否要修改TxR2P这些交易里的源交易字节这一项？改成源交易指针？或者源交易ID？
	args := &TxD2PArgs{
		From:        gsm.Opts().Account(),
		FromID:      *gsm.Opts().UserID(),
		P2DBytes:p2dBytes,
		Response:response,
		Description: description,
	}
	log.Info("交易参数构建成功！")

	// 3. 构建交易并广播
	err = newTXAndBroadcast(TX_D2P, gsm, args)
	if err != nil {
		return err
	}

	return nil
}