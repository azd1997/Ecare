package api

import (
	"github.com/azd1997/Ecare/common/ecoinlib"
	"strconv"
	"time"
)

// 外部应用要怎么使用？
// opts := ecoin.DefaultOption().SetUserID(MyID)
// 这边根据ID应自动解析出role并只提供能用的api(不会)
// 那么近一步
// 构建cli程序，要用到这里各个对外Api,
// 也就是说这里提供的api都需要使用opts参数

// 此外，提供一个匿名的全局变量表示ecoin系统，之所以匿名是不想让外部多创建
// 或者不可导出结构体和导出变量

//var Ecoin struct{
//	accounts map[ecoinlib.UserID]*ecoinlib.Account	// 只是用于内部验证，所以不导出，只在内部逻辑进行增删改查
//	ledger	ecoinlib.Chain	// 对于外部来讲也不需要直接操作ledger
//	logger ecoinlib.Logger	// 外部不直接操作
//	Opts ecoinlib.Option  // 外部可见
//} = struct {
//	accounts map[ecoinlib.UserID]*ecoinlib.Account
//	ledger   ecoinlib.Chain
//	logger   ecoinlib.Logger
//	Opts     ecoinlib.Option
//}{
//	accounts: make(map[ecoinlib.UserID]*ecoinlib.Account),
//	ledger: *chain(),
//	logger: *logger(),
//	Opts: DefaultOption(),
//}

var Ecoin = ecoin{
	accounts: make(map[ecoinlib.UserID]*ecoinlib.Account),
	ledger:   ecoinlib.Chain{},
	logger:   ecoinlib.Logger{},
	Opts:     ecoinlib.Option{},
}

type ecoin struct {
	accounts map[ecoinlib.UserID]*ecoinlib.Account
	ledger   ecoinlib.Chain
	logger   ecoinlib.Logger
	Opts     ecoinlib.Option
}


func (e *ecoin) DefaultOption() *ecoinlib.Option {
	e.Opts = *ecoinlib.DefaultOption()
	return &e.Opts
}

func chain() *ecoinlib.Chain {
	return &ecoinlib.Chain{}
}

func logger() *ecoinlib.Logger {
	return &ecoinlib.Logger{}
}

func (e *ecoin) TestMethod() {
	ecoinlib.Error("%s", e.Opts.UserID())
	ecoinlib.Info("%s", e.Opts.UserID())
	tx := ecoinlib.Transaction{
		ID:          []byte("aaaaaaaaa"),
		CreateTime:  time.Now().Unix(),
		SubmitTime:  time.Now().Unix(),
		PassTime:    0,
		From:        "eiger",
		To:          "zr",
		Amount:      50,
		Description: "eiger to zr",
		Signature:   nil,
	}
	ecoinlib.Info("\n" + tx.String())

	b1 := ecoinlib.Block{
		BlockHeader: ecoinlib.BlockHeader{
			Id:1,
			CreateTime:time.Now().Unix() - 100,
			CommitTime:time.Now().Unix(),
			PrevHash:[]byte("PrevHashTest"),
			Hash:[]byte("HashTest"),
			CreateBy:"eiger",
		},
		BlockBody:   ecoinlib.BlockBody{},
	}
	ecoinlib.Info("\n" + b1.String())

	b2 := ecoinlib.Block{
		BlockHeader: ecoinlib.BlockHeader{
			Id:2,
			CreateTime:time.Now().Unix() - 100,
			CommitTime:time.Now().Unix(),
			PrevHash:[]byte("PrevHashTest2"),
			Hash:[]byte("HashTest2"),
			CreateBy:"eiger",
		},
		BlockBody:   ecoinlib.BlockBody{},
	}
	b3 := ecoinlib.Block{
		BlockHeader: ecoinlib.BlockHeader{
			Id:3,
			CreateTime:time.Now().Unix() - 100,
			CommitTime:time.Now().Unix(),
			PrevHash:[]byte("PrevHashTest3"),
			Hash:[]byte("HashTest3"),
			CreateBy:"eiger",
		},
		BlockBody:   ecoinlib.BlockBody{},
	}
	c, _ := ecoinlib.InitChain(e.Opts.UserID(), strconv.Itoa(e.Opts.Port()), e.Opts.DbPathTemp(), e.Opts.GenesisMsg)
	e.ledger = *c
	e.ledger.AddBlock(&b1)
	e.ledger.AddBlock(&b2)
	e.ledger.AddBlock(&b3)
	e.ledger.PrintBlockHeaders(1,2)
}

