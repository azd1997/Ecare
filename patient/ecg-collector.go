package patient

import (
	"bufio"
	"github.com/azd1997/Ecare/common"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type EcgCollector struct {
	// nothing
}

var E_EcgCollector *EcgCollector

func InitEcgCollector() (err error) {
	log.Println(E_Config.MqttClientId)
	log.Println(E_Config.MqttUserPassword)
	log.Println(E_Config.MqttUserName)
	log.Println(E_Config.BrokerIp)
	log.Println(E_Config.MockEcgSource)
	log.Println(E_Config.Port)
	E_EcgCollector = &EcgCollector{}
	E_EcgCollector.CollectEcgLoop()
	return
}

func (c *EcgCollector) CollectEcgLoop() {

	var (
		err        error
		ecg        *os.File
		reader     *bufio.Reader
		line       string
		strSlice   []string
		mlii, v5   float64
		point      common.DataPointReadable
		pointBytes []byte
	)

	// 打开心电数据源，即虚拟心电数据来源
	if ecg, err = os.Open(E_Config.MockEcgSource); err != nil {
		log.Fatalf("打开%s失败： %s", E_Config.MockEcgSource, err)
	}
	defer ecg.Close()

	// "              0\t -0.145\t -0.065\n"

	reader = bufio.NewReader(ecg)

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
		pointBytes = point.Serialize()
		// 发布到broker
		if err = E_MqttClient.Pub(E_Config.MqttClientId+"/ecg", string(pointBytes)); err != nil {
			log.Println(err)
		}
	}
}

func resolveECG(strSlice []string) (mlii, v5 float64, err error) {
	// "              0\t -0.145\t -0.065\n"

	strSlice[1] = strings.ReplaceAll(strSlice[1], " ", string(""))
	strSlice[2] = strings.ReplaceAll(strSlice[2], " ", string(""))

	mlii, err = strconv.ParseFloat(strSlice[1], 64)
	if err != nil {
		return 0.0, 0.0, err
	}
	v5, err = strconv.ParseFloat(strSlice[2], 64)
	if err != nil {
		return 0.0, 0.0, err
	}
	return mlii, v5, nil

}

func splitECGLine(line string) (strSlice []string) {
	// "              0\t -0.145\t -0.065\n"
	line = strings.TrimSuffix(line, string('\n'))
	return strings.Split(line, string('\t'))
}
