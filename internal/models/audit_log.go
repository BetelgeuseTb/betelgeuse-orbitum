package domain

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID            int64
	ActorUserID   *int64
	ActorClientID *int64
	Action        string
	Result        string
	IP            string
	OrbitID       int64
	Details       json.RawMessage
	CreatedAt     time.Time
}
