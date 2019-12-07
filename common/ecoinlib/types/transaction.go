package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
	"log"
	"time"
)

// todo: 之后将response规范为结构体，便于审查

// todo: 新想法： 虽然病人的心电诊断类交易没法进行仲裁，但是可以考虑后面增加一个病人再返回的评论或者说反馈交易。

// TODO: 由于在新建交易的函数内之前采取了传递...interface{}而后再解析再检查的策略，使函数调用时显得比较难看，
//  考虑将这一版本注释掉，重新处理新建交易的传参。将参数封装为结构体！

// NOTICE: 为了保证Transaction的底层独立性，而且在new函数中作了一些基础的检验， 所以不提供verify方法
// （verify需要检查余额是否足够、签名是否有效，必然会和其他库文件耦合）。

// NOTICE: 暂定 医院H编号为1， 研究机构R编号为2， 医生D编号为10， 病人P编号为11

// NOTICE: TODO: 交易的验证：
// 交易检查的内容：
// a. 交易双方地址的有效性、角色的权限、转帐者的余额是否足够、转账者签名是否与其公钥匹配
// b. r2p、p2h、p2d需要检查目标数据是否在broker中
// c. p2r需要检查返回的数据使用凭证是否可用（这里考虑是否增加三段交易：用以做交易的反馈）
// 在这些检查项里边bc无疑是在验证节点（转发节点）本机就需要建立网络连接才能验证的，需要和broker连接才能进行验证。
// 这种网络操作是比较耗时的，应认为避免在验证阶段去做这个事情，而是交给交易双方自行判断。其他的验证则比较容易交给TX类本身来实现。
// 那么问题来了：交易双方如果发现对方发来的交易数据有问题，应该怎么做呢？（想象一下现实生活中淘宝购物的过程：用户下单 -> 商家发货 -> 用户发现商品有问题然后退货 -> 商家重新发货 -> ... ）
// 也就是说要将这个过程描述，需要两类交易。以病人和研究机构的交易过程来讲就是： r2p -> p2r -> r2p -> p2r -> ...
// 怎么表示交易完成而不互相扯皮呢？
// 交易发起方交易增加一个标志：txComplete（买家同意说交易确实完成了）。
// r2p（交易发起方）中若将之置为true，则其转账金额（临时扣除，记录在问“未完成交易池内”）退回；
// r2p complete : 交易发起方直接付钱给对方了，对方并不会返回结果给你
// r2p !complete -> p2r -> r2p complete : 交易发起方接收到对方的应答，认为是可以的，所以交易完成，金额支付给到对方
// r2p !complete -> p2r -> r2p !complete ?
// 这个过程天然是保护买家的权益的，买家若不满意只需要不停地！complete，交易就永远不会完成，钱不会到卖家那。当然钱也就卡在了“未完成交易池内”。
// 带来的问题：
// 1. 确实是卖家作恶给了假商品，但僵持下去买家的钱被卡住了。
// 2. 买家作恶，卖家收不到钱。
// 怎么解决？
// 问题1解决方法可以是：
// 僵持三次后如果买家仍然 ！complete 则出块节点检查区块中交易时进行仲裁，也就是返回结果的验证，
// 并以仲裁结果来更新全局状态机，而验证节点再进行验证。而且还可以据此判断哪一方信用出现问题，可以进行信用评分（暂时不实现，TODO）
// 这个解决办法怎么实现呢？(怎么知道已经僵持三次了)： 三种：一是每次检查这种交易时都去区块链历史中不断通过交易查找找到交易的前部交易。
// 这种做法需要从后往前遍历区块再遍历区块内交易，直到找到有一个r2p其内的p2r指针为nil（表示是这个交易总体的最开始）。然后得到就知道这个交易
// 是总体交易的第几个回合（一个回合指r2p -> p2r）。这个做法一个是遍历比较耗时（迭代而且是数据库操作），另一个是其他常规的正常的交易也要进行这种检查，
// 严重浪费。 二是在交易体内增加回合字段，标记到了哪个第几个回合。这个做法是比较可行的。 当然，检查交易时需要检查标记的回合是否是真的，而这种做法是检查不了的。
// 三是 交易体内不存放前一个交易指针（哈希值），
// 而是存前一个交易的[]byte，那么每次检查交易都把全部交易展开，这种做法对于性能的损伤不是很大，而且好处是不用去搜索区块链数据库就可以直到到了第几个回合。
// 等等，这种做法意味着要检查折叠的交易是否真的已经发布过，还是得遍历区块链数据库去匹配。但比做法一好一点，这种做法只需要检查最外层折叠交易是否存在即可。
// 总的来说，做法三是最合适的，因为这种做法在处理其他操作时也有很多好处。节省了很多遍历操作。
// 此外！这种情况下区块链向前遍历可以设置一个最大遍历数？ 因为在r2p的交易过程中r端是自动完成的（解析p发来的数据然后去自动尝试获取目标数据），所以也就是说
// 当发生这种三次僵持时，前一个交易（p2r）和当前交易（p2r）的间隔是遍历区间。由于这个时间段是用户处理的时间，是不确定的，有可能很长，也有可能很短，
// 这就意味着要遍历整个区块链！这种情况下是不能设置最大遍历数的。
// 等等！其实是可以精确定位到该交易的！
// 考虑 r2p{p2r} 。构造r2p时 p2r已经是被承认被出块了的也被r端轮询得到了，这样的话r端是完全可以知道p2r是在哪一个区块中发现的！！！
// 那么在r2p构造时额外提供一个字段：前部交易的所在区块号！那么检查交易时直接查这一个区块！没有说明交易不合法！
// 至此 问题1 得到了一个比较好的解决办法了
// 问题2呢？买家作恶，也就是r作恶，p收不到钱。
// 一样的，三次僵持自动仲裁。p2r侧交易体内加一个 txComplete(卖家认为应该是完成了)，同样的三次僵持之后，如果 p2r txComplete，表示卖家（p，病人）
// 认为自己是给的正确的，则自动仲裁 (这句看法是错的，p端可以不停地txComplete)
// 重新回到该问题：卖家正常发货，买家收到之后却不回复确认交易（也就是第二次r2p）甚至是不停的发r2p说卖家给假东西，卖家只能重新发，但买家继续僵持使卖家收不到钱。
//  r2p -> p2r -> r2p (买家是否认为交易完成) -> p2r -> ...
// 这里有两个子问题：（1）买家不回复r2p；（2）买家不断不断发起r2p(!complete)。这个自然而然的被上面问题一的解决方案解决了。
// 那么问题在于问题2(1)区块打包者和卖家一直没得到r2p（回复），卖家就没办法申诉仲裁，矿工也不知道有这个情况。
// 自然而然的，想到设置超时：比如卖家超过一天还是没找到这个回复则主动构建一个申诉交易（由于没有上个r2p）所以需要一个新的交易体。来告诉矿工去处理。
// 这是一个解决办法。只是这个超时时间存在两个问题：一是设多少合适？二是尽量希望能够保留很长时间，给一些出问题的节点一些机会；三是这个时间由卖家掌控了
// 留下了攻击隐患。
// 考虑一种新的办法： 转发节点维护“未完成交易池”，前面想到的交易池其实存的只是待转账记录，这里则把意义扩大了，
// 节点收到了这类交易（三类:r2p对、p2h对、p2d对），后检查其有没有问题后打包到区块，同时添加入“未完成交易池”，当接收到r2p三次僵持自动仲裁，
// 仲裁结果出来后，每一个验证节点若认为谁对，则把钱交给交易的哪一方。这个过程是没有重新构建交易的，是每个转发节点在本地做这个事情。
// 交易池内每个交易体新增一个定时器和通道，时间一到告诉调用线程有哪个悬而未决的交易需要处理。
// 这样做有个问题，这种事件的处理在每个节点上是各自进行的，所以进行的时间先后不一致。假设 A -> B 金额是5，A还剩10，假如说仲裁结果是A的钱应该退回
// 但由于验证节点做这个事情时间不一致，所以如果A要进行一笔消费为6（超过5）的交易，可能会在有些还没有进行仲裁的节点上被拒绝。
// 另外一个问题是当事人如何知道仲裁结果。
// 解决办法： 每个出块节点在出块时检查未完成交易池，看有没有超时交易或者三次僵持了的（称为”超时仲裁“和”三次僵持仲裁“），
// 这两种情况的交易出块节点将之进行仲裁，构建新的仲裁交易事件，这就保证了转发节点基本是同步得到仲裁结果的。
// 至于当事人如何知道，还是在同步区块链过程中查看交易内容发现与自己有关，再处理
//
// 其他两类心电诊断类交易则是没办法去验证卖家（医院或者医生）返回结果的有效性。最多只能对返回消息的格式做一个规定。
// 因此这种情况下，只要返回交易符合格式，就只能认为应该得到报酬，病人只能自动构建交易结束，验证节点们修改全局状态。
// 好了，这些问题都得到解决了。
//
// 有一个新的问题：如果是p2h、p2d交易p收到交易结果不回复怎么办？
// 对于这类交易不设complete字段，医院构建合法的回应段交易交易就生效了。
//
// 关于心电诊断交易，TODO： 以后修改成病人发起任务，可以由医院机器自动诊断 + 选取多位医生都进行诊断，这样的诊断结果对于病人是比较合适的。
//
// 另一个问题： 为什么我不直接设一个字段“是否申请仲裁”？避免被恶意攻击，如果有攻击者故意制造交易不断申请仲裁，这会拖慢区块产生速度，影响较大。
// 但是这样的三次僵持，一样还是会有这种可能性啊？？？
// 仲裁事件只出现在r与p之间这种可以由系统自动去“绝对判断”的交易情形，对于交易双方而言是发生在这个系统中的商业活动。
// 整个系统应该是以病人的需求为第一优先目标的，所以这种商业交易（包括以后增加的其他这种模式的交易）都应该尽量将耗时的验证交由交易双方自己去做。
// 所以“三次僵持”的目的不仅仅是为了解决自动处理交易矛盾，也是为了鼓励商业交易双方诚信交易，让交易双方自己承担不诚信交易浪费的时间和性能代价。
// 同时也提供给这类商业交易多次磋商的机会，
// 从这个角度来看，依然采取 三次僵持 。

