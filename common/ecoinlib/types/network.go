package types

// 这个文件涵盖多机互联状态同步相关的代码函数

// 判断节点地址是否在已知集合内
//func NodeIsKnown(nodeAddr string, KnownNodeAddrList []string) bool {
//	for _, node := range KnownNodeAddrList {
//		if node == nodeAddr {
//			return true
//		}
//	}
//	return false
//}
//
//func BytesToCmd(cmdBytes []byte) string {
//	var cmd []byte
//	for _, b := range cmdBytes {
//		if b != 0x0 {
//			cmd = append(cmd, b)
//		}
//	}
//	return fmt.Sprintf("%s", cmd)
//}
//
//func CmdToBytes(cmd string, commandLength int) []byte {
//	var cmdBytes []byte = make([]byte, commandLength)
//	for i, c := range cmd {
//		cmdBytes[i] = byte(c)
//	}
//	return cmdBytes
//}
//
//func GobEncode(data interface{}) (res []byte, err error) {
//	var buf bytes.Buffer
//	enc := gob.NewEncoder(&buf)
//	if err = enc.Encode(data); err != nil {
//		return nil, fmt.Errorf("GobEncode: %s", err)
//	}
//	return buf.Bytes(), nil
//}
//
//func HandleConnection(conn net.Conn, c *Chain, commandLength int) {
//	// 读取连接得到的请求
//	var req []byte
//	var err error
//	if req, err = ioutil.ReadAll(conn); err != nil {
//		log.Fatal(fmt.Errorf("HandleConnection: %s", err))
//	}
//
//	// 从request解析command
//	command := BytesToCmd(req[:commandLength])
//	log.Printf("Received %s command\n", command)
//
//	// 对命令做处理
//	switch command {
//	case "addr":
//		HandleAddr(req, commandLength)
//	case "block":
//		HandleBlock(req, c)
//	case "inv":
//		HandleInv(req, c)
//	case "getblocks":
//		HandleGetBlocks(req, c)
//	case "getdata":
//		HandleGetData(req, c)
//	case "tx":
//		HandleTx(req, c)
//	case "version":
//		HandleVersion(req, c)
//	default:
//		log.Println("Unknown command")
//	}
//}
//
//// net-address
//type Addrs struct {
//	AddrList []string
//}
//
//// 接收到Addr请求，从request得到别人传来的Addrs，更新本地的，然后再向对方请求Blocks
//func HandleAddr(request []byte, commandLength int) {
//	// 从请求解析出对方发来的可用节点集合
//	var buf bytes.Buffer
//	var payload Addrs
//	var err error
//	buf.Write(request[commandLength:])
//	decoder := gob.NewDecoder(&buf)
//	if err = decoder.Decode(&payload); err != nil {
//		log.Fatal(fmt.Errorf("HandleAddr: GobDecode: %s", err))
//	}
//
//	// 更新已知节点
//	api.KnownNodes = append(api.KnownNodes, payload.AddrList...)
//	fmt.Printf("there are %d known nodes\n", len(api.KnownNodes))
//	RequestBlocks()
//}
//
//func SendAddr(nodeAddress string, commandLength int) {
//	nodes := Addrs{api.KnownNodes}
//	nodes.AddrList = append(nodes.AddrList, api.NodeAddress)
//	payload, _ := GobEncode(nodes)
//	request := append(CmdToBytes("addr", ),)
//}
