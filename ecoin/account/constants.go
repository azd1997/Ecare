package account

const (
	ACCOUNT_VERSION = byte(0x00)
	CHECKSUM_LENGTH = 4
	//SELFACCOUNTS_FILEPATH_TEMP  = "./tmp/self-accounts/self-accounts_%s.accounts"
	//SELFACCOUNT_FILEPATH_TEMP  = "./tmp/self-accounts/self-account_%s.account"
)

const (
	A = iota	// 判断角色是否是A类节点
	B 			// 判断角色是否是B类节点
	All			// 判断角色是否是All节点（0~99，AB的并集）
	Single		// 查询角色是否是指定的某一种
)

const (
	AllRole = iota				// 默认值
	Hospital 					// 1
	ResearchInstitution 		// 2
	_
	_
	_
	_
	_
	_
	_
	Patient						// 10
	Doctor						// 11
)