package ecoin

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/azd1997/Ecare/ecoinlib/log"
	"io/ioutil"
	"math/big"
	rand2 "math/rand"
	"os"
)

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

/*********************************************************************************************************************
                                                    PathExists相关
*********************************************************************************************************************/

const (
	NOT_EXISTS int = iota
	FILE_EXISTS
	DIR_EXISTS
	UNKNOWN_ERROR
)

// 文件或者文件夹存不存在
func PathExists(path string) (flag int, err error) {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return DIR_EXISTS, nil
		}
		return FILE_EXISTS, nil
	}
	if os.IsNotExist(err) {
		return NOT_EXISTS, nil
	}
	return UNKNOWN_ERROR, err
}

// 文件存不存在
func FileExists(path string) (bool, error) {
	flag, err := PathExists(path)
	switch flag {
	case FILE_EXISTS:
		return true, nil
	case UNKNOWN_ERROR:
		return false, err
	default:
		return false, nil
	}
}

// 文件夹存不存在
func DirExists(path string) (bool, error) {
	flag, err := PathExists(path)
	switch flag {
	case DIR_EXISTS:
		return true, nil
	case UNKNOWN_ERROR:
		return false, err
	default:
		return false, nil
	}
}

/*********************************************************************************************************************
                                                    SaveFile相关
*********************************************************************************************************************/

// saveFileWithGobEncode 存入文件
func saveFileWithGobEncode(filePath string, data interface{}) error {
	// 检查文件的上层目录是否存在。如果只是目录下没有这个文件，ioutil.WriteFile会创建文件。
	// 但它不会创建其上级所需的目录，所以需要检测一番
	if err := EnsureDirOfFileExists(filePath); err != nil {
		return err
	}

	dataBytes, err := GobEncode(data)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filePath, dataBytes, 0644); err != nil {
		return err
	}
	return nil
}

// saveFileWithJsonMarshal 存入文件
func saveFileWithJsonMarshal(filePath string, data interface{}) error {
	// 检查文件的上层目录是否存在。如果只是目录下没有这个文件，ioutil.WriteFile会创建文件。
	// 但它不会创建其上级所需的目录，所以需要检测一番
	if err := EnsureDirOfFileExists(filePath); err != nil {
		return err
	}

	dataBytes, err := JsonMarshalIndent(data)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filePath, dataBytes, 0644); err != nil {
		return err
	}
	return nil
}

/*********************************************************************************************************************
                                                    Hash相关
*********************************************************************************************************************/

// Hash 32B哈希。如果要修改哈希算法，只需在这里重新定义哈希的具体类型即可
// 使用[32]byte ，使用起来太不方便。
type Hash []byte

// BytesToHash 将长度为32的字节切片转换为Hash，若返回Hash{}，说明有错
func BytesToHash(data []byte) [32]byte {
	var res [32]byte
	if len(data) != cap(res) {
		return [32]byte{}	// 若返回Hash{}，说明有错
	}

	for i := 0; i < cap(res); i++ {
		res[i] = data[i]
	}

	return res
}

// RandomHash 生成随机的Hash。只是用来作一些测试
func RandomHash() Hash {
	res := make([]byte, 32)
	for i:=0; i < 32; i++ {
		res[i] = byte(uint(rand2.Intn(256)))
	}
	return res
}

// ZeroHASH 全局零哈希变量
var ZeroHASH = ZeroHash()

// ZeroHash 生成全0哈希
func ZeroHash() (zero Hash) {
	zero = make([]byte, 32)
	for i:=0; i < 32; i++ {
		zero[i] = byte(0)
	}
	return zero
}

/*********************************************************************************************************************
                                                    CmdToBytes相关
*********************************************************************************************************************/

// BytesToCmd 对命令作字节切片到字符串的转换
func BytesToCmd(cmdBytes []byte) string {
	var cmd []byte
	for _, b := range cmdBytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}
	return string(cmd)
}

