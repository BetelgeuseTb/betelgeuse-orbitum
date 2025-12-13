package configs

type AppConfig struct {
	ServerPort   string
	JwtSecretKey string
}

func GetAppConfig() AppConfig {
	return AppConfig{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		JwtSecretKey: getEnv("JWT_SECRET_KEY", ""),
	}
}
