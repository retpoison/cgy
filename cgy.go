package main

import (
	"github/retpoison/cgy/config"
	"github/retpoison/cgy/tui"
)

func main() {
	var config *config.Config = config.GetConfig()
	tui.Run(config)
}
