package models

import (
	"encoding/json"
	"time"
)

type AccessToken struct {
	ID             int64
	JTI            string
	OrbitID        int64
	ClientID       int64
	UserID         *int64
	IsJWT          bool
	TokenString    string
	Scope          json.RawMessage
	IssuedAt       time.Time
	TokenType      string
	Revoked        bool
	Metadata       json.RawMessage
	RefreshTokenID *int64
	CreatedAt      time.Time
	ExpiresAt      time.Time
}
