package transaction

import "errors"

var (

	/*本包直接使用到的错误*/

	ErrUnmatchedReceiver     = errors.New("unmatched tx receiver")
	ErrUnmatchedSender       = errors.New("unmatched tx sender")
	ErrUnknownTxType         = errors.New("unknown transaction type")
	ErrWrongArgs             = errors.New("wrong arguments for new tx")
	ErrWrongTime             = errors.New("wrong time tx")
	ErrUnknownPurchaseType   = errors.New("unknown purchase type")
	ErrWrongTxId             = errors.New("wrong tx id")
	ErrNotSufficientBalance  = errors.New("not sufficient balance")
	ErrNotTxBytes            = errors.New("not bytes of a tx")
	ErrNotCommercialTxBytes  = errors.New("not bytes of a commercial tx")
	ErrInconsistentSignature = errors.New("inconsistent signature")
	ErrNilSourceTx           = errors.New("nil source tx")
	ErrUnknownArbitrationCode = errors.New("unknown arbitration code")

	/*除上面外，外部调用本包产生的与tx直接相关的错误*/

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
