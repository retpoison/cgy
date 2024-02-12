package main

import (
	"cgy/config"
	"cgy/tui"
	"log"
	"os"
)

func main() {
	var config *config.Config = config.GetConfig()
	f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	tui.Run(config)
}
