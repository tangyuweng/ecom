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
}

type ServerConfig struct {
	Port string
}

type JWTConfig struct {
	Secret   string
	Duration int
}

type MySQLConfig struct {
	Host     string
	Port     int
	DBName   string
	Username string
	Password string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	mysqlPort, err := strconv.Atoi(getEnv("MYSQL_PORT", "3306"))
	if err != nil {
		mysqlPort = 3306
	}

	jwtDuration, err := strconv.Atoi(getEnv("JWT_DURATION", "24"))
	if err != nil {
		jwtDuration = 24
	}

	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", ":3000"),
		},
		JWT: JWTConfig{
			Secret:   getEnv("JWT_SECRET", "default_secret"),
			Duration: jwtDuration,
		},
		MySQL: MySQLConfig{
			Host:     getEnv("MYSQL_HOST", "localhost"),
			Port:     mysqlPort,
			DBName:   getEnv("MYSQL_DBNAME", "ecom"),
			Username: getEnv("MYSQL_USERNAME", "root"),
			Password: getEnv("MYSQL_PASSWORD", ""),
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
