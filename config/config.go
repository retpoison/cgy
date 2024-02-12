package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
)

type Config struct {
	Channels   []string `json:"channels"`
	Instance   string   `json:"instance"`
	Program    string   `json:"program"`
	Options    string   `json:"options"`
	LogFile    string   `json:"logFile"`
	configPath string
}

var conf = Config{}

func init() {
	setDefaults()
	parseFlags()
	os.OpenFile(conf.configPath, os.O_RDONLY|os.O_CREATE, 0666)
	readConfig()
}

func setDefaults() {
	conf.Channels = []string{}
	conf.Instance = "https://pipedapi.kavin.rocks"
	conf.Program = "mpv"
	conf.Options = `--keep-open=yes --force-window=yes --audio-file=%audio% --title=%title%`
	conf.LogFile = "cgy.log"
	conf.configPath = "cgy.json"
}

func parseFlags() {
	flag.StringVar(&conf.configPath, "config",
		conf.configPath, "path to the config file")
	flag.StringVar(&conf.configPath, "c",
		conf.configPath, "path to the config file")
	flag.StringVar(&conf.Instance, "instance",
		conf.Instance, "piped instance")
	flag.StringVar(&conf.Instance, "i",
		conf.Instance, "piped instance")
	flag.StringVar(&conf.LogFile, "log",
		conf.LogFile, "path to the log file")
	flag.StringVar(&conf.LogFile, "l",
		conf.LogFile, "path to the log file")

	flag.Parse()
}

func readConfig() {
	content, err := os.ReadFile(conf.configPath)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal(content, &conf)
	if err != nil {
		log.Println(err)
	}
}

func save() {
	content, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile(conf.configPath, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func GetConfig() *Config {
	return &conf
}

func (c *Config) Set(key string, value interface{}) {
	rv := reflect.ValueOf(c).Elem()
	fv := rv.FieldByName(key)
	fv.Set(reflect.ValueOf(value))
	save()
}

func (c *Config) AddChannel(channel string) {
	c.Channels = append(c.Channels, channel)
	save()
}

func (c *Config) RemoveChannel(channel string) {
	var chs []string
	for _, v := range c.Channels {
		if v != channel {
			chs = append(chs, v)
		}
	}
	c.Channels = chs
	save()
}
