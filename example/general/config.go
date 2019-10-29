package main

import (
	"encoding/json"
	ecoin "github.com/azd1997/Ecare/ecoinlib"
	"io/ioutil"
	"log"
	"os"
)

//var E_config *Config
var E_Opts *ecoin.Option

type Config struct {
	Ipv4 string	`json:"ipv4"`		// TODO： 其实不应该有ipv4，应该由程序或许本机网卡IP，只设置监听端口一项
	Port uint `json:"port"`
	SeedNode string `json:"seedNode"` // 种子节点
	BrokerAddr string `json:"brokerAddr"` // 数据存储Broker服务器
	// 其他的配置都根据这二者去找到对应文件得到

	RegisterInfo RegisterInfo `json:"registerInfo"`
}

type RegisterInfo struct {
	Name string	`json:"name"`
	Phone string 	`json:"phone"`
	Institution string 	`json:"institution"`
}

func initConfig() {
	var (
		content []byte
		config Config
		err error
	)

	//log.Println("cfgFile: ", cfgFile)

	// 读取配置文件，得到[]byte内容
	if content, err = ioutil.ReadFile(cfgFile); err != nil {
		log.Println("读取配置失败: ", err)
		os.Exit(1)
	}

	// 反序列化
	if err = json.Unmarshal(content, &config); err != nil {
		log.Println("读取配置失败: ", err)
		os.Exit(1)
	}

	// 赋值单例
	//E_config = &config		// 其实这里可以不要了，但是先留着，免得有用
	E_Opts = ecoin.DefaultOption().SetIpv4(config.Ipv4).SetPort(int(config.Port)).
		SetSeedNode(config.SeedNode).SetBrokerAddr(config.BrokerAddr).
		SetName(config.RegisterInfo.Name).SetPhone(config.RegisterInfo.Phone).SetInstitution(config.RegisterInfo.Institution)


	//log.Print(E_config)

	return
}