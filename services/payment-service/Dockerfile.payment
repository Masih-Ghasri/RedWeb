FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-w -s" -o /bin/payment-service ./cmd/worker/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
RUN adduser -D -g '' appuser
USER appuser

WORKDIR /app
COPY --from=builder /bin/payment-service .

CMD ["./payment-service"]