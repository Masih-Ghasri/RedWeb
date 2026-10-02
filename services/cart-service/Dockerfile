FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build context is the repository root:
#     docker build -f services/cart-service/Dockerfile .
#
# Only this module is copied — not the whole repo — so editing an unrelated
# service does not invalidate the download layer. Deliberately modelled on
# Dockerfile.order (multi-stage, non-root) rather than on
# services/catalog-service/Dockerfile, which is broken: it copies a root
# go.sum that does not exist and runs its runtime stage as root.
WORKDIR /src
COPY services/cart-service/ ./services/cart-service/
WORKDIR /src/services/cart-service

RUN go mod download && \
    go build -ldflags="-w -s" -o /bin/cart-service ./cmd/api

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

USER appuser
WORKDIR /app
COPY --from=builder /bin/cart-service .

EXPOSE 8082
CMD ["./cart-service"]