// 综上，格式方面的校验由交易类本身实现，出块节点和验证节点在检查时调用即可。而交易的实质内容是不是正确由交易双方自行判断，对于商业性质交易采取三次僵持策略
// 对于病人心电诊断类交易采取回应即生效策略

// 综上，需要进行的修改是：
// TX接口增加verify方法，传入*gsm来验证一些基本的信息以及格式是否无误。
// TxR2P增加p2rBytes和txComplete属性
// gsm增加 未处理交易池 。目前存入其中的只会是txr2p、p2h、p2d类
// 增加 仲裁交易

// 在类型修改的基础上修改方法

// 1. 构建交易时本身就作了交易构造时的校验： 转出账户是否余额足够、接收账户地址格式是否正确（转账地址是否存在由区块打包时做检查）、转账内容是否有效
// 2. 构建区块时和接收区块时对区块内交易做检查：

// NOTICE: 交易二段是怎么实现的？节点在线期间会根据设置的查询周期（比如一个小时）去遍历这段时间新增的区块内交易有没有与自己相关的交易，
// 有就展示到应用界面，当用户回应了这个交易，则自动将该交易从展示列表移除。
// 假设提供一个轮询开关，如果用户关闭了在线轮询，或者是刚上线，这时怎么去控制迭代范围避免遍历整个区块链呢？增加一个记录点，标志每次轮询区间的最末尾区块的编号
// 这个问题就解决了

// 交易要检查：
// 1.转账者、接收者是否存在
// 2.转账金额非负为整
// 3.转账者余额是否足够

// 一笔交易由转账者构建，A当然可以创建这个交易，但这个问题在于怎么确保其他人无法创建以A的地址和签名的交易

// 注意：结构体转json只会转导出元素，开头小写的属性不会被转为json

const (
	ECG_DIAG_AUTO = iota + 1
	ECG_DIAG_DOCTOR

	TX_AUTO = iota
	TX_COINBASE
	TX_GENERAL
	TX_R2P
	TX_P2R
	TX_P2H
	TX_H2P
	TX_P2D
	TX_D2P
	TX_ARBITRATE
)

// Hash 32B哈希。如果要修改哈希算法，只需在这里重新定义哈希的具体类型即可
type Hash [32]byte

// BytesToHash 将长度为32的字节切片转换为Hash
func BytesToHash(data []byte) Hash {
	var res Hash
	if len(data) != cap(res) {
		log.Fatal("BytesToHash needs a length-32 byte slice")
	}

	for i := 0; i < cap(res); i++ {
		res[i] = data[i]
	}

	return res
}

// UnixTimeStamp Unix时间戳
type UnixTimeStamp uint64

// Coin 数字货币
type Coin uint

// Signature 签名
type Signature []byte

// TargetData 目标数据，在这里表示所要查找的心电数据记录
type TargetData struct {
	StartTime     UnixTimeStamp // =0 表示不填
	EndTime       UnixTimeStamp
	NumsOfRecords uint // 若start, end均已正常设置，则该项无效
}

// IsValid 检查目标数据是否存在可取用
func (t *TargetData) IsOk(storage DataStorage) (ok bool, err error) {
	// 1. 从TargetData解析索引

	// 2. 去查询数据是否在指定broker中
	// TODO: 实现DataStorage接口，传入结构体指针，用以查询
	ok, err = storage.Query(t.StartTime, t.EndTime, t.NumsOfRecords)

	return ok, err
}

// Serializer 序列化接口，本项目中block和tx实现了这个接口
type Serializer interface {
	Serialize() (result []byte, err error)
}

