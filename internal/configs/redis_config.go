package configs

type RedisConfig struct {
	RedisAddr string
	RedisDB   int
	RedisPass string
}

func GetRedisConfig() RedisConfig {
	return RedisConfig{
		RedisAddr: getEnv("REDIS_URL", "127.0.0.1:6379"),
		RedisDB:   getInt("REDIS_DB", 0),
		RedisPass: getEnv("REDIS_PASSWORD", ""),
	}
}
