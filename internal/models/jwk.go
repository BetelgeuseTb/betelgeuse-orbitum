package domain

import (
	"encoding/json"
	"time"
)

type JWKey struct {
	ID               int64
	OrbitID          int64
	Kid              string
	Use              string
	Alg              string
	Kty              string
	PublicKeyJWK     json.RawMessage
	PrivateKeyCipher string
	IsActive         bool
	NotBefore        *time.Time
	ExpiresAt        *time.Time
	Metadata         json.RawMessage
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