// Hasher 取哈希接口，本项目中block和tx实现了这个接口
type Hasher interface {
	Hash() (hash Hash, err error)
}

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
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_GENERAL:
		tx = &TxGeneral{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_R2P:
		tx = &TxR2P{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_P2R:
		tx = &TxP2R{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_P2H:
		tx = &TxP2H{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_H2P:
		tx = &TxH2P{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_P2D:
		tx = &TxP2D{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_D2P:
		tx = &TxD2P{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
	case TX_ARBITRATE:
		tx = &TxArbitrate{}
		err = tx.Deserialize(txBytes)
		return tx, utils.WrapError("DeserializeTX", err)
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
		return nil, utils.WrapError("DeserializeTX", ErrUnknownTransactionType)
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

/*********************************************************************************************************************
                                                    ArgsOfNewTX接口
*********************************************************************************************************************/

// ArgsOfNewTX 新建交易时传入的参数结构体的接口。这样子做可以省掉上一版本中ParseArgs的步骤
type ArgsOfNewTX interface {
	CheckArgsValue(gsm *GlobalStateMachine) (err error)
}

//// BaseArgs 基本的参数项，这里其实是放了gsm *GlobalStateMachine进去
//type BaseArgs struct {
//	Gsm *GlobalStateMachine
//}

/*********************************************************************************************************************
                                                    TxCoinbase相关
*********************************************************************************************************************/

// TxCoinbaseArgs 新建交易函数newTxCoinbase()的传参
type TxCoinbaseArgs struct {
	//	BaseArgs
	To          UserID
	Amount      Coin
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxCoinbaseArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(gsm.opts.ChecksumLength()); !valid {
		return ErrInvalidUserID
	}
	// coinbase交易只允许出块节点构建，而出块节点的roleNo 0~9
	if args.To.RoleNo > 9 {
		return ErrInvalidUserID
	}

	// 检查 amount 有效性
	// TODO: 检查coinbase奖励是否合乎规则

	// TODO: 检查 description 格式，以及代码注入？

	return nil
}

// TxCoinbase 出块奖励交易，只允许A类账户接收，A类账户目前包括医院H和第三方研究机构R
// 由于coinbase交易没有转账者，且必须由出块者构建，所以不设置签名项划定归属。
type TxCoinbase struct {
	BaseTransaction `json:"baseTransaction"`
}

// newTxCoinbase 新建出块奖励交易。
func newTxCoinbase(args *TxCoinbaseArgs) (tx *TxCoinbase, err error) {
	// TODO： 注意： 参数的检查交给gsm去做了
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxCoinbase", err)
	//}

	// 构造tx
	tx = &TxCoinbase{
		BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxCoinbase", err)
	}
	tx.ID = id
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxCoinbase) TypeNo() uint {
	return TX_COINBASE
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxCoinbase) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID = Hash{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxCoinbase_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxCoinbase) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxCoinbase) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxCoinbase{} 再调用
func (tx *TxCoinbase) Deserialize(data []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	tx1 := &TxCoinbase{}
	if tx != tx1 {
		return utils.WrapError("TxCoinbase_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(data)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxCoinbase_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxCoinbase) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxCoinbase{
		BaseTransaction:BaseTransaction{
			Id:Hash{},
			Time:UnixTimeStamp(0),
			To:UserID{},
			Amount:Coin(1),
			Description:string(""),
		}}*/

	// 要记住检验交易有两种情况下被调用：一是加入未打包交易池之前要检查交易（情况A）；二是收到区块后要检查区块内交易（情况B）。

	// 检查时间戳是否比现在早（至于是不是早太多就不检查了，早太多的话余额那里是不会给过的）（情况A）； 时间戳是否比区块时间早（情况B）
	// 但是要注意情况A调用检查一定比情况B早，所以只要满足情况A就一定满足情况B (或者说，如果情况A不通过，也就不会进入到情况B检查)。所以，只检查情况A就好
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxCoinbase_IsValid", ErrWrongTimeTX)
	}

	// 检查coinbase接收者ID的有效性和角色的权限与可用性
	userIDValid, _ := tx.To.IsValid(gsm.opts.ChecksumLength()) // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return utils.WrapError("TxCoinbase_IsValid", ErrInvalidUserID)
	}
	if tx.To.RoleNo >= 10 {
		return utils.WrapError("TxCoinbase_IsValid", ErrNoCoinbasePermitRole)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return utils.WrapError("TxCoinbase_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return utils.WrapError("TxCoinbase_IsValid", ErrUnavailableUserID)
	}

	// 检查coinbase金额
	if tx.Amount != toEcoinAccount.Role().CoinbaseReward() {
		return utils.WrapError("TxCoinbase_IsValid", ErrWrongCoinbaseReward)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxCoinbase_IsValid", ErrWrongTXID)
	}

	// TODO： Coinbase还有一个检查点：其由出块节点构造，但在验证过程中必须检查是不是填了出块节点账户。因此在出块节点检查区块时需要有一个区块的检查方法
	// 而这个方法检查所有交易有效性，并对coinbase（在打包交易时始终放在交易列表第一位）再增加这一个处理。
	// 如果要在这里做这个检查，就必须穿入一个*Block作参数。但是其他类型交易不需要这个参数，会破坏整体接口的实现。

	return nil
}

/*********************************************************************************************************************
                                                    TxGeneral相关
*********************************************************************************************************************/

// TxGeneralArgs 新建交易函数newTxGeneral()的传参
type TxGeneralArgs struct {
	//	BaseArgs
	From        *Account
	FromID      UserID
	To          UserID
	Amount      Coin
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxGeneralArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.FromID != fromID {
		return ErrWrongArguments
	}

	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(gsm.opts.ChecksumLength()); !valid {
		return ErrInvalidUserID
	}

	// 检查 amount 有效性(余额是否足够)
	selfId, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.Amount > gsm.Accounts.Map[selfId.ID].Balance() {
		return ErrNotSufficientBalance
	}

	// TODO: 检查 description 格式，以及代码注入？

	return nil
}

// TxGeneral 通用交易， 一方转给另一方，无需确认
type TxGeneral struct {
	BaseTransaction `json:"baseTransaction"`
	From            UserID    `json:"from"`
	Sig             Signature `json:"sig"`
}

// newTxGeneral 新建普通转账交易。
func newTxGeneral(args *TxGeneralArgs) (tx *TxGeneral, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxGeneral", err)
	//}

	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxGeneral", err)
	//}

	// 构造tx
	tx = &TxGeneral{
		BaseTransaction: BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
		From: args.FromID,
		Sig:  Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxGeneral", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxGeneral", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxGeneral) TypeNo() uint {
	return TX_GENERAL
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxGeneral) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{} // 置空值
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxGeneral_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxGeneral) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxGeneral) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxGeneral{} 再调用
func (tx *TxGeneral) Deserialize(data []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	tx1 := &TxGeneral{}
	if tx != tx1 {
		return utils.WrapError("TxGeneral_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(data)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxGeneral_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxGeneral) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxGeneral{
			BaseTransaction: BaseTransaction{
				Id:          Hash{},
				Time:        UnixTimeStamp(time.Now().Unix()),
				To:          to,
				Amount:      amount,
				Description: description,
			},
			From: fromID,
			Sig:  Signature{},
		}
	*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxGeneral_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	userIDValid, _ := tx.To.IsValid(gsm.opts.ChecksumLength()) // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return utils.WrapError("TxGeneral_IsValid", ErrInvalidUserID)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return utils.WrapError("TxGeneral_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return utils.WrapError("TxGeneral_IsValid", ErrUnavailableUserID)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	userIDValid, _ = tx.From.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxGeneral_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return utils.WrapError("TxGeneral_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return utils.WrapError("TxGeneral_IsValid", ErrUnavailableUserID)
	}
	if tx.Amount > fromEcoinAccount.Balance() {
		return utils.WrapError("TxGeneral_IsValid", ErrNotSufficientBalance)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return utils.WrapError("TxGeneral_IsValid", ErrInconsistentSignature)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxGeneral_IsValid", ErrWrongTXID)
	}

	return nil
}

/*********************************************************************************************************************
                                                    TxR2P相关
*********************************************************************************************************************/

// TxR2PArgs 新建交易函数newTxR2P()的传参
type TxR2PArgs struct {
	//	BaseArgs
	From           *Account
	FromID         UserID
	To             UserID
	Amount         Coin
	Description    string
	PurchaseTarget TargetData
	P2RBytes       []byte
	TxComplete     bool
	Storage        DataStorage
}

// CheckArgsValue 检查参数值是否合规
func (args *TxR2PArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.FromID != fromID {
		return ErrWrongArguments
	}

	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(gsm.opts.ChecksumLength()); !valid {
		return ErrInvalidUserID
	}

	// 检查 amount 有效性(余额是否足够)
	selfId, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.Amount > gsm.Accounts.Map[selfId.ID].Balance() {
		return ErrNotSufficientBalance
	}

	// TODO: 检查 description 格式，以及代码注入？

	// 检查storage是否有效
	if !args.Storage.IsOk() {
		return ErrNotOkStorage
	}

	// 检查 purchaseTarget是否存在？
	if ok, _ := args.PurchaseTarget.IsOk(args.Storage); !ok {
		return ErrNonexistentTargetData
	}

	// 检验p2rBytes。 要么是[]byte{}(表示是初始交易)，要么是可以反序列化为p2r交易
	if bytes.Compare(args.P2RBytes, []byte{}) != 0 {
		p2r := TxP2R{}
		if err = p2r.Deserialize(args.P2RBytes); err != nil {
			return ErrWrongSourceTX
		}
		// 检查p2r.ID是否在未完成交易池
		if _, ok := gsm.UCTXP.Map[p2r.ID]; !ok {
			return ErrTXNotInUCTXP
		}
		// 检查p2r.From是否为args.To
		if p2r.From != args.To {
			return ErrUnmatchedTxReceiver
		}
	}

	// 参数有效
	return nil
}

// TxR2P 第三方研究机构向病人发起的数据交易的阶段一交易
type TxR2P struct {
	BaseTransaction `json:"baseTransaction"`
	From            UserID     `json:"from"`
	Sig             Signature  `json:"sig"`
	PurchaseTarget  TargetData `json:"purchaseTarget"`
	P2RBytes        []byte     `json:"p2rBytes, omitempty"`
	TxComplete      bool       `json:"txComplete"` // 注意：在上层调用也就是block类中验证交易时，需要检查txComplete来进行“三次僵持“策略的实现
}

// newTxR2P 新建R2P转账交易。
func newTxR2P(args *TxR2PArgs) (tx *TxR2P, err error) {

	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxR2P", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxR2P", err)
	//}

	// 构造tx
	tx = &TxR2P{
		BaseTransaction: BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
		From:           args.FromID,
		Sig:            Signature{},
		PurchaseTarget: args.PurchaseTarget,
		P2RBytes:       args.P2RBytes,
		TxComplete:     args.TxComplete,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxR2P", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxR2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

// commercial 商业性质
func (tx *TxR2P) commercial() {
	// 啥事也不干
}

// TypeNo 获取交易类型编号
func (tx *TxR2P) TypeNo() uint {
	return TX_R2P
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxR2P) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, fmt.Errorf("TxGeneral_Hash: %s", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxR2P) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxR2P) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxR2P{} 再调用
func (tx *TxR2P) Deserialize(r2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	tx1 := &TxR2P{}
	if tx != tx1 {
		return utils.WrapError("TxR2P_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(r2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxR2P_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxR2P) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxR2P{
		BaseTransaction: BaseTransaction{
			Id:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          to,
			Amount:      amount,
			Description: description,
		},
		From:           fromID,
		Sig:            Signature{},
		PurchaseTarget: purchaseTarget,
		P2RBytes:       p2rBytes,
		TxComplete:     txComplete,
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxR2P_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	userIDValid, _ := tx.To.IsValid(gsm.opts.ChecksumLength()) // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return utils.WrapError("TxR2P_IsValid", ErrInvalidUserID)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return utils.WrapError("TxR2P_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return utils.WrapError("TxR2P_IsValid", ErrUnavailableUserID)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	userIDValid, _ = tx.From.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxR2P_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return utils.WrapError("TxR2P_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return utils.WrapError("TxR2P_IsValid", ErrUnavailableUserID)
	}
	if tx.Amount > fromEcoinAccount.Balance() {
		return utils.WrapError("TxR2P_IsValid", ErrNotSufficientBalance)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return utils.WrapError("TxR2P_IsValid", ErrInconsistentSignature)
	}

	// TODO： PurchaseTarget可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个P2R交易，为空则正确；不为空必须是符合P2R交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.P2RBytes == nil || bytes.Compare(tx.P2RBytes, []byte{}) == 0 {
		return utils.WrapError("TxR2P_IsValid", ErrEmptySoureTX)
	}
	if bytes.Compare(tx.P2RBytes, []byte{}) != 0 {
		prevTx := &TxP2R{}
		err := prevTx.Deserialize(tx.P2RBytes)
		if err != nil {
			return utils.WrapError("TxR2P_IsValid", err)
		}
		if _, ok := gsm.UCTXP.Map[prevTx.ID]; !ok {
			return utils.WrapError("TxR2P_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxR2P_IsValid", ErrWrongTXID)
	}

	return nil
}

/*********************************************************************************************************************
                                                    TxP2R相关
*********************************************************************************************************************/

// TxP2RArgs 新建交易函数newTxP2R()的传参
type TxP2RArgs struct {
	//	BaseArgs
	From        *Account
	FromID      UserID
	R2PBytes    []byte
	Response    []byte
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxP2RArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.FromID != fromID {
		return ErrWrongArguments
	}

	// 检查r2pBytes
	r2p := &TxR2P{}
	if err = r2p.Deserialize(args.R2PBytes); err != nil {
		return ErrNotTxBytes
	}
	// 检查r2p是否在未完成交易池中
	if _, ok := gsm.UCTXP.Map[r2p.ID]; !ok {
		return ErrTXNotInUCTXP
	}
	// 检查r2p内to是否和此时的from对应，都是本机拥有的账户
	selfId, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if selfId != r2p.To {
		return ErrUnmatchedTxReceiver
	}

	// TODO: 检查 response 有效性

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}

// TxP2R 第三方研究机构向病人发起的数据交易的阶段二交易
type TxP2R struct {
	ID          Hash          `json:"id"`
	Time        UnixTimeStamp `json:"time"`
	From        UserID        `json:"from"`
	R2PBytes    []byte        `json:"r2pBytes"`
	Response    []byte        `json:"response"` // 比如说请求数据的密码
	Description string        `json:"description"`
	Sig         Signature     `json:"sig"`
}

// newTxP2R 新建P2R转账交易(R2P交易二段)。
func newTxP2R(args *TxP2RArgs) (tx *TxP2R, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxP2R", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxP2R", err)
	//}

	// 构造tx
	tx = &TxP2R{
		ID:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:        args.FromID,
		R2PBytes:    args.R2PBytes,
		Response:    args.Response,
		Description: args.Description,
		Sig:         Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxP2R", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxP2R", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxP2R) TypeNo() uint {
	return TX_P2R
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2R) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxP2R_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2R) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2R) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxR2P{} 再调用
func (tx *TxP2R) Deserialize(p2rBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	tx1 := &TxP2R{}
	if tx != tx1 {
		return utils.WrapError("TxP2R_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2rBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxP2R_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2R) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxP2R{
		Id:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:fromID,
		R2PBytes:    r2pBytes,
		Response:    response,
		Description: description,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxP2R_IsValid", ErrWrongTimeTX)
	}

	// 检查fromID的有效性、可用性和from签名是否匹配
	userIDValid, _ := tx.From.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxP2R_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return utils.WrapError("TxP2R_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return utils.WrapError("TxP2R_IsValid", ErrUnavailableUserID)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return utils.WrapError("TxP2R_IsValid", ErrInconsistentSignature)
	}

	// TODO： Response可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个R2P交易，为空则正确；不为空必须是符合R2P交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.R2PBytes == nil || bytes.Compare(tx.R2PBytes, []byte{}) == 0 {
		return utils.WrapError("TxP2R_IsValid", ErrEmptySoureTX)
	}
	if bytes.Compare(tx.R2PBytes, []byte{}) != 0 {
		prevTx := &TxR2P{}
		err := prevTx.Deserialize(tx.R2PBytes)
		if err != nil {
			return utils.WrapError("TxP2R_IsValid", err)
		}
		if _, ok := gsm.UCTXP.Map[prevTx.ID]; !ok {
			return utils.WrapError("TxP2R_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxP2R_IsValid", ErrWrongTXID)
	}

	return nil
}

/*********************************************************************************************************************
                                                    TxP2H相关
*********************************************************************************************************************/

// TxP2HArgs 新建交易函数newTxP2H()的传参
type TxP2HArgs struct {
	//	BaseArgs
	From           *Account
	FromID         UserID
	To             UserID
	Amount         Coin
	Description    string
	PurchaseTarget TargetData
	PurchaseType   uint8
	Storage        DataStorage
}

// CheckArgsValue 检查参数值是否合规
func (args *TxP2HArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.FromID != fromID {
		return ErrWrongArguments
	}

	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(gsm.opts.ChecksumLength()); !valid {
		return ErrInvalidUserID
	}
	if args.To.RoleNo != 1 {
		return ErrWrongRoleUserID
	}

	// 检查 amount 有效性(余额是否足够)
	selfId, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.Amount > gsm.Accounts.Map[selfId.ID].Balance() {
		return ErrNotSufficientBalance
	}

	// TODO: 检查 description 格式，以及代码注入？

	// 检查storage是否有效
	if !args.Storage.IsOk() {
		return ErrNotOkStorage
	}

	// 检查 purchaseTarget是否存在？
	if ok, _ := args.PurchaseTarget.IsOk(args.Storage); !ok {
		return ErrNonexistentTargetData
	}

	// 检查 purchaseType
	if args.PurchaseType != ECG_DIAG_AUTO && args.PurchaseType != ECG_DIAG_DOCTOR {
		return ErrUnknownPurchaseType
	}

	// 参数有效
	return nil
}

// TxP2H 病人向医院发起的心电数据诊断，分人工和机器自动分析两种。阶段一
type TxP2H struct {
	BaseTransaction `json:"baseTransaction"`
	From            UserID     `json:"from"`
	Sig             Signature  `json:"sig"`
	PurchaseTarget  TargetData `json:"purchaseTarget"`
	PurchaseType    uint8      `json:"purchaseType"` // Auto/Doctor 0/1
}

// newTxP2H 新建P2H转账交易。
func newTxP2H(args *TxP2HArgs) (tx *TxP2H, err error) {

	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxP2H", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxP2H", err)
	//}

	// 构造tx
	tx = &TxP2H{
		BaseTransaction: BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
		From:           args.FromID,
		Sig:            Signature{},
		PurchaseTarget: args.PurchaseTarget,
		PurchaseType:   args.PurchaseType,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxP2H", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxP2H", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxP2H) TypeNo() uint {
	return TX_P2H
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2H) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxP2H_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2H) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2H) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxP2H{} 再调用
func (tx *TxP2H) Deserialize(p2hBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	tx1 := &TxP2H{}
	if tx != tx1 {
		return utils.WrapError("TxP2H_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2hBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxP2H_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2H) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxP2H{
		BaseTransaction: BaseTransaction{
			Id:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          to,
			Amount:      amount,
			Description: description,
		},
		From:           fromID,
		Sig:            Signature{},
		PurchaseTarget: purchaseTarget,
		PurchaseType:   purchaseType,
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxP2H_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	userIDValid, _ := tx.To.IsValid(gsm.opts.ChecksumLength()) // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return utils.WrapError("TxP2H_IsValid", ErrInvalidUserID)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return utils.WrapError("TxP2H_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return utils.WrapError("TxP2H_IsValid", ErrUnavailableUserID)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	userIDValid, _ = tx.From.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxP2H_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return utils.WrapError("TxP2H_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return utils.WrapError("TxP2H_IsValid", ErrUnavailableUserID)
	}
	if tx.Amount > fromEcoinAccount.Balance() {
		return utils.WrapError("TxP2H_IsValid", ErrNotSufficientBalance)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return utils.WrapError("TxP2H_IsValid", ErrInconsistentSignature)
	}

	// TODO： PurchaseTarget可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查purchaseType
	if tx.PurchaseType != ECG_DIAG_AUTO && tx.PurchaseType != ECG_DIAG_DOCTOR {
		return utils.WrapError("TxP2H_IsValid", ErrUnknownPurchaseType)
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxP2H_IsValid", ErrWrongTXID)
	}

	return nil
}

/*********************************************************************************************************************
                                                    TxH2P相关
*********************************************************************************************************************/

// TxH2PArgs 新建交易函数newTxH2P()的传参
type TxH2PArgs struct {
	//	BaseArgs
	From        *Account
	FromID      UserID
	P2HBytes    []byte
	Response    []byte
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxH2PArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查FromID
	fromID, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.FromID != fromID {
		return ErrWrongArguments
	}

	// 检查p2hBytes
	p2h := &TxP2H{}
	if err = p2h.Deserialize(args.P2HBytes); err != nil {
		return ErrNotTxBytes
	}
	// 检查p2h是否在未完成交易池中
	if _, ok := gsm.UCTXP.Map[p2h.ID]; !ok {
		return ErrTXNotInUCTXP
	}
	// 检查p2h内to是否和此时的from对应，都是本机拥有的账户
	selfId, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if selfId != p2h.To {
		return ErrUnmatchedTxReceiver
	}

	// TODO: 检查 response 有效性

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}

// TxH2P 病人向医院发起的心电数据诊断，分人工和机器自动分析两种。阶段二
type TxH2P struct {
	ID          Hash          `json:"id"`
	Time        UnixTimeStamp `json:"time"`
	From        UserID        `json:"from"`
	P2HBytes    []byte        `json:"p2hBytes"`
	Response    []byte        `json:"response"` // 比如说请求数据的密码
	Description string        `json:"description"`
	Sig         Signature     `json:"sig"`
}

// newTxH2P 新建H2P转账交易(P2H交易二段)。
func newTxH2P(args *TxH2PArgs) (tx *TxH2P, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxH2P", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxH2P", err)
	//}

	// 构造tx
	tx = &TxH2P{
		ID:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:        args.FromID,
		P2HBytes:    args.P2HBytes,
		Response:    args.Response,
		Description: args.Description,
		Sig:         Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxH2P", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxH2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxH2P) TypeNo() uint {
	return TX_H2P
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxH2P) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxH2P_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxH2P) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxH2P) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxH2P{} 再调用
func (tx *TxH2P) Deserialize(h2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	tx1 := &TxH2P{}
	if tx != tx1 {
		return utils.WrapError("TxH2P_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(h2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxH2P_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxH2P) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxH2P{
		Id:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:fromID,
		P2HBytes:    p2hBytes,
		Response:    response,
		Description: description,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxH2P_IsValid", ErrWrongTimeTX)
	}

	// 检查fromID的有效性、可用性和from签名是否匹配
	userIDValid, _ := tx.From.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxH2P_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return utils.WrapError("TxH2P_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return utils.WrapError("TxH2P_IsValid", ErrUnavailableUserID)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return utils.WrapError("TxH2P_IsValid", ErrInconsistentSignature)
	}

	// TODO： Response可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个P2R交易，为空则正确；不为空必须是符合P2R交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.P2HBytes == nil || bytes.Compare(tx.P2HBytes, []byte{}) == 0 {
		return utils.WrapError("TxH2P_IsValid", ErrEmptySoureTX)
	}
	if bytes.Compare(tx.P2HBytes, []byte{}) != 0 {
		prevTx := &TxP2H{}
		err := prevTx.Deserialize(tx.P2HBytes)
		if err != nil {
			return utils.WrapError("TxH2P_IsValid", err)
		}
		if _, ok := gsm.UCTXP.Map[prevTx.ID]; !ok {
			return utils.WrapError("TxH2P_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxH2P_IsValid", ErrWrongTXID)
	}

	return nil
}

/*********************************************************************************************************************
                                                    TxP2D相关
*********************************************************************************************************************/

// TxP2DArgs 新建交易函数newTxP2D()的传参
type TxP2DArgs struct {
	//	BaseArgs
	From           *Account
	FromID         UserID
	To             UserID
	Amount         Coin
	Description    string
	PurchaseTarget TargetData
	Storage        DataStorage
}

// CheckArgsValue 检查参数值是否合规
func (args *TxP2DArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查 to 的有效性
	if valid, _ := args.To.IsValid(gsm.opts.ChecksumLength()); !valid {
		return ErrInvalidUserID
	}
	// 检查to账号是不是医生（11）
	if args.To.RoleNo != 11 {
		return ErrWrongRoleUserID
	}

	// 检查 amount 有效性(余额是否足够)
	selfId, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if args.Amount > gsm.Accounts.Map[selfId.ID].Balance() {
		return ErrNotSufficientBalance
	}

	// TODO: 检查 description 格式，以及代码注入？

	// 检查storage是否有效
	if !args.Storage.IsOk() {
		return ErrNotOkStorage
	}

	// 检查 purchaseTarget是否存在？
	if ok, _ := args.PurchaseTarget.IsOk(args.Storage); !ok {
		return ErrNonexistentTargetData
	}

	// 参数有效
	return nil
}

// TxP2D 病人向下班医生发起的心电诊断交易，阶段一		TODO: 暂时只支持找指定医生诊断；后边考虑广播交易等待医生解决
type TxP2D struct {
	BaseTransaction `json:"baseTransaction"`
	From            UserID     `json:"from"`
	Sig             Signature  `json:"sig"`
	PurchaseTarget  TargetData `json:"purchaseTarget"`
}

// newTxP2D 新建P2D转账交易。
func newTxP2D(args *TxP2DArgs) (tx *TxP2D, err error) {

	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxP2D", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxP2D", err)
	//}

	// 构造tx
	tx = &TxP2D{
		BaseTransaction: BaseTransaction{
			ID:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          args.To,
			Amount:      args.Amount,
			Description: args.Description,
		},
		From:           args.FromID,
		Sig:            Signature{},
		PurchaseTarget: args.PurchaseTarget,
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxP2D", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxP2D", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxP2D) TypeNo() uint {
	return TX_P2D
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxP2D) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxP2D_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxP2D) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxP2D) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxP2D{} 再调用
func (tx *TxP2D) Deserialize(p2dBytes []byte) (err error) {
	// 防止非空TxP2D调用该方法改变了自身内容
	tx1 := &TxP2D{}
	if tx != tx1 {
		return utils.WrapError("TxP2D_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(p2dBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxP2D_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxP2D) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxP2D{
		BaseTransaction: BaseTransaction{
			Id:          Hash{},
			Time:        UnixTimeStamp(time.Now().Unix()),
			To:          to,
			Amount:      amount,
			Description: description,
		},
		From:           fromID,
		Sig:            Signature{},
		PurchaseTarget: purchaseTarget,
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxP2D_IsValid", ErrWrongTimeTX)
	}

	// 检查to id有效性和账号是否可用
	userIDValid, _ := tx.To.IsValid(gsm.opts.ChecksumLength()) // 另起一个变量userIDValid，避免阅读时被误导而已。
	if !userIDValid {
		return utils.WrapError("TxP2D_IsValid", ErrInvalidUserID)
	}
	toEcoinAccount, ok := gsm.Accounts.Map[tx.To.ID]
	if !ok {
		return utils.WrapError("TxP2D_IsValid", ErrNonexistentUserID)
	}
	if !toEcoinAccount.Available() {
		return utils.WrapError("TxP2D_IsValid", ErrUnavailableUserID)
	}

	// 检查fromID的有效性、可用性和from余额是否足够,from签名是否匹配
	userIDValid, _ = tx.From.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxP2D_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return utils.WrapError("TxP2D_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return utils.WrapError("TxP2D_IsValid", ErrUnavailableUserID)
	}
	if tx.Amount > fromEcoinAccount.Balance() {
		return utils.WrapError("TxP2D_IsValid", ErrNotSufficientBalance)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return utils.WrapError("TxP2D_IsValid", ErrInconsistentSignature)
	}

	// TODO： PurchaseTarget可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxP2D_IsValid", ErrWrongTXID)
	}

	return nil
}

/*********************************************************************************************************************
                                                    TxD2P相关
*********************************************************************************************************************/

// TxD2PArgs 新建交易函数newTxD2P()的传参
type TxD2PArgs struct {
	//	BaseArgs
	From        *Account
	FromID      UserID
	P2DBytes    []byte
	Response    []byte
	Description string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxD2PArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查p2dBytes
	p2d := &TxP2D{}
	if err = p2d.Deserialize(args.P2DBytes); err != nil {
		return ErrNotTxBytes
	}
	// 检查p2h是否在未完成交易池中
	if _, ok := gsm.UCTXP.Map[p2d.ID]; !ok {
		return ErrTXNotInUCTXP
	}
	// 检查p2h内to是否和此时的from对应，都是本机拥有的账户
	selfId, err := args.From.UserID(gsm.opts.ChecksumLength(), gsm.opts.Version())
	if err != nil {
		return err
	}
	if selfId != p2d.To {
		return ErrUnmatchedTxReceiver
	}

	// TODO: 检查 response 有效性

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}

// TxP2D 病人向下班医生发起的心电诊断交易，阶段一
type TxD2P struct {
	ID          Hash          `json:"id"`
	Time        UnixTimeStamp `json:"time"`
	From        UserID        `json:"from"`
	P2DBytes    []byte        `json:"p2dBytes"`
	Response    []byte        `json:"response"` // 比如说请求数据的密码
	Description string        `json:"description"`
	Sig         Signature     `json:"sig"`
}

// newTxD2P 新建D2P转账交易(P2D交易二段)。
func newTxD2P(args *TxD2PArgs) (tx *TxD2P, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxD2P", err)
	//}
	//
	//// 获取转账者UserID
	//fromID, err := args.From.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxD2P", err)
	//}

	// 构造tx
	tx = &TxD2P{
		ID:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:        args.FromID,
		P2DBytes:    args.P2DBytes,
		Response:    args.Response,
		Description: args.Description,
		Sig:         Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxD2P", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.From.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxD2P", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxD2P) TypeNo() uint {
	return TX_D2P
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxD2P) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxD2P_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxD2P) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxD2P) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxD2P{} 再调用
func (tx *TxD2P) Deserialize(d2pBytes []byte) (err error) {
	// 防止非空TxR2P调用该方法改变了自身内容
	tx1 := &TxD2P{}
	if tx != tx1 {
		return utils.WrapError("TxD2P_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(d2pBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxD2P_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxD2P) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxD2P{
		Id:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		From:fromID,
		P2DBytes:    p2dBytes,
		Response:    response,
		Description: description,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxD2P_IsValid", ErrWrongTimeTX)
	}

	// 检查fromID的有效性、可用性和from签名是否匹配
	userIDValid, _ := tx.From.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxD2P_IsValid", ErrInvalidUserID)
	}
	fromEcoinAccount, ok := gsm.Accounts.Map[tx.From.ID]
	if !ok {
		return utils.WrapError("TxD2P_IsValid", ErrNonexistentUserID)
	}
	if !fromEcoinAccount.Available() {
		return utils.WrapError("TxD2P_IsValid", ErrUnavailableUserID)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, fromEcoinAccount.PubKey()) {
		return utils.WrapError("TxD2P_IsValid", ErrInconsistentSignature)
	}

	// TODO： Response可用性检查。这部分交给交易双方自己做，除非达到仲裁条件，由验证节点进行仲裁才会再上层的handleTX方法中去处理

	// 检查前部交易是不是一个P2D交易，为空则错误；不为空必须是符合P2D交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.P2DBytes == nil || bytes.Compare(tx.P2DBytes, []byte{}) == 0 {
		return utils.WrapError("TxD2P_IsValid", ErrEmptySoureTX)
	}
	if bytes.Compare(tx.P2DBytes, []byte{}) != 0 {
		prevTx := &TxP2D{}
		err := prevTx.Deserialize(tx.P2DBytes)
		if err != nil {
			return utils.WrapError("TxD2P_IsValid", err)
		}
		if _, ok := gsm.UCTXP.Map[prevTx.ID]; !ok {
			return utils.WrapError("TxD2P_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxD2P_IsValid", ErrWrongTXID)
	}

	return nil
}

