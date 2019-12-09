package database

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	KEY_LEN   = 30
	VALUE_LEN = 1000
)

func checksum(data []byte) string {
	checksum := md5.Sum(data)
	return hex.EncodeToString(checksum[:])
}

func Bytes(n int) []byte {
	d := make([]byte, n)
	rand.Read(d)

	return d
}

type src struct {
	Data     []byte
	Checksum string
}

func prepareData(n int) src {
	data := Bytes(n)
	checksum := md5.Sum(data)
	return src{Data: data, Checksum: hex.EncodeToString(checksum[:])}
}

func writeAndGet(db Database, parallel int) {
	var writeTime int64
	var readTime int64
	var writeCount int64
	var readCount int64
	wg := sync.WaitGroup{}
	wg.Add(parallel)
	for r := 0; r < parallel; r++ {
		go func() {
			defer wg.Done()
			EXPIRE_AT := time.Now().Add(100 * time.Minute).Unix()
			keys := [][]byte{}
			values := [][]byte{}
			validations := []string{}
			const loop = 100000
			for i := 0; i < loop; i++ {
				key := prepareData(KEY_LEN).Data
				keys = append(keys, key)
				value := prepareData(VALUE_LEN)
				values = append(values, value.Data)
				validations = append(validations, value.Checksum)
			}
			begin := time.Now()
			for i, key := range keys {
				value := values[i]
				db.SetWithTTL(key, value, EXPIRE_AT)
			}
			atomic.AddInt64(&writeTime, time.Since(begin).Nanoseconds())
			atomic.AddInt64(&writeCount, int64(len(keys)))

			begin = time.Now()
			for _, key := range keys {
				db.Get(key)
			}
			atomic.AddInt64(&readTime, time.Since(begin).Nanoseconds())
			atomic.AddInt64(&readCount, int64(len(keys)))
		}()
	}
	wg.Wait()

	fmt.Printf("write %d op/ns, read %d op/ns\n", atomic.LoadInt64(&writeTime)/atomic.LoadInt64(&writeCount), atomic.LoadInt64(&readTime)/atomic.LoadInt64(&readCount))
}

func batchWriteAndGet(db Database, parallel int) {
	var writeTime int64
	var readTime int64
	var writeCount int64
	var readCount int64

	loop := 100
	wg := sync.WaitGroup{}
	wg.Add(parallel)
	for r := 0; r < parallel; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < loop; i++ {
				EXPIRE_AT := time.Now().Add(100 * time.Minute).Unix()
				keys := [][]byte{}
				values := [][]byte{}
				expire_ats := []int64{}
				for j := 0; j < 1000; j++ {
					key := prepareData(KEY_LEN).Data
					keys = append(keys, key)
					value := prepareData(VALUE_LEN).Data
					values = append(values, value)
					expire_ats = append(expire_ats, EXPIRE_AT)
				}
				begin := time.Now()
				db.BatchSetWithTTL(keys, values, expire_ats)
				atomic.AddInt64(&writeTime, time.Since(begin).Nanoseconds())
				atomic.AddInt64(&writeCount, 1)

				begin = time.Now()
				db.BatchGet(keys)
				atomic.AddInt64(&readTime, time.Since(begin).Nanoseconds())
				atomic.AddInt64(&readCount, 1)
			}
		}()
	}
	wg.Wait()

	fmt.Printf("batch write %d op/ns, batch read %d op/ns\n", atomic.LoadInt64(&writeTime)/atomic.LoadInt64(&writeCount), atomic.LoadInt64(&readTime)/atomic.LoadInt64(&readCount))
}

func TestBadger(t *testing.T) {
	badger, _ := OpenDatabase("badger", "tmp/badgerdb")
	defer badger.Close()

	writeAndGet(badger, 1)
	batchWriteAndGet(badger, 1)

	fmt.Println("parallel test")
	writeAndGet(badger, 10)
	batchWriteAndGet(badger, 10)

	fmt.Println("please watch the memory")
	fmt.Println("badger......")
	badger.IterDB(func(k, v []byte) error {
		return nil
	})
}

func TestRocks(t *testing.T) {
	rocks, _ := OpenDatabase("rocks", "/tmp/rocks")
	writeAndGet(rocks, 1)
	batchWriteAndGet(rocks, 1)

	fmt.Println("parallel test")
	writeAndGet(rocks, 10)
	batchWriteAndGet(rocks, 10)

	fmt.Println("please watch the memory")
	fmt.Println("rocksdb......")
	rocks.IterDB(func(k, v []byte) error {
		return nil
	})
}


