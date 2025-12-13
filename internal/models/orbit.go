package domain

import (
	"encoding/json"
	"time"
)

type Orbit struct {
	ID            int64
	Name          string
	DisplayName   string
	Description   string
	Issuer        string
	Domain        string
	Config        json.RawMessage
	DefaultScopes json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}
