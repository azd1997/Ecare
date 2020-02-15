package eaddr

import "github.com/azd1997/Ecare/ecoin/common"

// 作恶行为相关

// BadType
const (
	BadUnknown  = iota
	BadConnFail // 注意连接失败一次扣固定分值。并视为节点掉线，不继续扣分，也不会再给它发消息

)

// GoodType
const (
	GoodUnknown = 100 + iota
	GoodBlock   // 出了一个好的区块
	GoodAddrs	// 积极回复了我的GetAddrsMsg
)

// Credit
var CreditPolicy = map[uint8]int{

	// 作恶情况
	BadUnknown:  -1,
	BadConnFail: -1,

	// 合规情况
	GoodUnknown: 1,
	GoodBlock:   1,
	GoodAddrs:1,
}


// 作恶记录，双链表节点
type BadRecord struct {
	Time       common.TimeStamp
	BadType    uint8 // 作恶类型
	Punish     int   // 惩罚，负值
	Prev, Next *BadRecord
}

