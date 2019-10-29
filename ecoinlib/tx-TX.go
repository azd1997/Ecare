package ecoin

/*********************************************************************************************************************
                                                    TX接口
*********************************************************************************************************************/

// TX 标志一笔交易，接口
type TX interface {
	String() string
	Serialize() (result []byte, err error)
	Deserialize(data []byte) (err error)
	Hash() (id Hash, err error)
	IsValid(gsm *GlobalStateMachine) (err error)
	TypeNo() uint
	Id() Hash
}

// NewTransactionWithArgsCheck 新建一个交易，传入交易类型与其他参数，构建具体的交易。
func newTransactionWithArgsCheck(typ uint, gsm *GlobalStateMachine, args ArgsOfNewTX) (TX, error) {
	switch typ {
	case TX_COINBASE:
		// 1. 检查参数
		args, ok := args.(*TxCoinbaseArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxCoinbase(args) // *TxCoinbase 实现了 TX 接口， 粗略的可以认为一个×TxCoinbase是一个TX
	case TX_GENERAL:
		// 1. 检查参数
		args, ok := args.(*TxGeneralArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxGeneral(args)
	case TX_R2P:
		// 1. 检查参数
		args, ok := args.(*TxR2PArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxR2P(args)
	case TX_P2R:
		// 1. 检查参数
		args, ok := args.(*TxP2RArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxP2R(args)
	case TX_P2H:
		// 1. 检查参数
		args, ok := args.(*TxP2HArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxP2H(args)
	case TX_H2P:
		// 1. 检查参数
		args, ok := args.(*TxH2PArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxH2P(args)
	case TX_P2D:
		// 1. 检查参数
		args, ok := args.(*TxP2DArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxP2D(args)
	case TX_D2P:
		// 1. 检查参数
		args, ok := args.(*TxD2PArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxD2P(args)
	case TX_ARBITRATE:
		// 1. 检查参数
		args, ok := args.(*TxArbitrateArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		if err := args.CheckArgsValue(gsm); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxArbitrate(args)
	default:
		return nil, ErrUnknownTransactionType
	}
}

// NewTransaction 新建一个交易，传入交易类型与其他参数，构建具体的交易。 一定要严格检查输入参数顺序和类型！！！
func newTransaction(typ uint, args ArgsOfNewTX) (TX, error) {
	switch typ {
	case TX_COINBASE:
		// 1. 检查参数
		args, ok := args.(*TxCoinbaseArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxCoinbase(args) // *TxCoinbase 实现了 TX 接口， 粗略的可以认为一个×TxCoinbase是一个TX
	case TX_GENERAL:
		// 1. 检查参数
		args, ok := args.(*TxGeneralArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxGeneral(args)
	case TX_R2P:
		// 1. 检查参数
		args, ok := args.(*TxR2PArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxR2P(args)
	case TX_P2R:
		// 1. 检查参数
		args, ok := args.(*TxP2RArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxP2R(args)
	case TX_P2H:
		// 1. 检查参数
		args, ok := args.(*TxP2HArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxP2H(args)
	case TX_H2P:
		// 1. 检查参数
		args, ok := args.(*TxH2PArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxH2P(args)
	case TX_P2D:
		// 1. 检查参数
		args, ok := args.(*TxP2DArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxP2D(args)
	case TX_D2P:
		// 1. 检查参数
		args, ok := args.(*TxD2PArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxD2P(args)
	case TX_ARBITRATE:
		// 1. 检查参数
		args, ok := args.(*TxArbitrateArgs)
		if !ok {
			return nil, ErrWrongArgsForNewTX
		}
		// 2. 新建交易
		return newTxArbitrate(args)
	default:
		return nil, ErrUnknownTransactionType
	}
}

// DeserializeTX 根据指定具体交易类型编号进行反序列化
func DeserializeTX(typ uint, txBytes []byte) (tx TX, err error) {
	switch typ {
	case TX_COINBASE:
		tx = &TxCoinbase{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_GENERAL:
		tx = &TxGeneral{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_R2P:
		tx = &TxR2P{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_P2R:
		tx = &TxP2R{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_P2H:
		tx = &TxP2H{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_H2P:
		tx = &TxH2P{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_P2D:
		tx = &TxP2D{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_D2P:
		tx = &TxD2P{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_ARBITRATE:
		tx = &TxArbitrate{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_AUTO:
		// 调用者不知道具体是哪种交易，则typ输TX_AUTO(0)，将自动适用所有类型去测试。
		txTypes := []TX{
			&TxCoinbase{},
			&TxGeneral{},
			&TxR2P{},
			&TxP2R{},
			&TxP2H{},
			&TxH2P{},
			&TxP2D{},
			&TxD2P{},
			&TxArbitrate{},
		}
		for _, tx = range txTypes {
			err = tx.Deserialize(txBytes)
			if err == nil {
				return tx, nil
			}
		}
		return nil, WrapError("DeserializeTX", ErrNotTxBytes)
	default:
		return nil, WrapError("DeserializeTX", ErrUnknownTransactionType)
	}
}

// BaseTransaction 基交易，包含所有具体交易类型包含的共同属性。
type BaseTransaction struct {
	ID          Hash          `json:"id"`
	Time        UnixTimeStamp `json:"time"`
	To          UserID        `json:"to"`
	Amount      Coin          `json:"amount"`
	Description string        `json:"description"`
}

