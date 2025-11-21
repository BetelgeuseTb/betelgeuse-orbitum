package security

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	Sub     string   `json:"sub"`
	Aud     string   `json:"aud"`
	Scope   []string `json:"scope"`
	Roles   []int    `json:"roles"`
	JTI     string   `json:"jti"`
	Iss     string   `json:"iss"`
	Iat     int64    `json:"iat"`
	Exp     int64    `json:"exp"`
	Client  string   `json:"client"`
	Session string   `json:"sid,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager interface {
	SignAccessToken(c Claims) (string, time.Time, error)
	Validate(tokenStr string, audience string, issuer string) (*Claims, error)
}

type RS256JWTManager struct {
	priv *rsa.PrivateKey
	kid  string
}

func NewRS256JWTManager(kp KeyProvider) *RS256JWTManager {
	return &RS256JWTManager{priv: kp.PrivateKey(), kid: kp.KID()}
}

func (m *RS256JWTManager) SignAccessToken(c Claims) (string, time.Time, error) {
	now := time.Now()
	jti := uuid.NewString()
	exp := now.Add(time.Minute * 15)
	c.JTI = jti
	c.Iat = now.Unix()
	c.Exp = exp.Unix()
	t := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub":    c.Sub,
		"aud":    c.Aud,
		"scope":  c.Scope,
		"roles":  c.Roles,
		"jti":    c.JTI,
		"iss":    c.Iss,
		"iat":    c.Iat,
		"exp":    c.Exp,
		"client": c.Client,
		"sid":    c.Session,
	})
	t.Header["kid"] = m.kid
	signed, err := t.SignedString(m.priv)
	return signed, exp, err
}

func (m *RS256JWTManager) Validate(tokenStr string, audience string, issuer string) (*Claims, error) {
	keyFunc := func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodRS256 {
			return nil, errors.New("bad alg")
		}
		return &m.priv.PublicKey, nil
	}
	tok, err := jwt.Parse(tokenStr, keyFunc, jwt.WithAudience(audience), jwt.WithIssuer(issuer))
	if err != nil {
		return nil, err
	}
	claimsMap, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, errors.New("invalid")
	}
	var c Claims
	c.Sub, _ = claimsMap["sub"].(string)
	c.Aud, _ = claimsMap["aud"].(string)
	c.Iss, _ = claimsMap["iss"].(string)
	c.JTI, _ = claimsMap["jti"].(string)
	c.Client, _ = claimsMap["client"].(string)
	c.Session, _ = claimsMap["sid"].(string)
	c.Scope = strSlice(claimsMap["scope"])
	c.Roles = intSlice(claimsMap["roles"])
	if exp, ok := claimsMap["exp"].(float64); ok {
		c.Exp = int64(exp)
	}
	if iat, ok := claimsMap["iat"].(float64); ok {
		c.Iat = int64(iat)
	}
	return &c, nil
}

func strSlice(v any) []string {
	a, ok := v.([]any)
	if !ok {
		return []string{}
	}
	out := make([]string, 0, len(a))
	for _, i := range a {
		if s, ok := i.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func intSlice(v any) []int {
	a, ok := v.([]any)
	if !ok {
		return []int{}
	}
	out := make([]int, 0, len(a))
	for _, i := range a {
		switch t := i.(type) {
		case float64:
			out = append(out, int(t))
		case int:
			out = append(out, t)
		}
	}
	return out
}

func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
