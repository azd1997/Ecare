package singlechain

// 想要修改使用的数据库时，修改dbEngine常量和database库
// 目前只实现了badgerDB
const DBEngine = "badger"

var LastHashKey = []byte("lh")

