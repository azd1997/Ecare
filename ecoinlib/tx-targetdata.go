package ecoin

// TargetData 目标数据，在这里表示所要查找的心电数据记录
type TargetData struct {
	StartTime     UnixTimeStamp 	`json:"startTime"`// =0 表示不填
	EndTime       UnixTimeStamp `json:"endTime"`
	NumsOfRecords uint 	`json:"numsOfRecords"`// 若start, end均已正常设置，则该项无效
}

// IsValid 检查目标数据是否存在可取用
func (t *TargetData) IsOk(storage DataStorage) (ok bool, err error) {
	// 1. 从TargetData解析索引

	// 2. 去查询数据是否在指定broker中
	// TODO: 实现DataStorage接口，传入结构体指针，用以查询
	ok, err = storage.Query(t.StartTime, t.EndTime, t.NumsOfRecords)

	return ok, err
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
			StartTime:     UnixTimeStamp(slice[0]),
			EndTime:       0,
			NumsOfRecords: 0,
		}
	case 2:
		return &TargetData{
			StartTime:     UnixTimeStamp(slice[0]),
			EndTime:       UnixTimeStamp(slice[1]),
			NumsOfRecords: 0,
		}
	case 3:
		if slice[0] != 0 && slice[1] == 0 {
			return &TargetData{
				StartTime:     UnixTimeStamp(slice[0]),
				EndTime:       0,
				NumsOfRecords: 0,
			}
		}
		if slice[0] != 0 && slice[1] != 0 {
			return &TargetData{
				StartTime:     UnixTimeStamp(slice[0]),
				EndTime:       UnixTimeStamp(slice[1]),
				NumsOfRecords: 0,
			}
		}
		return &TargetData{
			StartTime:     UnixTimeStamp(slice[0]),
			EndTime:       UnixTimeStamp(slice[1]),
			NumsOfRecords: uint(slice[2]),
		}
	}
	return nil
}