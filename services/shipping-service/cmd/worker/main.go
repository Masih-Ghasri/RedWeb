package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	brokerIn "github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/adapter/in/broker/rabbitmq"
	postgresDb "github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/adapter/out/persistence/postgres"
	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/config"
	"github.com/Masih-Ghasri/RedWeb/services/shipping-service/internal/core/services"
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
	db.AutoMigrate(&postgresDb.ShipmentModel{})

	// 2. RabbitMQ Connection
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

	// 3. Dependency Injection
	shippingRepo := postgresDb.NewPostgresShippingRepository(db)
	shippingService := services.NewShippingService(shippingRepo)

	// 4. Setup Consumer
	consumer := brokerIn.NewPaymentProcessedConsumer(ch, shippingService)
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start shipping consumer: %v", err)
	}

	// 5. Keep Process Alive
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shipping Worker shutting down gracefully...")
}
