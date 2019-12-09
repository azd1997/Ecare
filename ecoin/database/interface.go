package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/azd1997/Ecare/ecoin/utils"
)

// TODO: 制定合理的日志记录系统。 参考https://blog.csdn.net/wslyk606/article/details/81670713
// BadgerDB与RocksDB性能对比：https://www.cnblogs.com/zhangchaoyang/articles/9427675.html

/*database.go用来隐藏和选择不同的数据库*/



type Database interface {
	Set(k, v []byte) error
	BatchSet(keys, values [][]byte) error
	SetWithTTL(k, v []byte, expireAt int64) error
	BatchSetWithTTL(keys, values [][]byte, expireAts []int64) error
	Get(k []byte) ([]byte, error)
	BatchGet(keys [][]byte) ([][]byte, error)
	Delete(k []byte) error
	BatchDelete(keys [][]byte) error
	Has(k []byte) bool
	IterDB(fn func(k, v []byte) error) int64
	IterKey(fn func(k []byte) error) int64
	Close() error
}



var dbOpenFunction = map[string]func(path string) (Database, error){
	"badger": openBadgerDb,
	//"bolt":   OpenBoltDb,
	//"couch":  OpenCouchDb,
	//"level":  OpenLevelDb,
	//"rocks":  OpenRocksDb,
	//"sqlite": OpenSqlite,
}


// 根据数据库引擎类型和数据库路径开启或创建新的数据库
func OpenDatabase(dbEngine string, path string) (Database, error) {
	if fc, exists := dbOpenFunction[dbEngine]; exists {
		return fc(path)
	} else {
		return nil, fmt.Errorf("unsupported storage engine: %v", dbEngine)
	}
}


// 检查数据库是否存在
// TODO:现在这个数据库存在检测只是badgerdb适用，后期再扩展
func DbExists(dbEngine string, path string) bool {

	// 当读取文件信息无错且文件非路径名时，说明我们确实找到了这个MANIFEST，数据库确实存在
	// 1.先检查数据库存放的路径存不存在，存在则还要求必须为Dir
	exists, err := utils.DirExists(path)
	if err != nil {
		log.Fatal("检查路径是否存在时发生未知错误：", err)
	}
	if !exists {
		log.Fatal(path, ": 数据库路径不存在")
	}

	// 确保了数据库指定的路径存在后，检查数据库MANIFEST文件是否存在
	exists, err = utils.FileExists(path + "/MANIFEST")
	if err != nil {
		log.Fatal("检查数据库MANIFEST文件是否存在时发生未知错误：", err)
	}
	if !exists {
		// 这属于正常情况，程序应返回结果，让调用程序继续向下执行
		//log.Println(path + "/MANIFEST", ": 数据库文件不存在")
		return false
	}

	return true
}

// 打开数据库，不成则retry一次
func OpenDatabaseWithRetry(dbEngine string, path string) (Database, error) {

	var (
		db Database
		err error
	)
	if db, err = OpenDatabase(dbEngine, path); err != nil {
		//报错信息包含“LOCK”，则retry
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dbEngine, path); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	}

	return db, nil
}

// retry
func retry(dbEngine string, path string) (Database, error) {

	var (
		lockPath string
		db Database
		err error
	)

	lockPath = filepath.Join(path, "LOCK")
	if err = os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	if db, err = OpenDatabase(dbEngine, path); err != nil {
		return nil, err
	}

	return db, nil
}


// TODO:所有实现Database接口的数据库引擎的配置文件应做配置文件分离，实现外部更改配置文件来修改程序内部的配置文件

