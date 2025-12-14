package models

import (
	"encoding/json"
	"time"
)

type Session struct {
	ID           int64
	OrbitID      int64
	UserID       int64
	ClientID     *int64
	StartedAt    time.Time
	LastActiveAt time.Time
	ExpiresAt    *time.Time
	Revoked      bool
	DeviceInfo   string
	IP           string
	Metadata     json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
