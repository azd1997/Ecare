package transaction

import (
	"errors"
	"github.com/azd1997/Ecare/ecoin/crypto"
)

// TxBase 基础交易。空交易，用来保证其他任何自定义交易符合
// 嵌套该空交易的结构体必须覆盖应该有的方法，这个空交易的所有方法返回报错值
// 暂未使用
type TxBase struct {}

func (t *TxBase) String() string                        { return ""}
func (t *TxBase) Serialize() (result []byte, err error) { return nil, ErrWrongTxCall}
func (t *TxBase) Deserialize(data []byte) (err error)   { return ErrWrongTxCall}
func (t *TxBase) Hash() (id crypto.Hash, err error)     { return nil, ErrWrongTxCall}
//func (t *TxBase) IsValid() (err error)                  { return ErrWrongTxCall}
func (t *TxBase) TypeNo() uint                          { return 100}
func (t *TxBase) ID() crypto.Hash                       { return nil}
func (t *TxBase) Response() *Response                   {return nil}

var ErrWrongTxCall = errors.New("wrong tx call")

func (t *TxBase) IsValid(validateFunc ValidateTxFunc) (err error)                  {
	if err = validateFunc(t); err != nil {
		return err
	}
	return nil
}