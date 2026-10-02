FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /src
COPY services/order-service/go.mod services/order-service/go.sum ./
RUN go mod download

COPY services/order-service/ ./
RUN go build -ldflags="-w -s" -o /bin/order-service ./cmd/api

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

USER appuser
WORKDIR /app
COPY --from=builder /bin/order-service .

EXPOSE 8081
CMD ["./order-service"]