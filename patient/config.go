package patient

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var E_Config *Config

type Config struct {
	MockEcgSource    string `json:"mockEcgSource"`
	Port             int    `json:"port"`
	BrokerIp         string `json:"brokerIp"`
	MqttClientId     string `json:"mqttClientId"`
	MqttProtocolVer  uint   `json:"mqttProtocolVer"` // 4——mqtt3.1.1; 3——mqtt3.1
	MqttUserName     string `json:"mqttUserName"`
	MqttUserPassword string `json:"mqttUserPassword"`
}

func InitConfig(configFile string) (err error) {
	var (
		content []byte
		config  Config
	)

	// 读json
	if content, err = ioutil.ReadFile(configFile); err != nil {
		return err
	}

	// json反序列化
	if err = json.Unmarshal(content, &config); err != nil {
		return err
	}

	// 赋值单例
	E_Config = &config

	log.Println(E_Config.MqttClientId)
	log.Println(E_Config.MqttUserPassword)
	log.Println(E_Config.MqttUserName)
	log.Println(E_Config.BrokerIp)
	log.Println(E_Config.MockEcgSource)
	log.Println(E_Config.Port)

	return nil
}
