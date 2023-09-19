package main

import (
	"cgy/config"
	"cgy/tui"
)

func main() {
	var config *config.Config = config.GetConfig()
	tui.Run(config)
}
