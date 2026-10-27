.PHONY: help build test vet tidy up down run-catalog run-order run-cart run-payment run-shipping run-mockbank

help:
	@echo "RedWeb - Microservices Management"
	@echo "Available commands:"
	@echo "  make build         - Build all services"
	@echo "  make test          - Run unit tests across all services"
	@echo "  make vet           - Run go vet static analysis"
	@echo "  make tidy          - Run go mod tidy on all services"
	@echo "  make up            - Start external infrastructure (Postgres, Mongo, Redis, RabbitMQ)"
	@echo "  make down          - Stop external infrastructure"
	@echo "  make run-catalog   - Run Catalog Service (port 8080)"
	@echo "  make run-order     - Run Order Service (port 8081)"
	@echo "  make run-cart      - Run Cart Service (port 8082)"
	@echo "  make run-payment   - Run Payment Worker"
	@echo "  make run-shipping  - Run Shipping Worker"
	@echo "  make run-mockbank  - Run Mock Bank (port 9090)"

build:
	go build ./services/catalog-service/...
	go build ./services/order-service/...
	go build ./services/cart-service/...
	go build ./services/payment-service/...
	go build ./services/shipping-service/...
	go build ./services/mock-bank/...

test:
	go test -v ./services/catalog-service/... ./services/order-service/... ./services/cart-service/...

vet:
	go vet ./services/catalog-service/... ./services/order-service/... ./services/cart-service/... ./services/payment-service/... ./services/shipping-service/... ./services/mock-bank/...

tidy:
	cd services/catalog-service && go mod tidy
	cd services/order-service && go mod tidy
	cd services/cart-service && go mod tidy
	cd services/payment-service && go mod tidy
	cd services/shipping-service && go mod tidy
	cd services/mock-bank && go mod tidy

up:
	docker compose -f docker-compose.external.yml up -d

down:
	docker compose -f docker-compose.external.yml down

run-catalog:
	go run ./services/catalog-service/cmd/api

run-order:
	go run ./services/order-service/cmd/api

run-cart:
	go run ./services/cart-service/cmd/api

run-payment:
	go run ./services/payment-service/cmd/worker

run-shipping:
	go run ./services/shipping-service/cmd/worker

run-mockbank:
	go run ./services/mock-bank/main.go
