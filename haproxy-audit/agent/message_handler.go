package agent

import (
	"github./securecom/haproxy-audit/config"
	"log"

	spoe "github.com/criteo/haproxy-spoe-go"
	"github./securecom/haproxy-audit/db"
	"github./securecom/haproxy-audit/models"
)

func HandleAuditResponse(msg spoe.Message) {
	agentMsg, err := NewAgentMessage(msg)
	if err != nil {
		if _, ok := err.(*NotAuditableError); ok {
			// Log out the not auditable errors.
			log.Println(err)
		}

		return
	}

	go handleMessage(agentMsg)
}

func handleMessage(msg *AgentMessage) {

	handler := NewMessageHandler(msg)

	err := handler.addEventLog()
	if err != nil {
		log.Printf("Unable to add event log %v", err)
	}
	log.Printf("Agent message logged %+v\n", msg)
}

type MessageHandler struct {
	Message *AgentMessage
	User    *models.User
	Person  *models.Person
	con     *db.DbConnection
	config	*config.Config
}

func NewMessageHandler(msg *AgentMessage) *MessageHandler {
	handler := &MessageHandler{}
	handler.Message = msg
	handler.con = db.DbConn
	handler.config = Config
	return handler
}

func (m *MessageHandler) addEventLog() error {

	err := models.AddEventLog(m.con, m.config.EventTable, m.Message.QueryAuthToken, m.Message.IpAddress,
		m.Message.Method, m.Message.PathApiVersion, m.Message.PathApi, m.Message.PathApiId,
		m.Message.PathConcept, m.Message.PathConceptId, m.Message.PathAction)
	if err != nil {
		return err
	}

	return nil
}
