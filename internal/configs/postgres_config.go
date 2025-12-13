package configs

type PostgresConfig struct {
	DatabaseName string
	UserName     string
	Password     string
}

func GetPostgresConfig() PostgresConfig {
	return PostgresConfig{
		DatabaseName: getEnv("POSTGRES_DB", "8080"),
		UserName:     getEnv("POSTGRES_USER", ""),
		Password:     getEnv("POSTGRES_PASS", ""),
	}
}
