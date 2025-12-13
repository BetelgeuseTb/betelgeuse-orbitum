package domain

import (
	"encoding/json"
	"time"
)

type RefreshToken struct {
	ID            int64
	ExpiresAt     time.Time
	TokenString   string
	JTI           string
	OrbitID       int64
	ClientID      int64
	UserID        *int64
	Revoked       bool
	RotatedFromID *int64
	RotatedToID   *int64
	Scopes        json.RawMessage
	Metadata      json.RawMessage
	LastUsedAt    *time.Time
	UseCount      int
	CreatedAt     time.Time
}
