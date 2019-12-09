package database

///*import (
//	"github.com/tecbot/gorocksdb"
//	"sync/atomic"
//
//	log "github.com/sirupsen/logrus"
//	"os"
//	"path"
//)
//
//// TODO: 想使用GoRocksDb，需要在电脑上先安装RocksDb，并做相关配置（这是因为GoRocksDb是RocksDb的GO API）。
//// RocksDb编译及编译产物DLL参考 https://www.cnblogs.com/crazylights/p/9950279.html
//
//type rocksDb struct {
//	db  *gorocksdb.DB
//}
//
//// 设置RocksDb选项
//var (
//	rocksOptions = gorocksdb.NewDefaultOptions()
//	readOptions = gorocksdb.NewDefaultReadOptions()
//	writeOptions = gorocksdb.NewDefaultWriteOptions()
//)
//
//// 开启Rocks数据库
//func openRocksDb(dbPath string) (Database, error) {
//	if err := os.MkdirAll(path.Dir(dbPath), os.ModePerm); err != nil { //如果dbPath对应的文件夹已存在则什么都不做，如果dbPath对应的文件已存在则返回错误
//		return nil, err
//	}
//
//	rocksOptions.SetCreateIfMissing(true)
//	rocksOptions.SetCompression(gorocksdb.NoCompression)
//	rocksOptions.SetWriteBufferSize(1000000)
//
//	db, err := gorocksdb.OpenDb(rocksOptions, dbPath)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	return &rocksDb{db}, err
//}
//
//
//
///*↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓实现Database接口↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓*/
//
////Set 为单个写操作开一个事务
//func (rd *rocksDb) Set(k, v []byte) error {
//
//	return rd.db.Put(writeOptions, k, v)
//
//}
//
////BatchSet 多个写操作使用一个事务
//func (rd *rocksDb) BatchSet(keys, values [][]byte) error {
//
//	wb := gorocksdb.NewWriteBatch()
//	defer wb.Destroy()
//	for i, key := range keys {
//		value := values[i]
//		wb.Put(key, value)
//	}
//	rd.db.Write(writeOptions, wb)
//	return nil
//}
//
////Set 为单个写操作开一个事务
//func (rd *rocksDb) SetWithTTL(k, v []byte, expireAt int64) error {
//
//	return rd.db.Put(writeOptions, k, v)
//
//}
//
////BatchSet 多个写操作使用一个事务
//func (rd *rocksDb) BatchSetWithTTL(keys, values [][]byte, expireAts []int64) error {
//
//	wb := gorocksdb.NewWriteBatch()
//	defer wb.Destroy()
//	for i, key := range keys {
//		value := values[i]
//		wb.Put(key, value)
//	}
//	rd.db.Write(writeOptions, wb)
//	return nil
//}
//
////Get 如果key不存在会返回error:Key not found
//func (rd *rocksDb) Get(k []byte) ([]byte, error) {
//
//	return rd.db.GetBytes(readOptions, k)
//}
//
////BatchGet 返回的values与传入的keys顺序保持一致。如果key不存在或读取失败则对应的value是空数组
//func (rd *rocksDb) BatchGet(keys [][]byte) ([][]byte, error) {
//
//	var slices gorocksdb.Slices
//	var err error
//	slices, err = rd.db.MultiGet(readOptions, keys...)
//	if err == nil {
//		values := make([][]byte, 0, len(slices))
//		for _, slice := range slices {
//			values = append(values, slice.Data())
//		}
//		return values, nil
//	}
//	return nil, err
//}
//
//func (rd *rocksDb) Delete(k []byte) error {
//
//	return rd.db.Delete(writeOptions, k)
//}
//
//func (rd *rocksDb) BatchDelete(keys [][]byte) error {
//
//	wb := gorocksdb.NewWriteBatch()
//	defer wb.Destroy()
//	for _, key := range keys {
//		wb.Delete(key)
//	}
//	rd.db.Write(writeOptions, wb)
//	return nil
//}
//
////Has 判断某个key是否存在
//func (rd *rocksDb) Has(k []byte) bool {
//
//	values, err := rd.db.GetBytes(readOptions, k)
//	if err == nil && len(values) > 0 {
//		return true
//	}
//	return false
//}
//
//func (rd *rocksDb) IterDB(fn func(k, v []byte) error) int64 {
//
//	var total int64
//	iter := rd.db.NewIterator(readOptions)
//	defer iter.Close()
//	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
//		//k := make([]byte, 4)
//		//copy(k, iter.Key().Data())
//		//value := iter.Value().Data()
//		//v := make([]byte, len(value))
//		//copy(v, value)
//		//fn(k, v)
//		if err := fn(iter.Key().Data(), iter.Value().Data()); err == nil {
//			atomic.AddInt64(&total, 1)
//		}
//	}
//	return atomic.LoadInt64(&total)
//}
//
////IterKey 只遍历key。key是全部存在LSM tree上的，只需要读内存，所以很快
//func (rd *rocksDb) IterKey(fn func(k []byte) error) int64 {
//
//	var total int64
//	iter := rd.db.NewIterator(readOptions)
//	defer iter.Close()
//	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
//		//k := make([]byte, 4)
//		//copy(k, iter.Key().Data())
//		//fn(k)
//		if err := fn(iter.Key().Data()); err == nil {
//			atomic.AddInt64(&total, 1)
//		}
//	}
//	return atomic.LoadInt64(&total)
//}
//
////Close 把内存中的数据flush到磁盘，同时释放文件锁
//func (rd *rocksDb) Close() error {
//
//	rd.db.Close()
//	return nil
//}
//
//
///*↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑实现Database接口↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑*/
//*/