/*
 * ecoinlib对外应提供这些：
 *
 * 1. var EcoinWorld 全局状态机； errors
 * 2. 各类常量
 * 3. 函数：
 * 		3.1 InitialChain()  仅允许role0创建
 * 		3.2 ContinueChain()
* 		3.2 ContinueChain()

* 		3.2 StartNode()
 *
 *
 *
*/

// 全局状态机：区块链账本、全局User信息（对于转发节点需要维护）、

var (
	NodeAddress    string
	AccountAddress string
	KnownNodes     []string
	BlockInTransit [][]byte
	MemoryPool     = make(map[string]ecoinlib.Transaction)
)

func init() {
	KnownNodes = append(KnownNodes, NodeAddress)
}

//// 1. StartNode 除了Role0账户第一次创建区块链外，所有区块链
//func StartNode(opts ecoinlib.Option) {
//	// 全局变量赋值
//	NodeAddress = fmt.Sprintf("localhost:%s", nodeId)
//	AccountAddress = accountAddress
//
//	// 本地节点开始监听
//	var ln net.Listener
//	var err error
//	if ln, err = net.Listen(PROTOCOL, NodeAddress); err != nil {
//		log.Fatal(fmt.Errorf("StartNode: TcpListen: %s", err))
//	}
//	defer ln.Close()
//
//	// 从数据库继续区块链
//	var chain *ecoinlib.Chain
//	if chain, err = ecoinlib.ContinueChain(nodeId); err != nil {
//		log.Fatal(fmt.Errorf("StartNode: %s", err))
//	}
//	defer chain.Db.Close()
//
//	// 如果本地节点不是已知节点集第一个，那就向已知节点集第一个发送版本version
//	if NodeAddress != KnownNodes[0] {
//		SendVersion(KnownNodes[0], chain)
//	}
//
//	// 循环，接受请求并处理
//	var conn net.Conn
//	for {
//		if conn, err = ln.Accept(); err != nil {
//			log.Fatal(fmt.Errorf("StartNode: TcpAccept: %s", err))
//		}
//		go HandleConnection(conn, chain)
//	}
//}

// 打印类

// 2. PrintBlockHeaders 打印整个区块链
func (e *ecoin) PrintBlockHeaders(start, end int) {
	if err := e.ledger.PrintBlockHeaders(start, end); err != nil {
		ecoinlib.Error("Ecoin: %s", err)
	}
}

// 3. PrintTransactionOfBlock int可正可负
func (e *ecoin) PrintTransactionOfBlock(id int) {
	b, err := e.ledger.GetBlockById(id)
	if err != nil {
		ecoinlib.Error("Ecoin: %s", err)
	}
	for _, tx := range b.Transactions {
		ecoinlib.Info("%s", tx.String())
	}
}

// 3. PrintTransactionsOfUserID
func (e *ecoin) PrintTransactionSOfUserID(userID ecoinlib.UserID) {

}

func (e *ecoin) PrintBlockHeadersOfCreator(creator ecoinlib.UserID) {

}

// 交易类
// 四类交易：
// 直接转账型（最简单，由转账者直接构建交易）
// R向P发起交易型（R构建交易阶段1，包含交易金额，交易目标、交易内容，这是一个交易，承认（检查交易数额合法，交易内容存在）后公示，P在运行期间内同步区块时检查与自身有关的交易，
// 当扫描到这类交易时若承认则构建阶段2交易（包括交易来源，交易金额，交易内容，交易结果（所请求内容的解锁权限）），广播出去被承认（检查阶段1交易的合法性，阶段2交易新增内容的合法性）后）
// 这里有个问题：很有可能R接收到阶段2交易时交易还未被承认。这里怎么解决：
// 因为都会维护转发节点集合C，所以必须得到所有健康可用的转发节点c1的承认，且可用节点数必须是C的2/3以上。
// P构建交易时使用所有转发节点公钥进行加密，

// 网络同步策略
// Role0 -> Hospital ->

// 这类嵌套交易怎么构建
func (e *ecoin) NewTransaction(to ecoinlib.UserID, amount ecoinlib.Balance, msg string) {

}

// network类
type version struct {
	version int
	bestHeight int
	addrFrom string
}

