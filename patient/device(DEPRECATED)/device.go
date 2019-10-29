package device_DEPRECATED_

//
//import (
//	"bufio"
//	"encoding/csv"
//	"fmt"
//	"github.com/azd1997/Ecare/patient/protocol"
//	"io"
//	"io/ioutil"
//	"log"
//	"os"
//	"strconv"
//	"strings"
//	"time"
//)
//
//// 设备。一个患者可能有多个设备
//type Device struct {
//	Chan chan protocol.DataPointReadable	// 数据由此传出
//	id 	[32]byte	// 设备ID
//}
//
//// 设备读数据方法，这里模拟心电数据的采集
//// 数据源为MIT-BIH数据库，解析之后得到的普通文本文件，该函数从心电数据源读取数据（每隔一秒读一行），读完则循环，模拟连续采集。
//// 心电数据文件内容为一行32B，从0开始，每行起始位置为0,32,64,96，......，前两行是头部不要
////func (d *Device) ReadECGOnce(reader io.Reader, line int) (point common.DataPoint, err error) {
////	ecgFile, err := os.Open("../mitdb/100.csv")
////	if err != nil {
////		return common.DataPoint{}, err
////	}
////	defer ecgFile.Close()
////
////	csvReader := csv.NewReader(ecgFile)
////
////	record, err := csvReader.Read()
////	if err != io.EOF {
////		log.Fatal(err)
////	}
////
////	return common.DataPoint{T: 1, V: 1}, nil
////}
//
//// 持续读ECG
//// 1 day = 86400s ; 1s = 10e9s ; 数据点间隔 1/360 * 10e9 = 2777777.777778ns = 2777777ns
//// 但是go定时达到100ms以下，误差比较明显，参考： https://draveness.me/golang/concurrency/golang-timer.html
//// 现在只能设置间隔为27ms。
//func (d *Device) ReadECGLoop(ecgFile string) (err error) {
//	// 打开心电数据文件
//	ecg, err := os.Open(ecgFile)
//	if err != nil {
//		return err
//	}
//	defer ecg.Close()
//
//	var reader *bufio.Reader
//	var record []string
//	var value float64
//	var data protocol.DataPointReadable
//	var
//
//ReadLoop:
//
//	fmt.Println("进入循环")
//
//	// csv.Reader
//	csvReader = csv.NewReader(ecg)
//
//	// 跳过前两行
//	for i:=0; i<2; i++ {
//		record, err = csvReader.Read()
//		if err != nil {
//			return err
//		}
//	}
//
//	fmt.Println("已跳过前头")
//
//	// time.Tick() 返回一个通道，但是这里不用接收
//	for range time.Tick(27 * time.Millisecond) {
//		// 读取一条记录
//		record, err = csvReader.Read()
//		if err != nil && err != io.EOF {
//			return err
//		}
//		// 如果是到文件结尾跳出循环，重新进行读取
//		if err == io.EOF {
//			goto ReadLoop
//		}
//
//		fmt.Println("record: ", record)
//
//		// 包装成数据点，发给接收器
//		value, err = strconv.ParseFloat(record[1], 64)
//		if err != nil {
//			value = 0.0
//		}
//		data = protocol.DataPointReadable{
//			T: time.Now(),
//			V: value,
//		}
//		d.Chan <- data
//	}
//
//	return nil
//}
//
//// 返回设备id
//func (d *Device) DeviceID() [32]byte {
//	return d.id
//}
//
//
//// (line []byte, fpos int, lineNo int, err error)
//func ReadFileAll(rd io.Reader)  {
//	// 创建带缓冲的Reader
//	r := bufio.NewReader(rd)
//
//	var line []byte
//	var err error
//	fPos := 0	// 该行起始的字节位置（在整个文件中的字节序号）
//
//	for i:=1; ;i++ {
//		// 按行读取，以'\n'byte为分隔，返回的line包含'\n'
//		line, err = r.ReadBytes('\n')
//		fmt.Printf("[line:%d, pos:%d]  %q\n", i, fPos, line)
//		if err != nil {break}
//		fPos += len(line)
//	}
//	if err != io.EOF {
//		log.Fatal(err)
//	}
//}
//
//// 要求文件每行字节数一致
//func ReadFileAtLineNo(rd io.ReadSeeker, lineStartPos int) (line []byte, nextLineStartPos int, err error) {
//	if _, err = rd.Seek(int64(lineStartPos), io.SeekStart); err != nil {
//		return nil, 0, nil
//	}
//	r := bufio.NewReader(rd)
//	line, err = r.ReadBytes('\n')
//	if err != nil {
//		if err != io.EOF {
//			log.Fatal(err)
//		}
//		//return nil, 0, err
//	}
//	return line, lineStartPos + len(line), nil
//}
//
//// 原本的csv.NewReader方法
///*// NewReader returns a new Reader that reads from r.
//func NewReader(r io.Reader) *Reader {
//	return &Reader{
//		Comma: ',',
//		r:     bufio.NewReader(r),
//	}
//}*/
//// 因为100.csv分隔符是'\n'，且无法修改（因为内部有不可见元素不能避开），除非完全重写，但不值得。
//// 所以写一个方法将100.csv改成逗号分隔的100.txt.单独使用
////func (d *Device) Csv2Txt(csvFile string) {
////	// 1. 打开文件，返回文件指针
////	file, err := os.Open(csvFile)
////	if err != nil {
////		if err != io.EOF {
////			log.Fatal(err)
////		}
////		log.Println(err)
////	}
////
////	txtFile := strings.TrimSuffix(csvFile, ".csv") + ".txt"
////	txt, err := os.Create(txtFile)
////	if err != nil {
////		log.Fatal(err)
////	}
////
////	reader := bufio.NewReader(file)
////
////	var line, newLine string
////
////	for {
////		line, err = reader.ReadString('\n')
////		if err != nil {
////			break
////		}
////		//strings
////	}
////
////	if err != io.EOF {
////		log.Fatal(err)
////	}
////
////	txt = txt
////	line = line
////	newLine = newLine
////
////}
//
//
//func resolveECG(strSlice []string) (mlii, v5 float64, err error) {
//	// "              0\t -0.145\t -0.065\n"
//
//	strSlice[1] = strings.ReplaceAll(strSlice[1], " ", string(""))
//	strSlice[2] = strings.ReplaceAll(strSlice[2], " ", string(""))
//
//	mlii, err = strconv.ParseFloat(strSlice[1], 64)
//	if err != nil {
//		return 0.0, 0.0, err
//	}
//	v5, err = strconv.ParseFloat(strSlice[2], 64)
//	if err != nil {
//		return 0.0, 0.0, err
//	}
//	return mlii, v5, nil
//
//}
//
//func splitECGLine(line string) (strSlice []string) {
//	line = strings.TrimSuffix(line, string('\n'))
//	return strings.Split(line, string('\t'))
//}
//
////func deleteHeadOfECGCsv(csvFile string) {
////	// 1. 打开文件，返回文件指针
////	file, err := os.Open(csvFile)
////	if err != nil {
////		if err != io.EOF {
////			log.Fatal(err)
////		}
////		log.Println(err)
////	}
////
////	file.Trun
////
////	// 删除前两行
////	rw := bufio.NewReadWriter(bufio.NewReader(file), bufio.NewWriter(file))
////	bs, err := rw.Peek(64)
////	rw.
////}
//// 由于文件操作只能覆写，其存储是连续的，所以与其写程序去删除心电数据文件头部还不如手动去删
