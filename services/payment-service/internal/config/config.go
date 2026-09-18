package config

import "os"

type Config struct {
	PostgresDSN    string
	RedisURI       string
	RabbitMQURI    string
	BankGatewayURL string
}

func Load() *Config {
	return &Config{
		PostgresDSN:    getEnv("POSTGRES_DSN", "host=localhost user=postgres password=postgres dbname=payment_db port=5432 sslmode=disable"),
		RedisURI:       getEnv("REDIS_URI", "localhost:6379"),
		RabbitMQURI:    getEnv("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/"),
		BankGatewayURL: getEnv("BANK_GATEWAY_URL", "http://localhost:9090"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
