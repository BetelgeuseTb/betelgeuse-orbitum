package models

import (
	"encoding/json"
	"time"
)

type User struct {
	ID                 int64
	OrbitID            int64
	Username           string
	Email              string
	EmailVerified      bool
	PasswordHash       string
	PasswordAlgo       string
	LastPasswordChange *time.Time
	DisplayName        string
	Profile            json.RawMessage
	IsActive           bool
	IsLocked           bool
	MFAEnabled         bool
	Metadata           json.RawMessage
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          *time.Time
}
