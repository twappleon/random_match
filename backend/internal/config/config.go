package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr      string
	MongoURI      string
	MongoDB       string
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
	CORSOrigins   []string
}

func Load() Config {
	return Config{
		HTTPAddr:      env("HTTP_ADDR", ":8080"),
		MongoURI:      env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:       env("MONGO_DB", "random_match"),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       envInt("REDIS_DB", 0),
		JWTSecret:     env("JWT_SECRET", "dev-secret"),
		CORSOrigins:   split(env("CORS_ORIGINS", "http://localhost:5173")),
	}
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil {
		return fallback
	}
	return value
}

func split(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			out = append(out, item)
		}
	}
	return out
}
