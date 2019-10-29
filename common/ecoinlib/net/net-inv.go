package net

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
	"github.com/azd1997/Ecare/common/ecoinlib/types"
	"github.com/azd1997/Ecare/common/ecoinlib/utils"
)

// Inv Inventory 存证，存的是所持有的块或者交易的哈希列表
type Inv struct {
	AddrFrom string
	Type string	// "block" "tx"
	Items []types.Hash
}

func (e *ecoin) SendInv(to string, typ string, items []types.Hash) error {
	inv := Inv{e.Opts.NodeAddress(), typ, items}
	payload, err := utils.GobEncode(inv)
	if err != nil {
		return utils.WrapError("SendInv", err)
	}
	request := append(CmdToBytes("inv", e.Opts.CommandLength()), payload...)
	err = e.SendData(to, request)
	if err != nil {
		return utils.WrapError("SendInv", err)
	}
	return nil
}

// HandleInv 处理接收到的inv存证，比自己新则向对方请求接下来的具体数据，没有自己新则向对方发送自己的存证
func (e *ecoin) HandleInv(request []byte) {
	// 解析inv
	var buf bytes.Buffer
	var payload Inv
	buf.Write(request[e.Opts.CommandLength():])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Error("HandleInv", err)
	}
	log.Info("Received inventory with %d %s", len(payload.Items), payload.Type)

	// 处理inv
	switch payload.Type {
	case "block":

	case "tx":
		txID := payload.Items[0]
		if e.txPool[hex.EncodeToString(txID)].ID == nil {
			err = e.SendGetData(payload.AddrFrom, "tx", [][]byte{txID})
			if err != nil {
				log.Error("HandleInv: %s", err)
			}
		}
	default:
		log.Error("HandleInv: %s", ErrUnknownInvType)
	}
}
