package utils

import (
	"bytes"
	"encoding/gob"
)

func GobEncode(data interface{}) (res []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(data); err != nil {
		return nil, WrapError("GobEncode", err)
	}
	return buf.Bytes(), nil
}

func GobDecode(data []byte, payload interface{}) error {
	var buff bytes.Buffer

	buff.Write(data)
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return WrapError("GobDecode", err)
	}
	return nil
}



