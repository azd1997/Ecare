package transaction

import (
	"github.com/azd1997/Ecare/ecoin/crypto"
	"github.com/azd1997/Ecare/ecoin/storage"
	"github.com/azd1997/Ecare/ecoin/utils"
)

// 由于想要实现的是，调用方不知道每一种交易内部实现，因此，调用方只需要使用接口类型和接口方法
// 而且由于transaction包调用太过频繁，是核心业务，所以尽量缩短命名
// 因此这里接口命名不使用Ixxx命名方式，直接使用大写缩写TX

// TX 标志一笔交易，接口
type TX interface {
	String() string
	Serialize() (result []byte, err error)
	Deserialize(data []byte) (err error)
	Hash() (id crypto.Hash, err error)

	// 函数是个很特殊的存在，只要签名匹配，即可以传入函数，也可以传入方法（裹挟着方法接收者的信息）
	// 这有利于包的独立性
	IsValid(ValidateTxFunc) (err error)

	TypeNo() uint
	ID() crypto.Hash
	//Response() *Response
}

// Args 新建交易时传入的参数结构体的接口。这样子做可以省掉上一版本中ParseArgs的步骤
type Args interface {
	// Check 只对Args的格式规范进行约束，像余额等等需要进行查询EAccounts或是Chain的检查移交给上层去做
	Check(CheckArgsFunc) (err error)
}

// NewTXWithArgsCheck 新建一个交易，传入交易类型与其他参数，构建具体的交易。
func NewTXWithArgsCheck(typ uint, args Args, argsFunc CheckArgsFunc) (TX, error) {
	switch typ {
	case TX_COINBASE:
		// 1. 检查参数
		args, ok := args.(*CoinbaseArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxCoinbase(args) // *TxCoinbase 实现了 TX 接口， 粗略的可以认为一个×TxCoinbase是一个TX
	case TX_GENERAL:
		// 1. 检查参数
		args, ok := args.(*GeneralArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxGeneral(args)
	case TX_R2P:
		// 1. 检查参数
		args, ok := args.(*R2PArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxR2P(args)
	case TX_P2R:
		// 1. 检查参数
		args, ok := args.(*P2RArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxP2R(args)
	case TX_P2H:
		// 1. 检查参数
		args, ok := args.(*P2HArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxP2H(args)
	case TX_H2P:
		// 1. 检查参数
		args, ok := args.(*H2PArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxH2P(args)
	case TX_P2D:
		// 1. 检查参数
		args, ok := args.(*P2DArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxP2D(args)
	case TX_D2P:
		// 1. 检查参数
		args, ok := args.(*D2PArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxD2P(args)
	case TX_ARBITRATE:
		// 1. 检查参数
		args, ok := args.(*ArbitrateArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		if err := args.Check(argsFunc); err != nil {
			return nil, err
		}
		// 2. 新建交易
		return newTxArbitrate(args)
	default:
		return nil, ErrUnknownTxType
	}
}

// NewTX 新建一个交易，传入交易类型与其他参数，构建具体的交易。 一定要严格检查输入参数顺序和类型！！！
func NewTX(typ uint, args Args) (TX, error) {
	switch typ {
	case TX_COINBASE:
		// 1. 检查参数
		args, ok := args.(*CoinbaseArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxCoinbase(args) // *TxCoinbase 实现了 TX 接口， 粗略的可以认为一个×TxCoinbase是一个TX
	case TX_GENERAL:
		// 1. 检查参数
		args, ok := args.(*GeneralArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxGeneral(args)
	case TX_R2P:
		// 1. 检查参数
		args, ok := args.(*R2PArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxR2P(args)
	case TX_P2R:
		// 1. 检查参数
		args, ok := args.(*P2RArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxP2R(args)
	case TX_P2H:
		// 1. 检查参数
		args, ok := args.(*P2HArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxP2H(args)
	case TX_H2P:
		// 1. 检查参数
		args, ok := args.(*H2PArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxH2P(args)
	case TX_P2D:
		// 1. 检查参数
		args, ok := args.(*P2DArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxP2D(args)
	case TX_D2P:
		// 1. 检查参数
		args, ok := args.(*D2PArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxD2P(args)
	case TX_ARBITRATE:
		// 1. 检查参数
		args, ok := args.(*ArbitrateArgs)
		if !ok {
			return nil, ErrWrongArgs
		}
		// 2. 新建交易
		return newTxArbitrate(args)
	default:
		return nil, ErrUnknownTxType
	}
}

// DeserializeTX 根据指定具体交易类型编号进行反序列化
func DeserializeTX(typ uint, txBytes []byte) (tx TX, err error) {
	switch typ {
	case TX_COINBASE:
		tx = &TxCoinbase{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_GENERAL:
		tx = &TxGeneral{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_R2P:
		tx = &TxR2P{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_P2R:
		tx = &TxP2R{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_P2H:
		tx = &TxP2H{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_H2P:
		tx = &TxH2P{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_P2D:
		tx = &TxP2D{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_D2P:
		tx = &TxD2P{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
		}
		return tx, nil
	case TX_ARBITRATE:
		tx = &TxArbitrate{}
		err = tx.Deserialize(txBytes)
		if err != nil {
			return nil, utils.WrapError("DeserializeTX", err)
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
		return nil, utils.WrapError("DeserializeTX", ErrNotTxBytes)
	default:
		return nil, utils.WrapError("DeserializeTX", ErrNotTxBytes)
	}
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

// Response 回应接口。 一切沟通皆回应。一层包一层，每一个Response都包含国网所有交易沟通细节
// 暂未使用
type Response interface {
	SourceTx() TX
	String() string
	TargetData() storage.TargetData
	Get(string) []byte // Get("diagnose") ; Get("datakey") ;

	//Serialize() []byte
	//Check()
}

// TODO: 再上一个版本耦合状态下，这里的交易验证和参数验证可以将GSM这类全局结构体传入
// 但现在不行，如果引入的话会导致循环引用。除非使用中间接口来解耦
// 比如说这里的交易类都实现了交易接口，将交易接口独立放置在一个包，然后本包引用接口包和gsm包，gsm包引用接口包
// 这里不选择接口解耦方式，而是将所有需要到全局结构体进行查询的校验内容工作交给上方去做
// 就像account包一样，所有包只做自己能做的事，把需要协作的工作交给上一层处理


type ValidateTxFunc func(tx TX) error

type CheckArgsFunc func(args Args) error

