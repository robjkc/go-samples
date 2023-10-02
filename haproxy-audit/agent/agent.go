package agent

import (
	"github./securecom/haproxy-audit/config"
	"log"
	"strconv"

	spoe "github.com/criteo/haproxy-spoe-go"
)

var (
	Config *config.Config
)

func StartAgent(config *config.Config) {
	Config = config

	agent := spoe.New(messageHandler)

	// Use the port set in the global agent config.
	log.Println("Starting Agent at port", config.Port)
	if err := agent.ListenAndServe(":" + strconv.Itoa(config.Port)); err != nil {
		log.Fatal(err)
	}
}

func messageHandler(messages *spoe.MessageIterator) ([]spoe.Action, error) {
	for messages.Next() {
		msg := messages.Message
		if msg.Name == "audit-response" {
			HandleAuditResponse(msg)
		}
	}

	return nil, nil
}
