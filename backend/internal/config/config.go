package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr                string
	AppEnv                  string
	MongoURI                string
	MongoDB                 string
	RedisAddr               string
	RedisPassword           string
	RedisDB                 int
	JWTSecret               string
	CORSOrigins             []string
	SnapshotDir             string
	VAPIDPublicKey          string
	VAPIDPrivateKey         string
	VAPIDSubject            string
	FirebaseProjectID       string
	AppleBundleID           string
	AppleIAPProductID       string
	AppleIAPIssuerID        string
	AppleIAPKeyID           string
	AppleIAPPrivateKey      string
	AppleIAPPrivateKeyPath  string
	AppleIAPEnvironment     string
	AppleIAPAllowUnverified bool
}

func Load() Config {
	return Config{
		HTTPAddr:                env("HTTP_ADDR", ":8080"),
		AppEnv:                  env("APP_ENV", "development"),
		MongoURI:                env("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:                 env("MONGO_DB", "random_match"),
		RedisAddr:               env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:           env("REDIS_PASSWORD", ""),
		RedisDB:                 envInt("REDIS_DB", 0),
		JWTSecret:               env("JWT_SECRET", "dev-secret"),
		CORSOrigins:             split(env("CORS_ORIGINS", "http://localhost:5173,http://localhost:5174,http://localhost,http://127.0.0.1:5173,http://127.0.0.1:5174,http://127.0.0.1")),
		SnapshotDir:             env("SNAPSHOT_DIR", "./snapshots"),
		VAPIDPublicKey:          env("VAPID_PUBLIC_KEY", ""),
		VAPIDPrivateKey:         env("VAPID_PRIVATE_KEY", ""),
		VAPIDSubject:            env("VAPID_SUBJECT", "mailto:admin@example.com"),
		FirebaseProjectID:       env("FIREBASE_PROJECT_ID", ""),
		AppleBundleID:           env("APPLE_BUNDLE_ID", "com.leon456.randommatch"),
		AppleIAPProductID:       env("APPLE_IAP_PRODUCT_ID", "premium_monthly"),
		AppleIAPIssuerID:        env("APPLE_IAP_ISSUER_ID", ""),
		AppleIAPKeyID:           env("APPLE_IAP_KEY_ID", ""),
		AppleIAPPrivateKey:      env("APPLE_IAP_PRIVATE_KEY", ""),
		AppleIAPPrivateKeyPath:  env("APPLE_IAP_PRIVATE_KEY_PATH", ""),
		AppleIAPEnvironment:     env("APPLE_IAP_ENVIRONMENT", "Production"),
		AppleIAPAllowUnverified: envBool("APPLE_IAP_ALLOW_UNVERIFIED", false),
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

func envBool(key string, fallback bool) bool {
	value := strings.ToLower(env(key, ""))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
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
