FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY pkg/ ./pkg/
COPY services/catalog-service/ ./services/catalog-service/


WORKDIR /app/services/catalog-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o catalog-service ./cmd/api

# مرحله Production (حجم بسیار پایین و امن)
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/services/catalog-service/catalog-service .

EXPOSE 8080
CMD ["./catalog-service"]