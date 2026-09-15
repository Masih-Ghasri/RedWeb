package config

import "os"

type Config struct {
	PostgresDSN string
	RabbitMQURI string
}

func Load() *Config {
	return &Config{
		PostgresDSN: getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=shipping_db port=5432 sslmode=disable"),
		RabbitMQURI: getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
