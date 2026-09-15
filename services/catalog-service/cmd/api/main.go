package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/adapter/out/persistence/mongodb"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/config"
	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/services"

	cataloghttp "github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/adapter/in/http"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Setup Database (MongoDB)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("failed to connect to mongodb: %v", err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			log.Fatalf("failed to disconnect mongodb: %v", err)
		}
	}(client, context.Background())
	db := client.Database(cfg.DBName)

	repo := mongodb.NewMongoProductRepository(db)
	service := services.NewProductService(repo)
	handler := cataloghttp.NewProductHandler(service) // alias for internal/adapter/in/http

	mux := http.NewServeMux()
	cataloghttp.RegisterRoutes(mux, handler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("Catalog Service starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

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
