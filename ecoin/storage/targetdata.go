package storage

import (
	"github.com/azd1997/Ecare/ecoin/common"
)

// TargetData 目标数据，在这里表示所要查找的心电数据记录
type TargetData struct {
	StartTime     common.TimeStamp 	`json:"startTime"`// =0 表示不填
	EndTime       common.TimeStamp `json:"endTime"`
	NumsOfRecords uint 	`json:"numsOfRecords"`// 若start, end均已正常设置，则该项无效

	Storage DataStorage
}

// IsValid 检查目标数据是否存在可取用
func (t *TargetData) IsOk() (err error) {
	// 1. 从TargetData解析索引

	// 2. 检查Storage是否可用

	// 3. 去查询数据是否在指定broker（Storage）中
	// TODO: 实现DataStorage接口，传入结构体指针，用以查询
	_, err = t.Storage.Query(t.StartTime, t.EndTime, t.NumsOfRecords)

	return err
}

// newTargetData 构造targetData
func newTargetData(slice []uint) *TargetData {
	l := len(slice)
	if l == 0 || l > 3 {
		return nil
	}
	switch l {
	case 1:
		return &TargetData{
			StartTime:     common.TimeStamp(slice[0]),
			EndTime:       0,
			NumsOfRecords: 0,
		}
	case 2:
		return &TargetData{
			StartTime:     common.TimeStamp(slice[0]),
			EndTime:       common.TimeStamp(slice[1]),
			NumsOfRecords: 0,
		}
	case 3:
		if slice[0] != 0 && slice[1] == 0 {
			return &TargetData{
				StartTime:     common.TimeStamp(slice[0]),
				EndTime:       0,
				NumsOfRecords: 0,
			}
		}
		if slice[0] != 0 && slice[1] != 0 {
			return &TargetData{
				StartTime:     common.TimeStamp(slice[0]),
				EndTime:       common.TimeStamp(slice[1]),
				NumsOfRecords: 0,
			}
		}
		return &TargetData{
			StartTime:     common.TimeStamp(slice[0]),
			EndTime:      common.TimeStamp(slice[1]),
			NumsOfRecords: uint(slice[2]),
		}
	}
	return nil
}
