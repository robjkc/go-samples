package config

import (
	"log"

	"github.com/spf13/viper"
)

var (
	defaults = map[string]interface{}{
		"eventTable": "event_logs",
	}
	configName  = "config"
	configPaths = []string{
		"/opt/haproxy-log-parser/",
		".",
	}
)

type Config struct {
	ConnectionString string
	EventTable       string
}

func LoadConfig() *Config {
	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
	viper.SetConfigName(configName)
	for _, p := range configPaths {
		viper.AddConfigPath(p)
	}
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("could not read config file: %v", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("could not decode config into struct: %v", err)
	}
	log.Printf("Config struct: %#v\n", config)
	return &config
}