package message

import (
	"errors"
	"fmt"
)

var NotJson = errors.New("not a json log message")
var ErrShouldntParse = errors.New("shouldn't parse")

type NotAuditableError struct {
	Status int
	Method string
	Host   string
	Path   string
}

func (e *NotAuditableError) Error() string {
	return fmt.Sprintf("Not auditable: %d %v %v %v", e.Status, e.Method, e.Host, e.Path)
}