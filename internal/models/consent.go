package models

import (
	"encoding/json"
	"time"
)

type Consent struct {
	ID        int64
	OrbitID   int64
	UserID    int64
	ClientID  int64
	Scopes    json.RawMessage
	GrantedAt time.Time
	ExpiresAt *time.Time
	Revoked   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
