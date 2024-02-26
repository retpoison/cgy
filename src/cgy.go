package cgy

import (
	"log"
	"os"
)

func InitConfig() {
	setDefaults()
	parseFlags()
	os.OpenFile(config.configPath, os.O_RDONLY|os.O_CREATE, 0666)
	readConfig()
}

func SetLogOutput() {
	f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}

func Run() {
	defer log.Writer().(*os.File).Close()
	run()
}
