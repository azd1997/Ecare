package tx

const (
	ECG_DIAG_AUTO = 1
	ECG_DIAG_DOCTOR = 2
)

const (
	TX_AUTO = iota
	TX_COINBASE
	TX_GENERAL
	TX_R2P
	TX_P2R
	TX_P2H
	TX_H2P
	TX_P2D
	TX_D2P
	TX_ARBITRATE
	numTXTypes
)

var TXTypes = map[uint8]string{
	TX_COINBASE:"Coinbase",
	TX_GENERAL:"General",
	TX_R2P:"R2P",
	TX_P2R:"P2R",
	TX_P2H:"P2H",
	TX_H2P:"H2P",
	TX_P2D:"P2D",
	TX_D2P:"D2P",
	TX_ARBITRATE:"Arbitrate",
}
