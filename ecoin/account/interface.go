package account

import "github.com/azd1997/Ecare/ecoin/common"

// IAccount 账户接口
type IAccount interface {

	// String 字符串打印，方便调试
	String() string

	// UserId 生成UserId
	UserId() (UserId, error)

	/*签名与验证*/

	// Sign 签名
	Sign(target []byte) (common.Signature, error)

	// VerifySign 验证签名
	VerifySign(target []byte, sig []byte, pubKey []byte) bool

	/*存储与读取*/

	// SaveFileWithGobEncode 使用Gob编码保存至文件
	SaveFileWithGobEncode(file string) error

	// LoadFileWithGobDecode 使用Gob解码从文件读取
	LoadFileWithGobDecode(file string) error

	// SaveFileWithJsonEncode 使用Json编码保存至文件
	SaveFileWithJsonEncode(file string) error

	// LoadFileWithJsonDecode 使用Json解码从文件读取
	LoadFileWithJsonDecode(file string) error
}


