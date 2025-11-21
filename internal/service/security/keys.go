package security

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

type KeyProvider interface {
	PrivateKey() *rsa.PrivateKey
	KID() string
}

type StaticRSAKeyProvider struct {
	key *rsa.PrivateKey
	kid string
}

func NewStaticRSAKeyProviderFromPEM(path string, kid string) (*StaticRSAKeyProvider, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid pem")
	}
	var pk *rsa.PrivateKey
	switch block.Type {
	case "RSA PRIVATE KEY":
		pk, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	default:
		keyAny, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, err2
		}
		rsaKey, ok := keyAny.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not rsa private key")
		}
		pk = rsaKey
	}
	if err != nil {
		return nil, err
	}
	return &StaticRSAKeyProvider{key: pk, kid: kid}, nil
}

func (s *StaticRSAKeyProvider) PrivateKey() *rsa.PrivateKey { return s.key }
func (s *StaticRSAKeyProvider) KID() string                  { return s.kid }