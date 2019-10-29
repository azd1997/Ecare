package ecoin

/*********************************************************************************************************************
                                                    UBTXP相关
*********************************************************************************************************************/

// UnBlockedTXPool 未打包交易池。检验合格的交易才会存入。但是由于可能存在存入时余额等符合条件，但是出块时（可能不是他第一个出块）又不符合条件
// 所以需要在出块前再检查一次
// 遇上这种情况原本的转账者如何得知？只要原本的转账者余额没发生变化，交易就仍有机会打包，。如果余额变化了，认为交易失效
type UnBlockedTXPool struct {
	list []TX
}

// Add 添加新的待出块交易
func (ubtxp *UnBlockedTXPool) Add(tx TX) {
	ubtxp.list = append(ubtxp.list, tx)
}

// Del 移除交易
// 出块时可能会有池中交易出现失效的情况，此时直接当出完块应该将池置空
func (ubtxp *UnBlockedTXPool) Clear() {
	ubtxp.list = []TX{}
}

// All 取出所有交易
func (ubtxp *UnBlockedTXPool) All() []TX {
	return ubtxp.list
}

// Get 根据交易哈希取交易。若返回值为nil，说明要查找的交易不存在
func (ubtxp *UnBlockedTXPool) Get(txId Hash) TX {
	for _, tx := range ubtxp.list {
		if string(tx.Id()) == string(txId) {
			return tx
		}
	}
	return nil
}


/*********************************************************************************************************************
                                                    TBTXP相关
*********************************************************************************************************************/

// UnBlockedTXPool 未打包交易池。检验合格的交易才会存入。但是由于可能存在存入时余额等符合条件，但是出块时（可能不是他第一个出块）又不符合条件
// 所以需要在出块前再检查一次
// 遇上这种情况原本的转账者如何得知？只要原本的转账者余额没发生变化，交易就仍有机会打包，。如果余额变化了，认为交易失效
type ToBlockedTXPool struct {
	list []TX
}

// Add 添加新的待出块交易
func (tbtxp *ToBlockedTXPool) Add(tx TX) {
	tbtxp.list = append(tbtxp.list, tx)
}

// Del 移除交易
// 出块时可能会有池中交易出现失效的情况，此时直接当出完块应该将池置空
func (tbtxp *ToBlockedTXPool) Clear() {
	tbtxp.list = []TX{}
}

// All 取出所有交易
func (tbtxp *ToBlockedTXPool) All() []TX {
	return tbtxp.list
}

// Get 根据交易哈希取交易。若返回值为nil，说明要查找的交易不存在
func (tbtxp *ToBlockedTXPool) Get(txId Hash) TX {
	for _, tx := range tbtxp.list {
		if string(tx.Id()) == string(txId) {
			return tx
		}
	}
	return nil
}