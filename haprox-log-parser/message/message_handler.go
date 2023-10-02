package message

import (
	"log"

	"haproxy-log-parser/config"
	"haproxy-log-parser/models"

	"haproxy-log-parser/db"
)

var (
	Config *config.Config
)

func HandleLogMessage(logMsg LogMessage) {
	err := logMsg.parse()
	if err != nil {
		if _, ok := err.(*NotAuditableError); ok {
			// Log out the not auditable errors.
			log.Println(err)
		}

		return
	}

	handleMessage(&logMsg)
}

func handleMessage(msg *LogMessage) {

	handler := NewMessageHandler(msg)

	err := handler.addEventLog()
	if err != nil {
		log.Printf("Unable to add event log %v", err)
	}
	log.Printf("Audit message logged %+v\n", msg)
}

type MessageHandler struct {
	Message *LogMessage
	con     *db.DbConnection
	config  *config.Config
}

func NewMessageHandler(msg *LogMessage) *MessageHandler {
	handler := &MessageHandler{}
	handler.Message = msg
	handler.con = db.DbConn
	handler.config = Config
	return handler
}

func (m *MessageHandler) addEventLog() error {

	err := models.AddEventLog(m.con, m.config.EventTable, m.Message.QueryAuthToken, m.Message.IpAddress,
		m.Message.Time, m.Message.Method, m.Message.PathApiVersion, m.Message.PathApi, m.Message.PathApiId,
		m.Message.PathConcept, m.Message.PathConceptId, m.Message.PathAction)
	if err != nil {
		return err
	}

	return nil
}