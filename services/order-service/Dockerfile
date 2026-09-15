FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/services/order-service
RUN CGO_ENABLED=0 GOOS=linux go build -o order-service ./cmd/api

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/services/order-service/order-service .

EXPOSE 8081
CMD ["./order-service"]