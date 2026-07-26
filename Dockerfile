# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.21.4-alpine AS build

WORKDIR /src

# Cache dependencies first for faster incremental builds.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Static binary; frontend assets are embedded via go:embed.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /out/qrforge ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.19

# wget is used by the container HEALTHCHECK.
RUN apk add --no-cache ca-certificates wget \
    && addgroup -S app && adduser -S -G app app

WORKDIR /app
COPY --from=build /out/qrforge /app/qrforge

ENV PORT=8080
EXPOSE 8080

USER app

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/app/qrforge"]
