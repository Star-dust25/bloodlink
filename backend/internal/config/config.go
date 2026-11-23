package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	GinMode            string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	DBMaxOpenConns     int
	DBMaxIdleConns     int
	DBConnMaxLifetimeM int
	JWTSecret          string
	JWTExpirationHours int
	AESKeyHex          string
	CORSAllowedOrigins []string
	SOAPEndpointURL    string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	soapURL := getEnv("SOAP_ENDPOINT_URL", fmt.Sprintf("http://localhost:%s/soap/v1/interoperabilidad", port))

	return &Config{
		Port:               port,
		GinMode:            getEnv("GIN_MODE", "debug"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "bloodlink"),
		DBPassword:         getEnv("DB_PASSWORD", "bloodlink_secret_2026"),
		DBName:             getEnv("DB_NAME", "bloodlink_db"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:     getEnvInt("DB_MAX_OPEN_CONNS", 40),
		DBMaxIdleConns:     getEnvInt("DB_MAX_IDLE_CONNS", 20),
		DBConnMaxLifetimeM: getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 30),
		JWTSecret:          getEnv("JWT_SECRET", "bloodlink_minsa_jwt_secret_key_2026_peru_red_nacional"),
		JWTExpirationHours: getEnvInt("JWT_EXPIRATION_HOURS", 8),
		AESKeyHex:          getEnv("AES_KEY_HEX", "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		CORSAllowedOrigins: strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:4200,http://localhost:8080"), ","),
		SOAPEndpointURL:    soapURL,
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Lima",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName, c.DBPort, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}
