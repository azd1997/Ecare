package common

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"time"
)

// 一个心拍250个点
type DataPack struct {
	Data      [250]DataPoint
	PatientID [32]byte // 病人ID
}

// V1: MLII; V2: V5
type DataPoint struct {
	T      int64 // 时间戳
	V1, V2 int64 // 心电电压值
}

type DataPointReadable struct {
	T      time.Time // 时间
	V1, V2 float64   // 心电电压值，单位mV
}

func (d *DataPoint) Readable() (dataReadable *DataPointReadable) {
	return &DataPointReadable{
		T:  time.Unix(d.T, 0),
		V1: 0, // 将原始ADC值转化为可读的电压读数
		V2: 0,
	}
}

func (d *DataPointReadable) UnReadable() (dataPoint *DataPoint) {
	return &DataPoint{
		T:  d.T.Unix(),
		V1: 0, // 将电压读数转换回ADC读数
		V2: 0,
	}
}

func (d *DataPointReadable) Serialize() (data []byte) {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(d); err != nil {
		log.Fatal(err)
	}
	return result.Bytes()
}

func (d *DataPointReadable) String() string {
	return fmt.Sprintf("{time: %v, mlii: %v, v5: %v}", d.T, d.V1, d.V2)
}

func Deserialize(data []byte) (d *DataPointReadable) {
	reader := bytes.NewReader(data)
	decoder := gob.NewDecoder(reader)
	if err := decoder.Decode(d); err != nil {
		log.Fatal(err)
	}
	return d
}

// 打包数据  指针类型占用8B，一个DataPoint占用超16B
func PackData(dataPoints [250]*DataPoint) (pack *DataPack) {
	return
}
