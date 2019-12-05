package eaccount

import "github.com/azd1997/Ecare/ecoin/account"

// IEcoinAccounts 公开信息集合接口，使用时可不关心底层实现
// 现在调用方只需要知道这些接口函数
type IEcoinAccounts interface {

	/*增删改查*/

	// AddEAccount 向账户表添加新账户
	AddEAccount(new *EAccount) error

	// DelEAccount 根据用户Id删除EAccount
	DelEAccount(id string) error

	// UpdateEAccount 更新某一个账户的状态。为了追求通用，直接传入一个保持该Id不变的新EAccount来进行更新
	UpdateEAccount(new *EAccount) error

	// GetEAccount 根据用户id查看公开账户
	GetEAccount(id string) *EAccount

	// GetAllUserID 获取所有账户的对应的UserID
	GetAllUserID() []account.UserId

	/*存储与加载*/

	// SaveFileWithGobEncode 将内存维护的自己账户表写入本地指定路径下的文件
	SaveFileWithGobEncode(file string) error

	// SaveFileWithJsonMarshal 将内存维护的自己账户表写入本地指定路径下的文件
	SaveFileWithJsonMarshal(file string) error

	// LoadFileWithGobDecode 从本地文件中读取自己账户表（用于加载）
	LoadFileWithGobDecode(file string) error

	// LoadFileWithJsonUnmarshal 从本地文件中读取自己账户表（用于加载）
	LoadFileWithJsonUnmarshal(file string) error


}