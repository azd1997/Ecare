package utils

import (
	"fmt"
	"os"

	"github.com/azd1997/Ecare/ecoin/log"
)

/*********************************************************************************************************************
                                                    error相关
*********************************************************************************************************************/

// WrapError 包装error，加上调用函数前缀
func WrapError(callFunc string, err error) error {
	return fmt.Errorf("%s: %s", callFunc, err)
}

// LogErr 记录错误
func LogErr(callFunc string, err error) {
	if err != nil {
		log.Error("%s", WrapError(callFunc, err))
	}
}

// LogErrAndExit 记录错误并退出进程
func LogErrAndExit(callFunc string, err error) {
	LogErr(callFunc, err)
	os.Exit(1)
}
