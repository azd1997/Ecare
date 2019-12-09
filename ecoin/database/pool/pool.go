/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2019/7/16 20:24
* @Description: The file is for
***********************************************************************/

package pool

import "errors"

var (
	//ErrClosed 连接池已经关闭Error
	ErrClosed = errors.New("pool is closed")
)

// Pool 基本方法
type Pool interface {
	Get() (interface{}, error)

	Put(interface{}) error

	Close(interface{}) error

	Release()

	Len() int
}
