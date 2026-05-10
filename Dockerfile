# EnvGuard — lightweight .env validation
# Usage: docker run --rm -v $(pwd):/workspace ghcr.io/firasmosbehi/envguard:latest validate
FROM alpine:3.19

LABEL org.opencontainers.image.source="https://github.com/firasmosbehi/envguard"
LABEL org.opencontainers.image.description="EnvGuard — validate .env files against schemas"
LABEL org.opencontainers.image.licenses="MIT"

# Install ca-certificates for HTTPS validation (URL format checks)
RUN apk add --no-cache ca-certificates

# Download the latest release binary for linux-amd64
ARG VERSION=0.1.7
ARG TARGETARCH
ARG TARGETOS

RUN wget -q -O /usr/local/bin/envguard \
    "https://github.com/firasmosbehi/envguard/releases/download/v${VERSION}/envguard-${TARGETOS:-linux}-${TARGETARCH:-amd64}" \
    && chmod +x /usr/local/bin/envguard

WORKDIR /workspace

ENTRYPOINT ["envguard"]
CMD ["validate"]
