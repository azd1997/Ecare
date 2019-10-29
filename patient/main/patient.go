package main

import (
	"flag"
	"github.com/azd1997/Ecare/common"
	"github.com/azd1997/Ecare/patient"
	"log"
	"runtime"
)

var configFile string

func initArgs() {
	// patient -config ./patient.json
	flag.StringVar(&configFile, "config", common.P_CONFIG_DEFAULT_DIR, "请提供配置文件路径")
	flag.Parse()
}

func initEnv() {
	// 线程数 = CPU数
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var err error

	initArgs()
	initEnv()

	if err = patient.InitConfig(configFile); err != nil {
		goto ERR
	}

	if err = patient.InitMqttClient(); err != nil {
		goto ERR
	}

	if err = patient.InitEcgCollector(); err != nil {
		goto ERR
	}

ERR:
	log.Fatal(err)
}