/*********************************************************************************************************************
                                                    TxArbitrate相关
*********************************************************************************************************************/

// TxArbitrateArgs 新建交易函数newTxArbitrate()的传参
type TxArbitrateArgs struct {
	//	BaseArgs
	Arbitrator       *Account
	ArbitratorID     UserID
	TargetTXBytes    []byte
	TargetTXComplete bool
	Description      string
}

// CheckArgsValue 检查参数值是否合规
func (args *TxArbitrateArgs) CheckArgsValue(gsm *GlobalStateMachine) (err error) {
	// 检查from? 不需要，因为就是往上给account调用的

	// 检查targetTXBytes
	var targetTX CommercialTX // 这里不需要额外检查targetTXBytes空还是不空，反正会报错
	//if err = targetTX.Deserialize(args.TargetTXBytes); err != nil { // TODO: 直接通过未赋值的接口调用方法会怎样？假如有多个结构体实现了该接口。答案：会发生空指针调用，panic
	//	return ErrNotTxBytes
	//}
	if targetTX, err = DeserializeCommercialTX(args.TargetTXBytes); err != nil {
		return ErrNotTxBytes
	}
	// 检查targetTX是否在未完成交易池中
	targetTXID, err := targetTX.Hash()
	if err != nil {
		return err
	}
	if _, ok := gsm.UCTXP.Map[targetTXID]; !ok {
		return ErrTXNotInUCTXP
	}

	// TODO: 检查 targetTXComplete 有效性。由上层去做

	// TODO: 检查 description 格式，以及代码注入？

	// 参数有效
	return nil
}

