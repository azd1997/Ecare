package ecoinlib

import "errors"

var (
	ErrInvalidUserID         = errors.New("invalid account userID")
	ErrNonexistentUserID     = errors.New("nonexistent account userID")
	ErrUnavailableUserID     = errors.New("unavailable account userID")
	ErrNegativeTransferAmount = errors.New("negative transfer amount")
	ErrNotSufficientBalance   = errors.New("not sufficient balance")
	ErrCoinbaseTxRequireRole0 = errors.New("coinbase tx require role0 account")
	ErrChainAlreadyExists     = errors.New("blockchain already exists")
	ErrChainNotExists         = errors.New("blockchain not exists")
	ErrInvalidTransaction     = errors.New("invalid transaction")
	ErrBlockAlreadyExists     = errors.New("block already exists")
	ErrBlockNotExists         = errors.New("block not exists")
	ErrTransactionNotExists   = errors.New("transaction not exists")
	ErrOutOfChainRange = errors.New("index out of chain range")
	ErrWrongArguments = errors.New("wrong arguments")
)


