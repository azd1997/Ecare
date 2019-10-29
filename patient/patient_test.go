package patient

import (
	"bufio"
	"github.com/azd1997/Ecare/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

// 测试方法： mosquitto_sub -h test.mosquitto.org -t "TestECG" -v
func TestEcgCollect(t *testing.T) {
	// 测试ECG数据的读与上传功能

	var (
		err      error
		ecg      *os.File
		reader   *bufio.Reader
		line     string
		strSlice []string
		mlii, v5 float64
		point    common.DataPointReadable
		//pointBytes []byte
	)

	// 打开心电数据源，即虚拟心电数据来源
	if ecg, err = os.Open("./mitdb/100.csv"); err != nil {
		log.Fatalf("打开%s失败： %s", "./mitdb/100.csv", err)
	}
	defer ecg.Close()

	// "              0\t -0.145\t -0.065\n"

	reader = bufio.NewReader(ecg)

	// mqtt
	var (
		opts   *mqtt.ClientOptions
		client mqtt.Client
		token  mqtt.Token
	)

	opts = mqtt.NewClientOptions().
		AddBroker("tcp://test.mosquitto.org:1883"). // broker ip
		SetClientID("Patient100").                  // client id
		SetProtocolVersion(4)                       // mqtt 3.1.1

	client = mqtt.NewClient(opts)

	// 开启连接
	if token = client.Connect(); token.Wait() && token.Error() != nil {
		t.Error(token.Error())
	}

ReadLoop:
	// 定时30ms读一行并解析
	for range time.Tick(30 * time.Millisecond) {
		// 读一行
		if line, err = reader.ReadString('\n'); err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			// io.EOF
			goto ReadLoop
		}
		// 解析数据
		strSlice = splitECGLine(line)
		if mlii, v5, err = resolveECG(strSlice); err != nil {
			log.Println(err) // 记录何时，解析数据失败
			continue
		}
		// 包装成数据点
		point = common.DataPointReadable{
			T:  time.Now(),
			V1: mlii,
			V2: v5,
		}
		//pointBytes = point.Serialize()
		// 发布到broker
		client.Publish("TestECG", 1, true, point.String())
	}

	// 这样测试成功后，在未集成区块链货币系统之前，对patient作整体测试，见patientTest1可执行文件
	// patientTest1中mqtt部分设置包括用户名密码等设置，测试未通过
	// patientTest2中mqtt部分将用户名密码以及Will注释掉，测试仍未通过
	// 现在在各个关键部位（配置读取）加上打印信息，调试一下，patientTest3
	// 首先是在
}
