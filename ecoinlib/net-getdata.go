package ecoin

import (
	"bytes"
	"encoding/gob"

	"github.com/azd1997/Ecare/ecoinlib/log"
)

// GetData 请求数据的消息体
type GetDataMsg struct {
	AddrFrom string
	Type string  // "block" "tx"
	IDs []Hash  // blockHash or txID
}

// SendGetData 向对方请求指定的多个区块或者交易数据
func (gsm *GlobalStateMachine) SendGetData(to string, typ string, ids []Hash) error {
	payload, err := GobEncode(GetDataMsg{gsm.Opts().NodeAddress().String(), typ, ids})
	if err != nil {
		return WrapError("SendGetData", err)
	}
	request := append(CmdToBytes("getdata"), payload...)
	err = gsm.SendData(to, request)
	if err != nil {
		return WrapError("SendGetData", err)
	}
	return nil
}

// HandleGetData 处理获取数据的请求
func (gsm *GlobalStateMachine) HandleGetData(request []byte) {
	// 解析request
	var buf bytes.Buffer
	var payload GetDataMsg
	buf.Write(request[COMMAD_LENGTH:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	LogErr("HandleGetData", err)

	// 处理payload
	switch payload.Type {
	case "block":
		block, err := gsm.Ledger.GetBlockByHash(payload.IDs[0])
		LogErr("HandleGetData", err)
		err = gsm.SendBlock(payload.AddrFrom, block)
		LogErr("HandleGetData", err)
	case "tx":
		tx := gsm.UBTXP.Get(payload.IDs[0])
		if tx == nil {	// 说明所要请求的交易不在本地UBTXP中
			log.Error("HandleGetData: %s", ErrTransactionNotExists)
		}
		err = gsm.SendTx(payload.AddrFrom, tx)
		LogErr("HandleGetData", err)
	default:
		log.Error("HandleGetData: %s", ErrUnknownGetDataType)
	}


}

