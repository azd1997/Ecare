package types

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/vrecan/death"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

// DbExists 判断指定路径下badger数据库是否存在
func DbExists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

// CloseDB 等待信号关闭数据库退出程序。这是为了避免程序异常退出数据库却未关闭
func CloseDB(c *Chain) {
	// 相当于注册一个中断回调时间。这样使得程序运行时被比如CTRL + C类似的操作触发程序退出，关闭数据库再关闭线程、进程
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		if err := c.Db.Close(); err != nil {
			log.Panic(err)
		}
	})
}

// openDB 根据配置打开数据库
func openDB(opts badger.Options) (db *badger.DB, err error) {
	if db, err = badger.Open(opts); err != nil {
		// 报错信息包含LOCK则retry
		if strings.Contains(err.Error(), "LOCK") {
			if db, err = retryOpenDB(opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database")
		}
		return nil, fmt.Errorf("openDB: %s", err)
	} else {
		return db, nil
	}
}

// retryOpenDB 在发现数据库有锁（未正常关闭后）移除锁并重新打开
func retryOpenDB(opts badger.Options) (db *badger.DB, err error) {
	lockPath := filepath.Join(opts.Dir, "LOCK")
	if err = os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf("retryOpenDB: removing LOCK: %s", err)
	}
	retryOpts := opts
	retryOpts.Truncate = true
	return badger.Open(retryOpts)
}
