package main

import (
	cgy "cgy/src"
)

func main() {
	cgy.InitConfig()
	cgy.SetLogOutput()
	cgy.InitTui()
	cgy.Run()
}
