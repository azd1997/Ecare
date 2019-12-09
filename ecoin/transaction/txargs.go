package transaction

// NewArgs 根据交易新建参数，得到各种交易类型的参数的默认状态
func NewArgs(typ uint8) Args {
	var args Args
	switch typ {
	case TX_COINBASE:
		args = &CoinbaseArgs{}

	case TX_GENERAL:
		args = &GeneralArgs{}

	}
	return args
}

// TODO