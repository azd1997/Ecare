package ecoin

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// UnCompleteTXPool 未完成交易池
type UnCompleteTXPool struct {
	// 需要持久化的话，value不能是TX，因为不能解码到一个接口。只能存[]byte
	Map map[string]TX `json:"uctxp"`// 一旦前部交易的新进展被加进来了，前部交易被删除
}

// SaveFileWithJsonMarshal 保存到文件
func (uctxp *UnCompleteTXPool) SaveFileWithJsonMarshal(port uint) (err error) {
	file := fmt.Sprintf(UCTXP_FILEPATH_TEMP_JSON, strconv.Itoa(int(port)))
	if err = saveFileWithJsonMarshal(file, uctxp); err != nil {
		return WrapError("UnCompleteTXPool_SaveFile", err)
	}
	return nil
}

// SaveFileWithGobEncode 保存到文件
func (uctxp *UnCompleteTXPool) SaveFileWithGobEncode(port uint) (err error) {
	file := fmt.Sprintf(UCTXP_FILEPATH_TEMP_GOB, strconv.Itoa(int(port)))
	GobRegister(&TxCoinbase{}, &TxGeneral{}, &TxR2P{}, &TxP2R{}, &TxP2H{}, &TxH2P{}, &TxP2D{}, &TxD2P{}, &TxArbitrate{})
	if err = saveFileWithGobEncode(file, uctxp); err != nil {
		return WrapError("UnCompleteTXPool_SaveFile", err)
	}
	return nil
}

// LoadFileWithJsonUnmarshal 从文件加载uctxp
func (uctxp *UnCompleteTXPool) LoadFileWithJsonUnmarshal(port uint) (err error) {

	file := fmt.Sprintf(UCTXP_FILEPATH_TEMP_JSON, strconv.Itoa(int(port)))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return WrapError("UnCompleteTXPool_LoadFile", err)
	}

	var uctxp1 UnCompleteTXPool

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return WrapError("UnCompleteTXPool_LoadFile", err)
	}

	if err = json.Unmarshal(fileContent, &uctxp1); err != nil {
		return WrapError("UnCompleteTXPool_LoadFile", err)
	}

	uctxp.Map = uctxp1.Map
	return nil
}

// LoadFileWithGobDecode 从文件加载uctxp
func (uctxp *UnCompleteTXPool) LoadFileWithGobDecode(port uint) (err error) {

	file := fmt.Sprintf(UCTXP_FILEPATH_TEMP_GOB, strconv.Itoa(int(port)))
	if _, err = os.Stat(file); os.IsNotExist(err) {
		return WrapError("UnCompleteTXPool_LoadFile", err)
	}

	var uctxp1 UnCompleteTXPool

	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		return WrapError("UnCompleteTXPool_LoadFile", err)
	}

	GobRegister(&TxCoinbase{}, &TxGeneral{}, &TxR2P{}, &TxP2R{}, &TxP2H{}, &TxH2P{}, &TxP2D{}, &TxD2P{}, &TxArbitrate{})
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	if err = decoder.Decode(&uctxp1); err != nil {
		return WrapError("UnCompleteTXPool_LoadFile", err)
	}

	uctxp.Map = uctxp1.Map
	return nil
}

// HasTX 是否存在某交易
func (uctxp *UnCompleteTXPool) HasTX(key Hash) bool {
	_, ok := uctxp.Map[string(key)]
	return ok
}

// GetTX 取交易
func (uctxp *UnCompleteTXPool) GetTX(key Hash) TX {
	if v, ok := uctxp.Map[string(key)]; ok {
		return v
	}
	return nil
}