// CmdToBytes 对命令作字符串到字节切片的转换
func CmdToBytes(cmd string) []byte {
	var cmdBytes []byte = make([]byte, COMMAD_LENGTH)
	for i, c := range cmd {
		cmdBytes[i] = byte(c)
	}
	return cmdBytes
}

/*********************************************************************************************************************
                                                    ecdsa-sign相关
*********************************************************************************************************************/

// Signature 签名
type Signature []byte

// VerifySignature 用公钥验证签名
func VerifySignature(target []byte, sig []byte, pubKey []byte) bool {
	// 从sig还原出r,s两个大数
	sigLen := len(sig)
	r, s := &big.Int{}, &big.Int{}
	r, s = r.SetBytes(sig[:sigLen / 2]), s.SetBytes(sig[sigLen / 2 :])		// 基于下标范围创建新切片的时候下标范围是半开区间 [start, end)

	// 还原ecdsa.PublicKey
	pubKeyLen := len(pubKey)
	x, y := &big.Int{}, &big.Int{}
	x, y = x.SetBytes(pubKey[: pubKeyLen/2]), y.SetBytes(pubKey[pubKeyLen/2 :])
	rawPubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	// 验证签名
	return ecdsa.Verify(rawPubKey, target, r, s)
}

// Sign 用私钥对目标进行签名
func Sign(target []byte, privKey *ecdsa.PrivateKey) (sig Signature, err error) {
	r, s, err := ecdsa.Sign(rand.Reader, privKey, target)
	if err != nil {
		return nil, WrapError("Sign", err)
	}
	sig = append(r.Bytes(), s.Bytes()...)
	return
}

/*********************************************************************************************************************
                                                    gob-encode相关
*********************************************************************************************************************/

// GobEncode 对目标作gob编码
func GobEncode(data interface{}) (res []byte, err error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err = enc.Encode(data); err != nil {
		return nil, WrapError("GobEncode", err)
	}
	return buf.Bytes(), nil
}

// GobRegister 批量注册实现接口的具体类型，需在编码前注册
func GobRegister(typs ...interface{}) {
	for _, arg := range typs {
		gob.Register(arg)
		//fmt.Printf("%T : %v\n", arg, arg)
	}
}

/*********************************************************************************************************************
                                                    json-marshal相关
*********************************************************************************************************************/

// TODO: go json会对[]byte作base64编码。
// 参考： https://www.cnblogs.com/fengbohello/p/4665883.html

// JsonMarshalIndent 将结构体、切片等转换为json字符串，并具备换行、缩进样式。
func JsonMarshalIndent(data interface{}) ([]byte, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")		// 带有换行和缩进的json marshal
	if err != nil {
		return nil, WrapError("JsonMarshalIndent", err)
	}
	return jsonBytes, nil
}

// JsonMarshalIndent 将结构体、切片等转换为json字符串，并具备换行、缩进样式。
func JsonMarshalIndentToString(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")		// 带有换行和缩进的json marshal
	if err != nil {
		return err.Error()
	}
	return string(jsonBytes)
}

// TODO: UnMarshal

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

/*********************************************************************************************************************
                                                    node-address相关
*********************************************************************************************************************/

func NodeIsKnown(nodeAddr string, KnownNodeAddrList []string) bool {
	for _, node := range KnownNodeAddrList {
		if node == nodeAddr {
			return true
		}
	}
	return false
}

// NodeExists 判断节点地址是否存在于列表。根据ipv4地址判断
func NodeExists(nodeAddr *Address, KnownNodeAddrList []*Address) bool {
	for _, node := range KnownNodeAddrList {
		if node.String() == nodeAddr.String() {
			return true
		}
	}
	return false
}

func NodeLocate(nodeAddr *Address, KnownNodeAddrList []*Address) int {
	for i, node := range KnownNodeAddrList {
		if node == nodeAddr {
			return i
		}
	}
	return -1
}

func MergeTwoNodeList(l1, l2 []*Address) (l3 []*Address) {
	l3 = l1
	for _, v := range l2 {
		if !NodeExists(v, l3) {
			l3 = append(l3, v)
		}
	}
	return l3
}