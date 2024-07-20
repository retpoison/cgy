package cgy

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
)

var config = Config{}

type Config struct {
	Channels   []string `json:"channels"`
	Instance   string   `json:"instance"`
	Program    string   `json:"program"`
	Options    []string `json:"options"`
	LogFile    string   `json:"logFile"`
	clean      bool
	configPath string
}

func setDefaults() {
	config.Channels = []string{}
	config.Instance = "https://pipedapi.kavin.rocks"
	config.Program = "mpv"
	config.Options = []string{"--keep-open=yes", "--force-window=yes",
		"--audio-file=%audio%", "--title=%title%",
		"--external-file=%thumbnail%", "--vid=1"}
	config.clean = false

	var home string = os.Getenv("HOME")
	configFolder := home + "/.config"
	logFolder := home + "/.cache"
	if _, err := os.Stat(configFolder); !os.IsNotExist(err) {
		config.configPath = configFolder + "/cgy.json"
	} else {
		config.configPath = "cgy.json"
	}
	if _, err := os.Stat(logFolder); !os.IsNotExist(err) {
		config.LogFile = logFolder + "/cgy.log"
	} else {
		config.LogFile = "cgy.log"
	}
}

func parseFlags(s int) {
	if s == 0 {
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
		flag.BoolVar(&config.clean, "clean",
			config.clean, "no config and log")
	}
	flag.Parse()
}

func readConfig() error {
	content, err := os.ReadFile(config.configPath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	if string(content) == "" {
		content = []byte("{}")
	}
	err = json.Unmarshal(content, &config)
	if err != nil {
		return fmt.Errorf("Unmarshal config: %w", err)
	}
	return nil
}

func save() error {
	if config.clean {
		return nil
	}
	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	err = os.WriteFile(config.configPath, content, 0644)
	if err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func (c *Config) set(key string, value interface{}) error {
	rv := reflect.ValueOf(c).Elem()
	fv := rv.FieldByName(key)
	fv.Set(reflect.ValueOf(value))
	return save()
}
