package agent

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	spoe "github.com/criteo/haproxy-spoe-go"
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

type AgentMessage struct {
	UniqueId  string
	IpAddress string
	Host      string
	Path      string
	PathQuery string
	Method    string
	Status    int

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

func NewAgentMessage(msg spoe.Message) (*AgentMessage, error) {
	agentMsg := &AgentMessage{}
	agentMsg.init(msg)

	if !agentMsg.shouldParse() {
		return nil, ErrShouldntParse
	}

	// Parse the path.
	err := agentMsg.parsePath()
	if err != nil {
		return nil, err
	}

	agentMsg.parseQuery()

	if !agentMsg.shouldAudit() {
		return nil, &NotAuditableError{
			Status: agentMsg.Status,
			Method: agentMsg.Method,
			Host:   agentMsg.Host,
			Path:   agentMsg.Path,
		}
	}

	return agentMsg, nil
}

func (m *AgentMessage) init(msg spoe.Message) {
	for msg.Args.Next() {
		arg := msg.Args.Arg
		value := fmt.Sprintf("%v", arg.Value)

		switch arg.Name {
		case "uid":
			m.UniqueId = value
		case "ip":
			m.IpAddress = value
		case "host":
			m.Host = value
		case "path":
			if len(value) <= 1 {
				m.Path = ""
			} else {
				// Remove leading "/" so our split string doesn't return a blank value.
				m.Path = value[1:]
				m.Path = strings.TrimSuffix(m.Path, "/")
			}
		case "query":
			m.PathQuery = value
		case "method":
			m.Method = value
		case "status":
			if status, err := strconv.Atoi(value); err == nil {
				m.Status = status
			}
		}
	}
}

func (m *AgentMessage) shouldParse() bool {
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

func (m *AgentMessage) shouldAudit() bool {
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

func (m *AgentMessage) parseQuery() {
	regex := *regexp.MustCompile(`auth_token=([^&]{20})`)

	res := regex.FindAllStringSubmatch(m.PathQuery, -1)
	for i := range res {
		//like Java: match.group(1), match.gropu(2), etc
		m.QueryAuthToken = res[i][1]
	}
}

func (m *AgentMessage) parsePath() error {
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

func (m *AgentMessage) parseV2Path() error {
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

func (m *AgentMessage) parseV1Path() error {
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

func (m *AgentMessage) parseGenericPath() error {
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

func (m *AgentMessage) IsV1() bool {
	return m.PathApiVersion == "1"
}

func (m *AgentMessage) IsV2() bool {
	return m.PathApiVersion == "v2"
}
