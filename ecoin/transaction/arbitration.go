package tx

type Arbitration uint8

func (a Arbitration) IsValid() (err error) {
	as := []Arbitration{
		Fault,
		BuyerFaultLevel1, BuyerFaultLevel2, BuyerFaultLevel3,
		SellerFaultLevel1, SellerFaultLevel2, SellerFaultLevel3,
		BothFaultLevel1, BothFaultLevel2, BothFaultLevel3,
	}
	for _, a1 := range as {
		if a1 == a {return nil}
	}
	return ErrUnknownArbitrationCode
}

const (
	Fault Arbitration = iota		// 用于没匹配到既有问题的情况下作出仲裁惩处

	BuyerFaultLevel1 = 20
	BuyerFaultLevel2 = 21
	BuyerFaultLevel3 = 22

	SellerFaultLevel1 = 30
	SellerFaultLevel2 = 31
	SellerFaultLevel3 = 32

	BothFaultLevel1 = 100
	BothFaultLevel2 = 101
	BothFaultLevel3 = 102
)

var ArbitrationText = map[Arbitration]string{
	TX_COINBASE:  "Coinbase",
	TX_GENERAL:   "General",
	TX_R2P:       "R2P",
	TX_P2R:       "P2R",
	TX_P2H:       "P2H",
	TX_H2P:       "H2P",
	TX_P2D:       "P2D",
	TX_D2P:       "D2P",
	TX_ARBITRATE: "Arbitrate",
}
