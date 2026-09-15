package config

import "os"

type Config struct {
	Port        string
	PostgresDSN string
	RedisURI    string
	RabbitMQURI string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8081"),
		PostgresDSN: getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=order_db port=5432 sslmode=disable"),
		RedisURI:    getEnv("REDIS_URI", "localhost:6379"),
		RabbitMQURI: getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
