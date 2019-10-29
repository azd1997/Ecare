package utils

import (
	"fmt"
	"github.com/azd1997/Ecare/common/ecoinlib/log"
)

func WrapError(callFunc string, err error) error {
	return fmt.Errorf("%s: %s", callFunc, err)
}

func LogErr(callFunc string, err error) {
	if err != nil {
		log.Error("%s", WrapError(callFunc, err))
	}
}