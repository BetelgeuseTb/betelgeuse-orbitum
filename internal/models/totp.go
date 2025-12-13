package domain

import "time"

type TOTP struct {
	ID           int64
	UserID       int64
	OrbitID      int64
	SecretCipher string
	Algorithm    string
	Digits       int
	Period       int
	Issuer       string
	Label        string
	LastUsedStep int64
	IsConfirmed  bool
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
