package singlechain

import "errors"

var (
	ErrBlockContainsInvalidTX = errors.New("block contains bad transaction")
	ErrWrongTime = errors.New("block with wrong time")
	ErrInconsistentMerkleRoot = errors.New("inconsistent merkle root")
)
