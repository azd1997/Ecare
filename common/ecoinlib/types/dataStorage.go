package types

// DataStorage 数据存储接口，标志数据库等
type DataStorage interface {
	Get(startTime UnixTimeStamp, endTime UnixTimeStamp, numsOfRecords uint) (records [][]byte, err error)
	Query(startTime UnixTimeStamp, endTime UnixTimeStamp, numsOfRecords uint) (ok bool, err error)
	IsOk() bool		// 判断可用与否
}

type MosquittoBroker struct {
	Addr string
}

// Query 查询数据记录是否存在
func (broker *MosquittoBroker) Query(startTime UnixTimeStamp, endTime UnixTimeStamp, numsOfRecords uint) (ok bool, err error) {
	return true, nil
}

// Get 获取数据记录
func (broker *MosquittoBroker) Get(startTime UnixTimeStamp, endTime UnixTimeStamp, numsOfRecords uint) (records [][]byte, err error) {
	return [][]byte{}, nil
}

func (broker *MosquittoBroker) IsOk() bool {
	return true
}