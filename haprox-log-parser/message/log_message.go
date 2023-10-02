package message

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	ignoreApis = []string{
		"authenticate",
		"mobile_devices",
		"report",
		"jobs",
	}

	ignoreConcepts = []string{
		"send_sempro_message",
		"disconnect",
		"armed_status",
		"cancel",
	}

	ignoreActions = []string{
		"refresh",
		"activate",
		"sign_in",
	}
)

type LogMessage struct {
	IpAddress string `json:"ip"`
	Time      string `json:"time"`
	Method    string `json:"method"`
	Host      string `json:"host"`
	Path      string `json:"path"`
	Query     string `json:"query"`
	Status    int    `json:"status"`

	// Derived fields from the path.
	PathApiVersion string
	PathApi        string
	PathApiId      string
	PathConcept    string
	PathConceptId  int
	PathAction     string

	// Derived fields from the query params
	QueryAuthToken string
}

func NewLogMessage(msg string) (LogMessage, error) {
	logMsg := LogMessage{}
	err := logMsg.init(msg)
	return logMsg, err
}

func (m *LogMessage) parse() error {
	if !m.shouldParse() {
		return ErrShouldntParse
	}

	// Parse the path.
	err := m.parsePath()
	if err != nil {
		return err
	}

	m.parseQuery()

	if !m.shouldAudit() {
		return &NotAuditableError{
			Status: m.Status,
			Method: m.Method,
			Host:   m.Host,
			Path:   m.Path,
		}
	}

	return nil
}

func (m *LogMessage) init(msg string) error {
	index := strings.Index(msg, "{\"host")
	if index < 0 {
		// This is not a JSON message.
		return NotJson
	}
	// Grab the JSON out of the log message.
	jsonString := msg[index : len(msg)-1]

	err := json.Unmarshal([]byte(jsonString), m)
	if err != nil {
		return err
	}

	// Remove leading "/" so our split string doesn't return a blank value.
	m.Path = m.Path[1:]
	m.Path = strings.TrimSuffix(m.Path, "/")

	return nil
}

func (m *LogMessage) shouldParse() bool {
	if m.Status < 200 || m.Status >= 300 {
		// Not a successfull status code.
		return false
	}

	if !(strings.HasPrefix(m.Host, "api") || strings.HasPrefix(m.Host, "vk")) {
		// Not a SCAPI message.
		return false
	}

	if !matchesOneOf(m.Method, "put", "post", "patch", "delete") {
		// Not an update call.
		return false
	}

	return true
}

func (m *LogMessage) shouldAudit() bool {
	if len(m.QueryAuthToken) == 0 {
		return false
	}

	if matchesOneOf(m.PathApi, ignoreApis...) {
		return false
	}

	if _, err := strconv.Atoi(m.PathConcept); err == nil {
		// Concept is a number so not valid.
		return false
	}

	if matchesOneOf(m.PathConcept, ignoreConcepts...) {
		return false
	}

	if matchesOneOf(m.PathAction, ignoreActions...) {
		return false
	}

	return true
}

func matchesOneOf(str string, matches ...string) bool {
	for _, match := range matches {
		if strings.EqualFold(str, match) {
			return true
		}
	}
	return false
}

func (m *LogMessage) parseQuery() {
	regex := *regexp.MustCompile(`auth_token=([^&]{20})`)

	res := regex.FindAllStringSubmatch(m.Query, -1)
	for i := range res {
		//like Java: match.group(1), match.gropu(2), etc
		m.QueryAuthToken = res[i][1]
	}
}

func (m *LogMessage) parsePath() error {
	s := strings.Split(m.Path, "/")

	if len(s) < 2 {
		return fmt.Errorf("path not long enough %s", m.Path)
	}

	m.PathApiVersion = s[0]
	if m.PathApiVersion == "v2" {
		err := m.parseV2Path()
		if err != nil {
			return err
		}
	} else if m.PathApiVersion == "1" {
		err := m.parseV1Path()
		if err != nil {
			return err
		}
	} else {
		err := m.parseGenericPath()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *LogMessage) parseV2Path() error {
	s := strings.Split(m.Path, "/")
	len := len(s)

	// example: v2/panels/568969/zone_informations
	if len >= 2 {
		m.PathApi = s[1]
	}
	if len >= 3 {
		m.PathApiId = s[2]
	}
	if len >= 4 {
		m.PathConcept = s[3]
	}
	if len < 5 {
		return nil
	}
	if len == 5 {
		if i, err := strconv.Atoi(s[4]); err == nil {
			m.PathConceptId = i
		} else {
			m.PathAction = s[4]
		}
	} else if len == 6 {
		if i, err := strconv.Atoi(s[4]); err == nil {
			m.PathConceptId = i
		}
		m.PathAction = s[5]
	} else {
		return fmt.Errorf("unable to parse V2 route with %d sections: %v", len, s)
	}
	return nil
}

func (m *LogMessage) parseV1Path() error {
	s := strings.Split(m.Path, "/")
	len := len(s)

	// example: api/1/panels/4-8489/favorites/44974/actions/195666/?app=Apple&auth_token=rvLFkLxnmS-65WjxS8ru&auth_user_code=1234&device_system_version=14.7.1&v=6.39.2_87dbab8
	if len >= 3 {
		m.PathApi = s[1]
		m.PathApiId = s[2]
	}
	if len < 4 {
		return nil
	}

	if len == 4 {
		m.PathConcept = s[3]
	} else if len == 5 {
		m.PathConcept = s[3]
		if i, err := strconv.Atoi(s[4]); err == nil {
			m.PathConceptId = i
		} else {
			m.PathAction = s[4]
		}
	} else if len == 6 {
		m.PathConcept = s[3]
		if i, err := strconv.Atoi(s[4]); err == nil {
			m.PathConceptId = i
		}
		m.PathAction = s[5]

	} else if len == 7 {
		if s[3] == "favorites" && s[5] == "actions" {
			m.PathConcept = "favorite_actions"
			if i, err := strconv.Atoi(s[6]); err == nil {
				m.PathConceptId = i
			}
		} else if s[3] == "rooms" && s[5] == "hotspots" {
			m.PathConcept = "room_hotspots"
			if i, err := strconv.Atoi(s[6]); err == nil {
				m.PathConceptId = i
			}
		} else {
			return fmt.Errorf("unable to parse V1 route with %d sections: %v", len, s)
		}
	} else {
		return fmt.Errorf("unable to parse V1 route with %d sections: %v", len, s)
	}
	return nil
}

func (m *LogMessage) parseGenericPath() error {
	s := strings.Split(m.Path, "/")
	len := len(s)

	if len > 0 {
		m.PathApi = s[0]
	}
	if len >= 1 {
		if m.PathApi == "users" {
			m.PathAction = s[1]
		} else {
			return fmt.Errorf("unable to parse generic route with %d sections: %v", len, s)
		}
	}

	return nil
}

func (m *LogMessage) IsV1() bool {
	return m.PathApiVersion == "1"
}

func (m *LogMessage) IsV2() bool {
	return m.PathApiVersion == "v2"
}