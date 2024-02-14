package cgy

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
)

var config = Config{}

type Config struct {
	Channels   []string `json:"channels"`
	Instance   string   `json:"instance"`
	Program    string   `json:"program"`
	Options    string   `json:"options"`
	LogFile    string   `json:"logFile"`
	configPath string
}

func setDefaults() {
	config.Channels = []string{}
	config.Instance = "https://pipedapi.kavin.rocks"
	config.Program = "mpv"
	config.Options = `--keep-open=yes --force-window=yes --audio-file=%audio% --title=%title%`
	config.LogFile = "cgy.log"
	config.configPath = "cgy.json"
}

func parseFlags() {
	flag.StringVar(&config.configPath, "config",
		config.configPath, "path to the config file")
	flag.StringVar(&config.configPath, "c",
		config.configPath, "path to the config file")
	flag.StringVar(&config.Instance, "instance",
		config.Instance, "piped instance")
	flag.StringVar(&config.Instance, "i",
		config.Instance, "piped instance")
	flag.StringVar(&config.LogFile, "log",
		config.LogFile, "path to the log file")
	flag.StringVar(&config.LogFile, "l",
		config.LogFile, "path to the log file")

	flag.Parse()
}

func readConfig() {
	content, err := os.ReadFile(config.configPath)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Println(err)
	}
}

func save() {
	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile(config.configPath, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Config) set(key string, value interface{}) {
	rv := reflect.ValueOf(c).Elem()
	fv := rv.FieldByName(key)
	fv.Set(reflect.ValueOf(value))
	save()
}

func (c *Config) addChannel(channel string) {
	c.Channels = append(c.Channels, channel)
	save()
}

func (c *Config) removeChannel(channel string) {
	var chs []string
	for _, v := range c.Channels {
		if v != channel {
			chs = append(chs, v)
		}
	}
	c.Channels = chs
	save()
}
