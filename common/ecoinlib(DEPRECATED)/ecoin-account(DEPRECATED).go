package ecoinlib

//
//type ecoinAccount struct {
//	userID Address
//	balance Balance
//	role    Role
//	available  bool
//}
//
//
//// 以下方法未被使用
//
///*balance相关*/
//// 获取账户地址的余额，不作账户地址存在与否的检查，如不存在，返回默认值0，存在也有可能返回0
//func (a *ecoinAccount) getBalance(addr Address) Balance {
//	return a.balance
//}
//
//// 更新余额表，单次单个地址
//func (a *ecoinAccount) updateBalance(addr Address, newBalance Balance) {
//	a.balance = newBalance
//}
//
///*role相关*/
//// 返回role信息(不能返回指针，因为不允许更改)
//func (a *ecoinAccount) getRole(userID Address) Role {
//	return a.role
//}
//
///*available相关*/
//// 检查账户是否可用
//func (a *ecoinAccount) isAvailable(userID Address) bool {
//	return a.available
//}
