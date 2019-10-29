package device_DEPRECATED_

//
//import (
//	"fmt"
//	"github.com/azd1997/Ecare/patient/protocol"
//	"os"
//	"testing"
//	"time"
//)
//
//func TestReadFileAll(t *testing.T) {
//	file, err := os.Open("../mitdb/100.csv")
//	if err != nil {
//		t.Error(err)
//	}
//	defer file.Close()
//
//	ReadFileAll(file)
//}
//
//func TestReadFileAtLineNo(t *testing.T) {
//	file, err := os.Open("../mitdb/100.csv")
//	if err != nil {
//		t.Error(err)
//	}
//	defer file.Close()
//
//	line, nextLine, err := ReadFileAtLineNo(file, 64)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Printf("本行内容： %q， 下一行行首： %d", line, nextLine)
//}
//
//func TestDevice_ReadECGLoop(t *testing.T) {
//	d := Device{
//		Chan: make(chan protocol.DataPointReadable),
//		id:   [32]byte{},
//	}
//
//	go d.ReadECGLoop("../mitdb/100.csv")
//
//	fmt.Println(<-d.Chan)
//	time.Sleep(1*time.Minute)
//}
//
//func TestResolveECG(t *testing.T) {
//	line := "              0\t -0.145\t -0.065\n"
//	strSlice := splitECGLine(line)
//	mlii, v5, err := resolveECG(strSlice)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Printf("mlii=%f, v5=%f\n", mlii, v5)
//}
