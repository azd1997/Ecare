package storage

import "github.com/azd1997/Ecare/ecoin/common"

// DataStorage 数据存储接口，标志数据库等
type DataStorage interface {
	Get(startTime common.TimeStamp, endTime common.TimeStamp, numsOfRecords uint) (records [][]byte, err error)
	Query(startTime common.TimeStamp, endTime common.TimeStamp, numsOfRecords uint) (ok bool, err error)
	IsOk() bool		// 判断可用与否
}

type MosquittoBroker struct {
	Addr string
}

// Query 查询数据记录是否存在
func (broker *MosquittoBroker) Query(startTime common.TimeStamp, endTime common.TimeStamp, numsOfRecords uint) (ok bool, err error) {
	return true, nil
}

// Get 获取数据记录
func (broker *MosquittoBroker) Get(startTime common.TimeStamp, endTime common.TimeStamp, numsOfRecords uint) (records [][]byte, err error) {
	return [][]byte{}, nil
}

func (broker *MosquittoBroker) IsOk() bool {
	return true
}
