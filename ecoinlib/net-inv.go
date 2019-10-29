package ecoin

import (
	"bytes"
	"encoding/gob"

	"github.com/azd1997/Ecare/ecoinlib/log"
)

// Inv Inventory 存证，存的是所持有的块或者交易的哈希列表
type Inv struct {
	AddrFrom string
	Type string	// "block" "tx"
	Items []Hash
}

func (gsm *GlobalStateMachine) SendInv(to string, typ string, items []Hash) error {
	inv := Inv{gsm.Opts().NodeAddress().String(), typ, items}
	payload, err := GobEncode(inv)
	if err != nil {
		return WrapError("SendInv", err)
	}
	request := append(CmdToBytes("inv"), payload...)
	err = gsm.SendData(to, request)
	if err != nil {
		return WrapError("SendInv", err)
	}
	return nil
}

// HandleInv 处理接收到的inv存证，比自己新则向对方请求接下来的具体数据，没有自己新则向对方发送自己的存证
func (gsm *GlobalStateMachine) HandleInv(request []byte) {
	// 解析inv
	var buf bytes.Buffer
	var payload Inv
	buf.Write(request[COMMAD_LENGTH:])
	dec := gob.NewDecoder(&buf)
	err := dec.Decode(&payload)
	if err != nil {
		log.Error("HandleInv: %s", err)
	}
	log.Info("Received inventory with %d %s", len(payload.Items), payload.Type)

	// 处理inv
	switch payload.Type {
	case "block":
		// TODO: 但是区块的存证往往是多个，所以本地接收到存证后需要判断本地是否有、是否有效，然后加入到本地待传区块列表。本地则从待传区块列表向别人请求区块
	case "tx":
		txID := payload.Items[0]	// TODO: 请求交易和接收交易应当是一个一个的来
		if gsm.UBTXP.Get(txID) == nil {	// 这个交易本地没有。注意接收交易时应同时检查UBTXP和TBTXP。但是请求
			err = gsm.SendGetData(payload.AddrFrom, "tx", []Hash{txID})
			if err != nil {
				log.Error("HandleInv: %s", err)
			}
		}
	default:
		log.Error("HandleInv: %s", ErrUnknownInvType)
	}
}
