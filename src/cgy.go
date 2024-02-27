package cgy

import (
	"log"
	"os"
)

func InitConfig() {
	setDefaults()
	parseFlags()
	_, err := os.OpenFile(config.configPath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
	}
	readConfig()
}

func SetLogOutput() {
	f, err := os.OpenFile(config.LogFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	log.SetOutput(f)
}

func Run() {
	defer log.Writer().(*os.File).Close()
	run()
}
