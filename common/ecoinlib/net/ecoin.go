package net

import (
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
)

var EcoinWorld ecoin

func init() {
	EcoinWorld = ecoin{}
}

// ecoin Ecoin核心结构体
type ecoin struct {
	//accounts map[types.UserID]types.Account
	//ledger types.Chain
	//logger log.Logger
	//Opts types.Option
	//addrList AddrList // TODO: 写保护
	//blockInTransit []byte // TODO: 写保护
	//txPool map[string]types.Transaction		// TODO: 写保护
	types.GlobalStateMachine
	logger *log.Logger
	blockInTransit []types.Hash
}

func (e *ecoin) DefaultOption() *types.Option {
	defaultOpts := types.DefaultOption()
	e.SetOpts(defaultOpts)
	return e.Opts()
}



