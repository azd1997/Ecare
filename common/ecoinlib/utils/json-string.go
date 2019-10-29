package utils

import (
	"encoding/json"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
)

// JsonMarshalIndent 将结构体、切片等转换为json字符串，并具备换行、缩进样式。
func JsonMarshalIndent(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")		// 带有换行和缩进的json marshal
	if err != nil {
		log.Error("JsonMarshalIndent: %s", err)
	}
	return string(jsonBytes)
}

// TODO: UnMarshal
