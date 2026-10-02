package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port       string
	RedisURI   string
	CartTTLHrs int
}

func Load() *Config {
	return &Config{
		Port:       getEnv("PORT", "8082"),
		RedisURI:   getEnv("REDIS_URI", "localhost:6379"),
		CartTTLHrs: getEnvInt("CART_TTL_HOURS", defaultCartTTLHours),
	}
}

// defaultCartTTLHours is seven days: long enough for a shopper to come back
// to a cart they abandoned mid-session, short enough that carts do not
// accumulate forever in Redis.
const defaultCartTTLHours = 24 * 7

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvInt parses an integer variable. A value that is present but
// unparsable falls back to the default with a warning rather than crashing
// the service at startup — a typo in an env var should not take the pod down.
func getEnvInt(key string, fallback int) int {
	raw, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		log.Printf("warning: %s=%q is not an integer, using default %d", key, raw, fallback)
		return fallback
	}
	return value
}