// 仲裁交易，针对商业性质交易如TxR2P的“三次僵持”提出的交易体
type TxArbitrate struct {
	ID   Hash          `json:"id"`
	Time UnixTimeStamp `json:"time"`
	// TargetTx 仲裁目标
	TargetTXBytes []byte `json:"targetTXBytes"`

	// ArbitrateResult    []byte        `json:"arbitrateResult"`

	// TargetTXComplete 目标交易是否完成，true表示完成，转账生效，否则退回
	TargetTXComplete bool `json:"targetTXComplete"`
	// Description 描述，可用来附加信息
	Description string `json:"description"`
	// Arbitrator 仲裁者
	Arbitrator UserID    `json:"arbitrator"`
	Sig        Signature `json:"sig"`
}

// newTxD2P 新建D2P转账交易(P2D交易二段)。
func newTxArbitrate(args *TxArbitrateArgs) (tx *TxArbitrate, err error) {
	//// 检验参数
	//if err = args.CheckArgsValue(); err != nil {
	//	return nil, utils.WrapError("newTxArbitrate", err)
	//}
	//
	//// 获取仲裁者UserID
	//arbitratorID, err := args.Arbitrator.UserID(args.Gsm.opts.ChecksumLength(), args.Gsm.opts.Version())
	//if err != nil {
	//	return nil, utils.WrapError("newTxArbitrate", err)
	//}

	// 构造tx
	tx = &TxArbitrate{
		ID:               Hash{},
		Time:             UnixTimeStamp(time.Now().Unix()),
		TargetTXBytes:    args.TargetTXBytes,
		TargetTXComplete: args.TargetTXComplete,
		Description:      args.Description,
		Arbitrator:       args.ArbitratorID,
		Sig:              Signature{},
	}

	// 设置Id
	id, err := tx.Hash()
	if err != nil {
		return nil, utils.WrapError("newTxArbitrate", err)
	}
	tx.ID = id
	// 设置签名
	sig, err := args.Arbitrator.Sign(id[:])
	if err != nil {
		return nil, utils.WrapError("newTxArbitrate", err)
	}
	tx.Sig = sig
	return tx, nil
}

