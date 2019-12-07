package tx

import "github.com/azd1997/Ecare/ecoin/utils"

type RespTrade struct {
	Source TX				// 这里的TX只能是R2P或者P2R
	Target TargetData
	ReplyInfo []byte
}

func (r *RespTrade) String() string {
	return utils.JsonMarshalIndentToString(r)
}

func (r *RespTrade) SourceTx() TX {
	return r.Source
}

func (r *RespTrade) TargetData() TargetData {
	return r.Target
}

func (r *RespTrade) Reply() []byte {
	return r.ReplyInfo
}
