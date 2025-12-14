package models

import (
	"encoding/json"
	"time"
)

type AuthCode struct {
	ID                  int64
	Code                string
	OrbitID             int64
	ClientID            int64
	UserID              *int64
	RedirectURI         string
	Scope               json.RawMessage
	CodeChallenge       string
	CodeChallengeMethod string
	Used                bool
	Metadata            json.RawMessage
	CreatedAt           time.Time
	ExpiresAt           time.Time
}
