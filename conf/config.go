package conf

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	JWT    JWTConfig
	MySQL  MySQLConfig
	Cookie CookieConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  int
	RefreshTokenExpiry int
}

type MySQLConfig struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
}

type CookieConfig struct {
	Domain string
	Secure bool
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	mysqlPort, err := strconv.Atoi(getEnv("MYSQL_PORT", "3306"))
	if err != nil {
		mysqlPort = 3306
	}

	accessTokenExpiry, err := strconv.Atoi(getEnv("JWT_ACCESS_TOKE_EXP", "24"))
	if err != nil {
		accessTokenExpiry = 24
	}

	refreshTokenExpiry, err := strconv.Atoi(getEnv("JWT_REFRESH_TOKE_EXP", "72"))
	if err != nil {
		refreshTokenExpiry = 72
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		redisDB = 0
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", ":3000"),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "default_secret"),
			AccessTokenExpiry:  accessTokenExpiry,
			RefreshTokenExpiry: refreshTokenExpiry,
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     mysqlPort,
			DBName:   getEnv("MYSQL_DBNAME", "ecom"),
			Username: getEnv("MYSQL_USERNAME", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
		},
		Cookie: CookieConfig{
			Domain: getEnv("COOKIE_DOMAIN", "localhost"),
			Secure: getEnv("COOKIE_SECURE", "false") == "true",
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       redisDB,
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
