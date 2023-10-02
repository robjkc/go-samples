package models

import (
	"database/sql"
	"fmt"
	"time"

	"haproxy-audit/db"
)

var utc *time.Location = nil

func AddEventLog(con *db.DbConnection, eventTable string, authToken string, ipAddress string, eventType string,
	apiVersion string, api string, apiId string, concept string, conceptId int, action string) error {

	sql := fmt.Sprintf(`INSERT INTO %s
		(event_at, auth_token, ip_address, event_type, api_version, api, api_id, concept, concept_id, action)
		VALUES (:eventAt, :authToken, :ipAddress, :eventType, :apiVersion, :api, :apiId, :concept, :conceptId, :action)`, eventTable)

	if utc == nil {
		// Load the UTC time zone - https://www.golangprograms.com/golang-get-current-date-and-time-in-est-utc-and-mst.html
		utc, _ = time.LoadLocation("UTC")
	}

	err := con.ExecuteUpdate(sql,
		db.Args{
			"eventAt":    time.Now().In(utc),
			"authToken":  authToken,
			"ipAddress":  ipAddress,
			"eventType":  eventType,
			"apiVersion": apiVersion,
			"api":        api,
			"apiId":      apiId,
			"concept":    NewNullString(concept),
			"conceptId":  NewNullInt32(conceptId),
			"action":     NewNullString(action),
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func NewNullInt32(i int) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{
		Int32: int32(i),
		Valid: true,
	}
}
