package utils

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)


/*********************************************************************************************************************
                                                    gob-encode相关
*********************************************************************************************************************/

// GobEncode 对目标作gob编码
func GobEncode(data interface{}) (res []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(data); err != nil {
		return nil, WrapError("GobEncode", err)
	}
	return buf.Bytes(), nil
}

// GobRegister 批量注册实现接口的具体类型，需在编码前注册
func GobRegister(typs ...interface{}) {
	for _, arg := range typs {
		gob.Register(arg)
		//fmt.Printf("%T : %v\n", arg, arg)
	}
}

/*********************************************************************************************************************
                                                    json-marshal相关
*********************************************************************************************************************/

// TODO: go json会对[]byte作base64编码。
// 参考： https://www.cnblogs.com/fengbohello/p/4665883.html

// JsonMarshalIndent 将结构体、切片等转换为json字符串，并具备换行、缩进样式。
func JsonMarshalIndent(data interface{}) ([]byte, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")		// 带有换行和缩进的json marshal
	if err != nil {
		return nil, WrapError("JsonMarshalIndent", err)
	}
	return jsonBytes, nil
}

// JsonMarshalIndent 将结构体、切片等转换为json字符串，并具备换行、缩进样式。
func JsonMarshalIndentToString(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")		// 带有换行和缩进的json marshal
	if err != nil {
		return err.Error()
	}
	return string(jsonBytes)
}

// TODO: UnMarshal
