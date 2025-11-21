package security

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type PasswordHasher interface {
	Hash(plaintext []byte) (string, error)
	Verify(plaintext []byte, encoded string) (bool, error)
}

type Argon2idHasher struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

func NewArgon2idHasher() *Argon2idHasher {
	return &Argon2idHasher{
		Time:    2,
		Memory:  64 * 1024,
		Threads: 2,
		KeyLen:  32,
	}
}

func (a *Argon2idHasher) Hash(plaintext []byte) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	key := argon2.IDKey(plaintext, salt, a.Time, a.Memory, a.Threads, a.KeyLen)
	params := fmt.Sprintf("$argon2id$v=%d$t=%d$m=%d,p=%d$%s$%s",
		19, a.Time, a.Memory, a.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	return params, nil
}

func (a *Argon2idHasher) Verify(plaintext []byte, encoded string) (bool, error) {
	var v, t, m, p uint32
	var saltB64, keyB64 string
	_, err := fmt.Sscanf(encoded, "$argon2id$v=%d$t=%d$m=%d,p=%d$%s$%s", &v, &t, &m, &p, &saltB64, &keyB64)
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, err
	}
	storedKey, err := base64.RawStdEncoding.DecodeString(keyB64)
	if err != nil {
		return false, err
	}
	key := argon2.IDKey(plaintext, salt, t, m, uint8(p), uint32(len(storedKey)))
	if !bytes.Equal(key, storedKey) {
		return false, nil
	}
	return true, nil
}

func VerifyOrError(h PasswordHasher, plaintext []byte, encoded string) error {
	ok, err := h.Verify(plaintext, encoded)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("invalid credentials")
	}
	return nil
}
