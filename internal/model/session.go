package model

import "time"

type Session struct {
	ID               UUID      `db:"id" json:"id"`
	UserID           UUID      `db:"user_id" json:"user_id"`
	RefreshTokenHash string    `db:"refresh_token_hash" json:"-"`
	UserAgent        string    `db:"user_agent" json:"user_agent"`
	IPAddress        string    `db:"ip_address" json:"ip_address"`
	ExpiresAt        time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
	Revoked          bool      `db:"revoked" json:"revoked"`
}
