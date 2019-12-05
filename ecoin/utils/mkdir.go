package utils

import "os"

/*********************************************************************************************************************
                                                    Mkdir相关
*********************************************************************************************************************/

// MkdirAll 创建目录，即便中间断层。
func MkdirAll(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// OpenFileAll 打开文件，文件不存在则创建
// 这个方法其实不需要，因为ioutil.Writefile里面用了这个os.OpenFile
func OpenFileAll(path string) (file *os.File, err error) {
	return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0766)
}

// ExtractDirFromFilePath 提取文件路径的目录。例如"./tmp/1.txt" -> "./tmp"
func ExtractDirFromFilePath(filePath string) (dir string) {
	var index int
	strBytes := []byte(filePath)
	for i:=len(strBytes)-1; i>=0; i-- {
		if strBytes[i] == '/' {
			index = i // index处为截断处，且不含index所指项
			break
		}
	}
	dir = string(strBytes[:index])
	return
}

// EnsureDirOfFileExists 确保文件的上级目录存在，不存在则创建
func EnsureDirOfFileExists(filePath string) error {
	// 检查文件的上层目录是否存在。如果只是目录下没有这个文件，ioutil.WriteFile会创建文件。
	// 但它不会创建其上级所需的目录，所以需要检测一番
	dir := ExtractDirFromFilePath(filePath)
	exists, err := DirExists(dir)
	if err != nil {
		return err
	}
	if !exists {	// 如果路径不存在就创建
		if err = MkdirAll(dir); err != nil {
			return err
		}
	}
	return nil
}
