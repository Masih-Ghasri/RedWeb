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

	carthttp "github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/adapter/in/http"
	cacheRedis "github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/adapter/out/cache/redis"
	persistenceRedis "github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/adapter/out/persistence/redis"
	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/config"
	"github.com/Masih-Ghasri/RedWeb/services/cart-service/internal/core/services"
)

func main() {
	cfg := config.Load()

	// 1. Redis
	// cart-service has exactly one dependency, so failing fast here beats
	// serving 500s on every request while a bad address sits unnoticed.
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisURI})
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		pingCancel()
		log.Fatalf("failed to connect to redis at %s: %v", cfg.RedisURI, err)
	}
	pingCancel()
	defer rdb.Close()

	// 2. Wiring (ports → adapters → use cases → handlers)
	// The TTL is applied on every Save, which is what sweeps abandoned carts.
	ttl := time.Duration(cfg.CartTTLHrs) * time.Hour
	cartRepo := persistenceRedis.NewRedisCartRepository(rdb, ttl)
	productReader := cacheRedis.NewRedisProductReader(rdb)

	cartService := services.NewCartService(cartRepo, productReader)
	cartHandler := carthttp.NewCartHandler(cartService)

	// 3. HTTP server
	mux := http.NewServeMux()
	carthttp.RegisterRoutes(mux, cartHandler)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Cart Service starting on port %s (cart TTL %s)", cfg.Port, ttl)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// 4. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("Server exiting")
}
