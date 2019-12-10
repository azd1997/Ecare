package singlechain

import "errors"

var (
	ErrBlockContainsInvalidTX = errors.New("block contains bad transaction")
	ErrWrongTime = errors.New("block with wrong time")
	ErrInconsistentMerkleRoot = errors.New("inconsistent merkle root")

	ErrChainAlreadyExists = errors.New("chain already exists")
	ErrChainNotExists = errors.New("chain not exists")
	ErrInvalidTX = errors.New("invalid transaction")
	ErrBlockAlreadyExists = errors.New("block already exists in chain")
	ErrBlockNotExists = errors.New("block not exists in chain")
	ErrIdOutOfChainRange = errors.New("index id out of chain range")
	ErrWrongArguments = errors.New("wrong input arguments")
	ErrTransactionNotExists = errors.New("transaction not exists in chain")
)
