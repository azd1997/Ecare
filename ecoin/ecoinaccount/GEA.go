package eaccount

import "sync"

// GEA Global EAccounts 集合，外部使用，可以用来检查
var GEA  = &EAccounts{
	Map:  make(map[string]*EAccount),
	Lock: sync.RWMutex{},
}

// 关于全局变量的设计
// 全局变量都设计为带读写锁的结构。
// 一方面被其他单个包使用（在ecoin库内部使用），另一方面又被整合成一个大一统的变量，提供给外部使用。