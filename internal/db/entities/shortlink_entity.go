package entities

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Shortlink represents a shortlink to a workflow JSON or URL.
type Shortlink struct {
	Slug          string      `json:"slug,omitempty" db:"slug"`                     // added by API
	Kind          string      `json:"kind" db:"kind"`                               // required, 'workflow' or 'url'
	Content       string      `json:"content" db:"content"`                         // required, JSON or URL
	CreatorIP     string      `json:"creator_ip,omitempty" db:"creator_ip"`         // added by API
	CreatedAt     CustomTime  `json:"created_at,omitempty" db:"created_at"`         // added by DB
	ExpiresAt     *CustomTime `json:"expires_at,omitempty" db:"expires_at"`         // optional
	Password      string      `json:"password,omitempty" db:"password"`             // optional
	AllowedVisits int         `json:"allowed_visits,omitempty" db:"allowed_visits"` // optional, -1 for unlimited
}

// CustomTime handles timestamp conversion between Go's time.Time and sqlite's TEXT.
type CustomTime struct {
	time.Time
}

const timeLayout = "2006-01-02 15:04:05"

// Scan converts a sqlite TEXT timestamp into a CustomTime.
func (ct *CustomTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		parsedTime, err := time.Parse(timeLayout, v)
		if err != nil {
			return fmt.Errorf("parsing time for CustomTime: %w", err)
		}
		ct.Time = parsedTime
	default:
		return fmt.Errorf("unsupported scan type for CustomTime: %T", v)
	}
	return nil
}

// Value converts a CustomTime into a sqlite TEXT timestamp
func (ct CustomTime) Value() (driver.Value, error) {
	return ct.Time.Format(timeLayout), nil
}
