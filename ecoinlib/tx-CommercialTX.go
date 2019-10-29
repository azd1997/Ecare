package ecoin


/*********************************************************************************************************************
                                                    CommercialTX接口
*********************************************************************************************************************/

// CommercialTX 商业性质交易，像R2P这样的交易属于商业性质，使用这个新的接口将它与其他类型TX区分开来
type CommercialTX interface {
	TX
	commercial() // 没有实际意义，只是为了让符合商业性质的交易实现它，从而区分开来。
	// 虽然现在商业交易只有R2P，但是为了之后的扩展性，还是设计了这个接口
}

// DeserializeCommercialTX 将字节切片反序列化为CommercialTX
func DeserializeCommercialTX(txBytes []byte) (tx CommercialTX, err error) {
	commercialTXTypes := []CommercialTX{
		&TxR2P{},
	} // 以后如果有新增的就从这加
	for _, tx = range commercialTXTypes {
		err = tx.Deserialize(txBytes)
		if err == nil {
			return
		}
	}
	return nil, ErrNotCommercialTxBytes
}