// TypeNo 获取交易类型编号
func (tx *TxArbitrate) TypeNo() uint {
	return TX_ARBITRATE
}

// Hash 计算交易哈希值，作为交易ID
func (tx *TxArbitrate) Hash() (hash Hash, err error) {
	txCopy := *tx
	txCopy.ID, txCopy.Sig = Hash{}, Signature{}
	var res []byte
	if res, err = txCopy.Serialize(); err != nil {
		return Hash{}, utils.WrapError("TxArbitrate_Hash", err)
	}

	return sha256.Sum256(res), nil
}

// Serialize 交易序列化为字节切片
func (tx *TxArbitrate) Serialize() (result []byte, err error) {
	return utils.GobEncode(tx)
}

// String 转换为字符串，用于打印输出
func (tx *TxArbitrate) String() string {
	return utils.JsonMarshalIndent(tx)
}

// Deserialize 反序列化，必须提前 tx := &TxArbitrate{} 再调用
func (tx *TxArbitrate) Deserialize(txAtbitrateBytes []byte) (err error) {
	// 防止非空TxArbitrate调用该方法改变了自身内容
	tx1 := &TxArbitrate{}
	if tx != tx1 {
		return utils.WrapError("TxArbitrate_Deserialize", ErrDeserializeRequireEmptyReceiver)
	}

	// 反序列化
	var buf bytes.Buffer
	buf.Write(txAtbitrateBytes)
	err = gob.NewDecoder(&buf).Decode(tx)
	if err != nil {
		return utils.WrapError("TxArbitrate_Deserialize", err)
	}
	return nil
}

