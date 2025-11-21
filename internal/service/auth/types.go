package auth

import (
	"time"
)

type LoginResult struct {
	AccessToken  string
	AccessExpiry time.Time
	RefreshToken string
	SessionID    string
	UserID       string
	Scopes       []string
	Roles        []int
}

type RefreshResult struct {
	AccessToken  string
	AccessExpiry time.Time
	SessionID    string
}

type ValidateResult struct {
	UserID string
	Client string
	Scopes []string
	Roles  []int
	JTI    string
	Exp    int64
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email     string
	Password  string
	ClientID  string
	Scopes    []string
	UserAgent string
	IP        string
}

type RefreshInput struct {
	RefreshToken string
	ClientID     string
	UserAgent    string
	IP           string
}
