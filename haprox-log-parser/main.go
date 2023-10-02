package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"haproxy-log-parser/config"
	"haproxy-log-parser/db"
	"haproxy-log-parser/message"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	MessageC chan<- message.LogMessage
)

func run() error {
	go handleSignals()
	err := work()
	if err != nil {
		return err
	}

	return nil
}

func work() error {

	in := bufio.NewReader(os.Stdin)

	for {
		msg, err := in.ReadString('\n')
		if err != nil {
			return err
		}

		// Create a new log message.
		logMessage, err := message.NewLogMessage(msg)
		if err != nil {
			// Unable to create a log message so ignore.
			log.Println("Go error", err)
			continue
		}

		// Add the log message onto the channel.
		MessageC <- logMessage

		time.Sleep(1 * time.Millisecond)
	}
}

func main() {
	// Init the log file.
	initLogFile()

	// Set the config to the global agent config.
	message.Config = config.LoadConfig()

	_, err := db.ConnectDb(message.Config.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}

	messageC := make(chan message.LogMessage, 100)
	MessageC = messageC
	numWorkers := 25

	// Create the message worker pool.
	log.Printf("Starting worker pool of %d workers\n", numWorkers)
	for w := 1; w <= numWorkers; w++ {
		go worker(w, messageC)
	}

	err = run()
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func worker(workerId int, messageC <-chan message.LogMessage) {
	for logMsg := range messageC {
		message.HandleLogMessage(logMsg)
	}
}

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)

	for s := range signals {
		switch s {
		case syscall.SIGUSR1, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt:
			// Catch signals that might terminate the process on behalf all goroutines.
			quit()
		}
	}
}

func quit() {
	// Perform any necessary cleanup here.
	os.Exit(1)
}

func initLogFile() {
	logFile := os.Getenv("LOGFILE")
	if len(logFile) == 0 {
		logFile = "/var/log/haproxy-log-parser.log"
	}

	log.SetOutput(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    100, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
}