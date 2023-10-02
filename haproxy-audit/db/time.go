package db

import (
	"fmt"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface for NullTime
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		// . snowflake already has converted to time.Time
		nt.Time, nt.Valid = v, true
		return nil
	}
	return fmt.Errorf("can't convert %T to time.Time", value)
}

func (nt *NullTime) IsAfter(checkTime time.Time) bool {

	if !nt.Valid {
		return false
	}

	return nt.Time.After(checkTime)
}
