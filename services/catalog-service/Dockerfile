FROM golang:1.26-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build context is the repository root:
#     docker build -f services/catalog-service/Dockerfile .
#
# Only this module is copied — not the whole repo — so editing an unrelated
# service does not invalidate the download layer. The previous version of this
# file copied a root-level `go.sum` that does not exist in the repository,
# which made every image build fail, and ran the runtime stage as root.
WORKDIR /src
COPY services/catalog-service/ ./services/catalog-service/
WORKDIR /src/services/catalog-service

RUN go mod download && \
    go build -ldflags="-w -s" -o /bin/catalog-service ./cmd/api

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && \
    adduser -D -g '' appuser

USER appuser
WORKDIR /app
COPY --from=builder /bin/catalog-service .

EXPOSE 8080
CMD ["./catalog-service"]