package models

import "time"

type RevokedToken struct {
	ID        int64
	JTI       string
	ExpiresAt time.Time
	OrbitID   int64
	Reason    string
	CreatedAt time.Time
}
