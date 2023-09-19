package config

import (
	"flag"
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Instance   string
	Channels   []string
	Program    string
	Options    string
	configPath string
	viper      *viper.Viper
}

var conf = Config{}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.StringVar(&conf.configPath, "config", "", "path to the config file")
	flag.StringVar(&conf.configPath, "c", "", "path to the config file")
	flag.StringVar(&conf.Instance, "instance", "", "piped instance")
	flag.StringVar(&conf.Instance, "i", "", "piped instance")
	flag.Parse()

	os.OpenFile(conf.configPath+"cgy.json", os.O_RDONLY|os.O_CREATE, 0666)

	initViper(&conf)
}

func initViper(c *Config) {
	c.viper = viper.New()
	c.viper.SetConfigFile("cgy.json")
	if c.configPath == "" {
		c.viper.AddConfigPath(".")
	} else {
		c.viper.AddConfigPath(c.configPath)
	}
	var err = c.viper.ReadInConfig()
	if err != nil {
		log.Println(err)
	}

	if !(c.viper.IsSet("program")) {
		c.Set("program", "mpv")
	}
	if !(c.viper.IsSet("options")) {
		c.Set("options", "--force-window=yes --title=%title%")
	}
	if !(c.viper.IsSet("instance")) {
		c.Set("options", "https://pipedapi.kavin.rocks")
	}

	var initConfig = func() {
		if c.Instance == "" {
			c.Instance = c.viper.GetString("instance")
		}
		c.Channels = c.viper.GetStringSlice("channels")
		c.Program = c.viper.GetString("program")
		c.Options = c.viper.GetString("options")
	}

	initConfig()
}

func GetConfig() *Config {
	return &conf
}

func (c *Config) Set(key, value string) {
	c.viper.Set(key, value)
	c.viper.WriteConfig()
}

func (c *Config) AddChannel(channel string) {
	c.Channels = append(c.Channels, channel)
	c.viper.Set("channels", c.Channels)
	c.viper.WriteConfig()
}

func (c *Config) RemoveChannel(channel string) {
	var chs []string
	for _, v := range c.Channels {
		if v != channel {
			chs = append(chs, v)
		}
	}
	c.Channels = chs
	c.viper.Set("channels", c.Channels)
	c.viper.WriteConfig()
}
