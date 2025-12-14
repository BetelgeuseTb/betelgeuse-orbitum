package models

import (
	"encoding/json"
	"time"
)

type Scope struct {
	ID          int64
	OrbitID     int64
	Name        string
	Description string
	IsDefault   bool
	IsActive    bool
	Metadata    json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
