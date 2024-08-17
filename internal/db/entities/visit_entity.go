package entities

// Visit represents an access to a shortlink.
type Visit struct {
	ID        int        `json:"id" db:"id"`
	Slug      string     `json:"slug" db:"slug"`
	TS        CustomTime `json:"ts" db:"ts"`
	Referer   string     `json:"referer" db:"referer"`
	UserAgent string     `json:"user_agent" db:"user_agent"`
}
