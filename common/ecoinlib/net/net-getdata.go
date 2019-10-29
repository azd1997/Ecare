package net

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
)

type GetData struct {
	AddrFrom types.Address
	Type string  // "block" "tx"
	IDs [][]byte  // blockHash or txID
}

func (e *ecoin) SendGetData(to types.Address, typ string, ids [][]byte) error {
	payload, err := utils.GobEncode(GetData{e.Opts.NodeAddress(), typ, ids})
	if err != nil {
		return utils.WrapError("SendGetData", err)
	}
	request := append(CmdToBytes("getdata", e.Opts.CommandLength()), payload...)
	err = e.SendData(to, request)
	if err != nil {
		return utils.WrapError("SendGetData", err)
	}
	return nil
}

func (e *ecoin) HandleGetData(request []byte) {
	// 解析request
	var buf bytes.Buffer
	var payload GetData
	buf.Write(request[e.Opts.CommandLength():])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	utils.LogErr("HandleGetData", err)

	// 处理payload
	switch payload.Type {
	case "block":
		block, err := e.ledger.GetBlockByHash(payload.IDs[0])
		utils.LogErr("HandleGetData", err)
		err = e.SendBlock(payload.AddrFrom, block)
		utils.LogErr("HandleGetData", err)
	case "tx":
		txID := hex.EncodeToString(payload.IDs[0])
		tx := e.txPool[txID]
		err = e.SendTx(payload.AddrFrom, &tx)
		utils.LogErr("HandleGetData", err)
	default:
		log.Error("HandleGetData: %s", ErrUnknownGetDataType)
	}


}