// IsValid 验证交易是否合乎规则
func (tx *TxArbitrate) IsValid(gsm *GlobalStateMachine) (err error) {

	/*	tx = &TxArbitrate{
		Id:          Hash{},
		Time:        UnixTimeStamp(time.Now().Unix()),
		TargetTXBytes:    targetTXBytes,
		TargetTXComplete:    targetTXComplete,
		Description: description,
		Arbitrator:arbitratorID,
		Sig:         Signature{},
	}*/

	// 检查交易时间有效性
	if tx.Time >= UnixTimeStamp(time.Now().Unix()) {
		return utils.WrapError("TxArbitrate_IsValid", ErrWrongTimeTX)
	}

	// 检查arbitratorID的有效性、可用性、角色权限和from签名是否匹配
	userIDValid, _ := tx.Arbitrator.IsValid(gsm.opts.ChecksumLength())
	if !userIDValid {
		return utils.WrapError("TxArbitrate_IsValid", ErrInvalidUserID)
	}
	arbitratorEcoinAccount, ok := gsm.Accounts.Map[tx.Arbitrator.ID]
	if !ok {
		return utils.WrapError("TxArbitrate_IsValid", ErrNonexistentUserID)
	}
	if !arbitratorEcoinAccount.Available() {
		return utils.WrapError("TxArbitrate_IsValid", ErrUnavailableUserID)
	}
	if arbitratorEcoinAccount.Role().No() >= 10 {
		return utils.WrapError("TxArbitrate_IsValid", ErrNoCoinbasePermitRole)
	}
	if !utils.VerifySignature(tx.ID[:], tx.Sig, arbitratorEcoinAccount.PubKey()) {
		return utils.WrapError("TxArbitrate_IsValid", ErrInconsistentSignature)
	}

	// TODO： 仲裁结果验证，这里不进行，丢给上层调用函数HandleTX去做。

	// 检查前部交易是不是一个未完成的商业性质交易，为空则错误；不为空必须是符合商业性质交易体且交易ID在未完成交易池中，否则认为是不合法交易
	if tx.TargetTXBytes == nil || bytes.Compare(tx.TargetTXBytes, []byte{}) == 0 {
		return utils.WrapError("TxArbitrate_IsValid", ErrEmptySoureTX)
	}
	// 其实可以把这个判断条件去掉，但是算了
	if bytes.Compare(tx.TargetTXBytes, []byte{}) != 0 {
		// 反序列化出商业交易
		var prevTx CommercialTX
		prevTx, err = DeserializeCommercialTX(tx.TargetTXBytes)
		if err != nil {
			return utils.WrapError("TxArbitrate_IsValid", err)
		}
		// 获取商业交易ID
		txId, err := prevTx.Hash()
		if err != nil {
			return utils.WrapError("TxArbitrate_IsValid", err)
		}

		if _, ok := gsm.UCTXP.Map[txId]; !ok {
			return utils.WrapError("TxArbitrate_IsValid", ErrNotUncompletedTX)
		}
	}

	// 验证交易ID是不是正确设置
	txHash, _ := tx.Hash()
	if txHash != tx.ID {
		return utils.WrapError("TxArbitrate_IsValid", ErrWrongTXID)
	}

	return nil
}
