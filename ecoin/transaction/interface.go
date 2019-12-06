package tx

import (
	"github.com/azd1997/Ecare/ecoin/common"
)


// 由于想要实现的是，调用方不知道每一种交易内部实现，因此，调用方只需要使用接口类型和接口方法
// 而且由于transaction包调用太过频繁，是核心业务，所以尽量缩短命名
// 因此这里接口命名不使用Ixxx命名方式，直接使用大写缩写TX

// TX 标志一笔交易，接口
type TX interface {
	String() string
	Serialize() (result []byte, err error)
	Deserialize(data []byte) (err error)
	Hash() (id common.Hash, err error)
	IsValid() (err error)
	TypeNo() uint
	Id() common.Hash
}

// Args 新建交易时传入的参数结构体的接口。这样子做可以省掉上一版本中ParseArgs的步骤
type Args interface {

	// Check 只对Args的格式规范进行约束，像余额等等需要进行查询EAccounts或是Chain的检查移交给上层去做
	Check() (err error)
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
		args, ok := args.(*GeneralArgs)
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
		args, ok := args.(*GeneralArgs)
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