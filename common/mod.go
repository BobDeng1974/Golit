package common

import (
	json "encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type ConfigEntry struct {
	Key   string
	Value string
}

type HueConfig struct {
	IP     string
	User   string
	Paired bool
}

type MqttConfig struct {
	Host         string
	QueryTimeout time.Duration
}

type Config struct {
	MQTT        MqttConfig
	Hue         HueConfig
	HostAddress string
}

const ConfigFile string = "./config.json"
const DatabaseFile string = "./golit.db"
const DatabaseSetup string = "./golit.sql"

func LoadConfig() Config {
	config := Config{}
	doc, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	json.Unmarshal(doc, &config)
	return config
}

func WriteConfig(cfg *Config) {
	data, err := json.Marshal(cfg)
	if err != nil {
		log.Fatal("Failed to marshal config")
		os.Exit(1)
	}
	ioerr := ioutil.WriteFile(ConfigFile, data, 0644)
	if ioerr != nil {
		log.Fatal("Failed to write config")
		os.Exit(1)
	}
}
