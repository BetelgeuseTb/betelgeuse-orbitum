package httptransport

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

func ParseBasicAuth(h string) (string, string, error) {
	if h == "" {
		return "", "", errors.New("missing")
	}
	if !strings.HasPrefix(strings.ToLower(h), "basic ") {
		return "", "", errors.New("bad scheme")
	}
	b64 := strings.TrimSpace(h[6:])
	dec, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(dec), ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("bad header")
	}
	return parts[0], parts[1], nil
}

func ParseBearer(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("missing")
	}
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return "", errors.New("bad scheme")
	}
	return strings.TrimSpace(h[7:]), nil
}

func PKCEChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
