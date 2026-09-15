package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	brokerIn "github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/adapter/in/broker/rabbitmq"
	brokerOut "github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/adapter/out/broker/rabbitmq"
	postgresDb "github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/adapter/out/persistence/postgres"
	redisCache "github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/adapter/out/redis"
	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/config"
	"github.com/Masih-Ghasri/RedWeb/services/payment-service/internal/core/services"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Database
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	db.AutoMigrate(&postgresDb.PaymentModel{})

	// 2. Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURI})

	// 3. RabbitMQ Connection
	conn, err := amqp.Dial(cfg.RabbitMQURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// 4. Dependency Injection
	paymentRepo := postgresDb.NewPostgresPaymentRepository(db)
	idempotencyRepo := redisCache.NewRedisIdempotencyRepository(rdb)
	eventPublisher := brokerOut.NewRabbitMQPublisher(ch)

	paymentService := services.NewPaymentService(paymentRepo, idempotencyRepo, eventPublisher)

	// 5. Setup Consumer
	consumer := brokerIn.NewOrderCreatedConsumer(ch, paymentService)
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	// 6. Keep Process Alive
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Payment Worker shutting down gracefully...")
}
