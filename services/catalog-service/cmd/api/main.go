package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	cataloghttp "github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/adapter/in/http"
	brokerOut "github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/adapter/out/broker/rabbitmq"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/adapter/out/persistence/mongodb"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/config"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/services"
)

func main() {
	cfg := config.Load()

	// 1. MongoDB
	// A bounded context keeps a hung dial from stalling startup forever; it
	// is released as soon as the connect attempt returns.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer func(client *mongo.Client) {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("failed to disconnect mongodb: %v", err)
		}
	}(client)

	db := client.Database(cfg.DBName)

	// 1b. Indexes
	// Idempotent, so running on every startup is cheap. Failure degrades
	// listing performance rather than correctness, hence a warning instead
	// of a fatal — but it gets its own deadline so a slow connect cannot eat
	// the budget and skip index creation silently.
	indexCtx, indexCancel := context.WithTimeout(context.Background(), 15*time.Second)
	if err := mongodb.EnsureIndexes(indexCtx, db); err != nil {
		log.Printf("warning: could not create product indexes: %v", err)
	}
	indexCancel()

	// 2. RabbitMQ
	// Publish-only: these are the product events order-service turns into its
	// Redis product cache. Failing here is fatal rather than a warning —
	// otherwise every product write would appear to succeed while never
	// reaching the cache, and order creation would fail with "product not
	// found in cache" for a product that plainly exists.
	rabbitConn, err := amqp.Dial(cfg.RabbitMQURI)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("failed to open RabbitMQ channel: %v", err)
	}
	defer rabbitChan.Close()

	publisher, err := brokerOut.NewRabbitMQPublisher(rabbitChan)
	if err != nil {
		log.Fatalf("failed to initialise event publisher: %v", err)
	}

	// 3. Wiring (ports → adapters → use cases → handlers)
	repo := mongodb.NewMongoProductRepository(db)
	service := services.NewProductService(repo, publisher)
	handler := cataloghttp.NewProductHandler(service)

	// 4. HTTP server
	mux := http.NewServeMux()
	cataloghttp.RegisterRoutes(mux, handler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Catalog Service starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 5. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
