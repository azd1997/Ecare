package utils

import "os"

/*********************************************************************************************************************
                                                    PathExists相关
*********************************************************************************************************************/

const (
	NOT_EXISTS int = iota
	FILE_EXISTS
	DIR_EXISTS
	UNKNOWN_ERROR
)

// 文件或者文件夹存不存在
func PathExists(path string) (flag int, err error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return DIR_EXISTS, nil
		}
		return FILE_EXISTS, nil
	}
	if os.IsNotExist(err) {
		return NOT_EXISTS, nil
	}
	return UNKNOWN_ERROR, err
}

// 文件存不存在
func FileExists(path string) (bool, error) {
	flag, err := PathExists(path)
	switch flag {
	case FILE_EXISTS:
		return true, nil
	case UNKNOWN_ERROR:
		return false, err
	default:
		return false, nil
	}
}

// 文件夹存不存在
func DirExists(path string) (bool, error) {
	flag, err := PathExists(path)
	switch flag {
	case DIR_EXISTS:
		return true, nil
	case UNKNOWN_ERROR:
		return false, err
	default:
		return false, nil
	}
}
