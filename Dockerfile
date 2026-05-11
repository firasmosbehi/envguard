# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o envguard ./cmd/envguard

# Final stage — minimal image with just the binary and CA certs
FROM scratch

LABEL org.opencontainers.image.source="https://github.com/firasmosbehi/envguard"
LABEL org.opencontainers.image.description="EnvGuard — validate .env files against schemas"
LABEL org.opencontainers.image.licenses="MIT"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/envguard /usr/local/bin/envguard

WORKDIR /workspace

ENTRYPOINT ["/usr/local/bin/envguard"]
CMD ["validate"]
