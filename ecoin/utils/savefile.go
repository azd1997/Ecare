package utils

import "io/ioutil"

/*********************************************************************************************************************
                                                    SaveFile相关
*********************************************************************************************************************/

// SaveFileWithGobEncode 存入文件
func SaveFileWithGobEncode(filePath string, data interface{}) error {
	// 检查文件的上层目录是否存在。如果只是目录下没有这个文件，ioutil.WriteFile会创建文件。
	// 但它不会创建其上级所需的目录，所以需要检测一番
	if err := EnsureDirOfFileExists(filePath); err != nil {
		return err
	}

	dataBytes, err := GobEncode(data)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filePath, dataBytes, 0644); err != nil {
		return err
	}
	return nil
}

// SaveFileWithJsonMarshal 存入文件
func SaveFileWithJsonMarshal(filePath string, data interface{}) error {
	// 检查文件的上层目录是否存在。如果只是目录下没有这个文件，ioutil.WriteFile会创建文件。
	// 但它不会创建其上级所需的目录，所以需要检测一番
	if err := EnsureDirOfFileExists(filePath); err != nil {
		return err
	}

	dataBytes, err := JsonMarshalIndent(data)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filePath, dataBytes, 0644); err != nil {
		return err
	}
	return nil
}
