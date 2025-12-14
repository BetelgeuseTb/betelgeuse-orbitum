package models

import (
	"encoding/json"
	"time"
)

type Client struct {
	ID                      int64
	OrbitID                 int64
	ClientID                string
	ClientSecretHash        string
	Name                    string
	Description             string
	RedirectURIs            json.RawMessage
	PostLogoutRedirectURIs  json.RawMessage
	GrantTypes              json.RawMessage
	ResponseTypes           json.RawMessage
	TokenEndpointAuthMethod string
	Contacts                json.RawMessage
	LogoURI                 string
	AppType                 string
	IsPublic                bool
	IsActive                bool
	AllowedCORSOrigins      json.RawMessage
	AllowedScopes           json.RawMessage
	Metadata                json.RawMessage
	CreatedAt               time.Time
	UpdatedAt               time.Time
	DeletedAt               *time.Time
}
