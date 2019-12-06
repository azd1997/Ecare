package tx

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
	//ErrBlockAlreadyExists     = errors.New("block already exists")
	ErrBlockNotExists         = errors.New("block not exists")
	ErrTransactionNotExists   = errors.New("transaction not exists")
	ErrOutOfChainRange = errors.New("index out of chain range")
	ErrWrongArguments = errors.New("wrong arguments")
	ErrUnknownTransactionType = errors.New("unknown transaction type")
	ErrWrongArgsForNewTX = errors.New("wrong arguments for new tx")
	ErrWrongArgsLengthForNewTX = errors.New("wrong arguments length for new tx")

	ErrNotOkStorage = errors.New("storage is not ok")
	ErrNonexistentTargetData = errors.New("nonexistent target data")
	ErrDeserializeRequireEmptyReceiver = errors.New("only empty receiver can call deserialize method")
	ErrNotTxBytes = errors.New("not bytes of a tx")
	ErrNotTxR2PBytes = errors.New("not bytes of a r2p tx")
	ErrNotCommercialTxBytes = errors.New("not bytes of a commercial tx")
	ErrWrongRoleUserID = errors.New("wrong role user-id")
	ErrWrongSourceTX = errors.New("wrong source tx")

	ErrWrongTimeTX = errors.New("wrong time tx")
	ErrNoCoinbasePermitRole = errors.New("no  coinbase permit role")
	ErrWrongCoinbaseReward = errors.New("wrong coinbase reward")
	ErrWrongTXID = errors.New("wrong tx id")
	ErrInconsistentSignature = errors.New("inconsistent signature")
	ErrNotUncompletedTX = errors.New("not a uncompleted tx")
	ErrUnknownPurchaseType = errors.New("unknown purchase type")
	ErrTXNotInUCTXP = errors.New("the tx is not in uctxp")
	ErrUnmatchedTxReceiver = errors.New("unmatched tx receiver")
	ErrEmptySoureTX = errors.New("empty source tx")
	ErrInvalidCoinbaseTX = errors.New("invalid coinbase tx")
	ErrInvalidArbitrateTX = errors.New("invalid arbitrate tx")
	ErrBlockContainsInvalidTX = errors.New("block contains invalid tx")
	ErrWrongTimeBlock = errors.New("wrong time block")
	ErrNotNextBlock = errors.New("not next block")
	ErrInconsistentMerkleRoot = errors.New("inconsistent merkle root hash")

	ErrLoadFileNeedEmptyReceiver = errors.New("loadfile need empty receiver")


	ErrUnavailableNode = errors.New("unavailable node address")
	ErrSendToSelf = errors.New("cannot send data to self")
	ErrUnKnownNode = errors.New("unknown node address")
	ErrUnknownInvType = errors.New("unknown inventory type")
	ErrUnknownGetDataType = errors.New("unknown getdata type")
	ErrNoValidTransaction = errors.New("no valid transaction in txPool")
	ErrBlockAlreadyExists = errors.New("this block already exists in local chain")
	ErrInvalidBlock = errors.New("invalid block for local chain")
)

// TODO: 错误太多太杂，考虑构造Error体

//type EcoinError interface {
//	error
//}
//
//type TXError struct {
//	CallFunc string
//	Erro string
//}
//
//func (e *TXError) Error() string {
//	return fmt.Sprintln(e.CallFunc + ": " + e.Erro)
//}