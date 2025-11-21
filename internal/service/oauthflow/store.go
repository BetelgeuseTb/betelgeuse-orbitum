package oauthflow

import (
	"context"
	"encoding/json"
	"time"

	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/service/cache"
)

type CodeMeta struct {
	ClientID            string    `json:"client_id"`
	UserID              string    `json:"user_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scopes              []string  `json:"scopes"`
	CodeChallenge       *string   `json:"code_challenge,omitempty"`
	CodeChallengeMethod *string   `json:"code_challenge_method,omitempty"`
	ExpiresAt           time.Time `json:"expires_at"`
}

type Store struct {
	cache cache.Cache
}

func NewStore(c cache.Cache) *Store {
	return &Store{cache: c}
}

func (s *Store) Save(ctx context.Context, code string, meta CodeMeta, ttl time.Duration) error {
	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return s.cache.Set(s.key(code), string(b), ttl)
}

func (s *Store) Get(ctx context.Context, code string) (*CodeMeta, bool, error) {
	str, ok, err := s.cache.Get(s.key(code))
	if err != nil || !ok {
		return nil, false, err
	}
	var m CodeMeta
	if err := json.Unmarshal([]byte(str), &m); err != nil {
		return nil, false, err
	}
	return &m, true, nil
}

func (s *Store) Delete(ctx context.Context, code string) error {
	return s.cache.Del(s.key(code))
}

func (s *Store) key(code string) string {
	return "oauthcode:" + code
}
