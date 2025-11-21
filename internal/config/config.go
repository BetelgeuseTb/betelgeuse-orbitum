package config

import (
	"os"
	"strconv"
)

type Config struct {
	GRPCAddr    string
	PostgresDSN string
	RedisAddr   string
	RedisDB     int
	RedisPass   string
	JWTKeyPath  string
	JWTKID      string
	Issuer      string
}

func FromEnv() Config {
	return Config{
		GRPCAddr:    getenv("GRPC_ADDR", ":8081"),
		PostgresDSN: getenv("POSTGRES_DSN", "postgres://orbitum:orbitum@localhost:5432/orbitum?sslmode=disable"),
		RedisAddr:   getenv("REDIS_ADDR", "127.0.0.1:6379"),
		RedisDB:     getint("REDIS_DB", 0),
		RedisPass:   getenv("REDIS_PASSWORD", ""),
		JWTKeyPath:  getenv("JWT_KEY_PATH", "./keys/jwt.pem"),
		JWTKID:      getenv("JWT_KID", "kid-default"),
		Issuer:      getenv("JWT_ISSUER", "http://localhost"),
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getint(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return def
}
