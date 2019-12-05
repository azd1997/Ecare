# accounts包设计

## 缘由

见 [account.go](/ecoin/doc/account.go)

## 具体设计

accounts包包含EcoinAccount与EcoinAccounts。

由于该包需要引入许多其他类型，属于比较聚合的一个包，所以后面再实现

