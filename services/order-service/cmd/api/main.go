package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	orderHttp "github.com/Masih-Ghasri/RedWeb/services/order-service/internal/adapter/in/http"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/adapter/out/broker/rabbitmq"
	redisCache "github.com/Masih-Ghasri/RedWeb/services/order-service/internal/adapter/out/cache/redis"
	postgresDb "github.com/Masih-Ghasri/RedWeb/services/order-service/internal/adapter/out/persistence/postgres"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/config"
	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/services"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Setup Postgres
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	db.AutoMigrate(&postgresDb.OrderModel{}, &postgresDb.OrderItemModel{}, &postgresDb.OutboxEventModel{})

	// 2. Setup Redis
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURI})

	// 3. Setup RabbitMQ
	amqpPublisher, err := rabbitmq.NewRabbitMQPublisher(cfg.RabbitMQURI)
	if err != nil {
		log.Fatalf("Failed to setup RabbitMQ Publisher: %v", err)
	}
	defer amqpPublisher.Close()

	// 4. Dependency Injection (Wiring)
	orderRepo := postgresDb.NewPostgresOrderRepository(db)
	outboxRepo := postgresDb.NewPostgresOutboxRepository(db)
	inventoryCache := redisCache.NewRedisInventoryCache(rdb)

	orderService := services.NewOrderService(orderRepo, inventoryCache)
	orderHandler := orderHttp.NewOrderHandler(orderService)

	// 5. Setup & Start Outbox Relay Worker (Background Process)
	relayWorker := services.NewOutboxRelayWorker(outboxRepo, amqpPublisher, 2*time.Second)
	go relayWorker.Start(ctx)

	// 6. Setup Router & Server
	mux := http.NewServeMux()
	orderHttp.RegisterRoutes(mux, orderHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("Order Service API running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// 7. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down services gracefully...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP Server shutdown error: %v", err)
	}
	log.Println("Service exited properly")
}
