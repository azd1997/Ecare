package tx

import "github.com/azd1997/Ecare/ecoin/utils"

func init() {
	utils.GobRegister(&TxCoinbase{}, &TxGeneral{}, &TxR2P{}, &TxP2R{},
		&TxP2H{}, &TxH2P{}, &TxP2D{}, &TxD2P{}, &TxArbitrate{})
}
