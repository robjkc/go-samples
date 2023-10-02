package main

import (
	"log"

	"haproxy-audit/agent"
	"haproxy-audit/config"
	"haproxy-audit/db"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	initLogFile()

	// Set the config to the global agent config.
	config := config.LoadConfig()

	_, err := db.ConnectDb(config.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	agent.StartAgent(config)
}

func initLogFile() {
	log.SetOutput(&lumberjack.Logger{
		Filename:   "/var/log/haproxy-audit.log",
		MaxSize:    1, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
}
