package ecoin

const (
	CHECKSUM_LENGTH = 4
	COMMAD_LENGTH   = 12
	NODE_VERSION    = 1
	ACCOUNT_VERSION = byte(0x00)

	CHAIN_DBPATH_TEMP = "./tmp/blocks/blocks_%s"
	COINBASE_MSG_TEMP = "[%s] | %s mined the %d block" // arg1: time; args2: userid; arg3: blockid;

	PROTOCOL     = "tcp"
	MAX_CONN_NUM = 100

	SELFACCOUNTS_FILEPATH_TEMP  = "./tmp/self-accounts/self-accounts_%s.accounts"
	SELFACCOUNT_FILEPATH_TEMP  = "./tmp/self-accounts/self-account_%s.account"
	ECOINACCOUNTS_FILEPATH_TEMP = "./tmp/ecoin-accounts/ecoin-accounts_%s.json"

	UCTXP_FILEPATH_TEMP_GOB = "./tmp/uctxp/uctxp_%s.gob"
	UCTXP_FILEPATH_TEMP_JSON = "./tmp/uctxp/uctxp_%s.json"
	ADDRS_FILEPATH_TEMP = "./tmp/addrs/addrs_%s.json"

	PERIOD = 30 // 单位 s 每30s出一个块

)
