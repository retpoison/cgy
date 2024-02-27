package cgy

import (
	"fmt"
	"io"
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
	if config.LogFile == "None" {
		log.SetOutput(io.Discard)
		return
	}
	fmt.Println(config)
	f, err := os.OpenFile(config.LogFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		defer log.SetOutput(io.Discard)
		log.Printf("error opening log file: %v", err)
	}
	err = os.Truncate(config.LogFile, 0)
	if err != nil {
		log.Printf("Failed to truncate: %v", err)
	}
	log.SetOutput(f)
}

func Run() {
	defer func() {
		switch log.Writer().(type) {
		case *os.File:
			log.Writer().(*os.File).Close()
		}
	}()

	run()
}
