# ==============================================================================
# Stage 1: Build Go Binary
# ==============================================================================
FROM golang:1.26-alpine AS builder

ENV GOTOOLCHAIN=auto
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -ldflags="-s -w" -o /app/bin/bloodlink-api ./cmd/api/main.go && \
    go build -ldflags="-s -w" -o /app/bin/bloodlink-loadtest ./cmd/loadtest/main.go

# ==============================================================================
# Stage 2: Minimal Secure Runtime
# ==============================================================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

ENV TZ=America/Lima
ENV GIN_MODE=release
ENV PORT=8080

WORKDIR /app

COPY --from=builder /app/bin/bloodlink-api /app/bloodlink-api
COPY --from=builder /app/bin/bloodlink-loadtest /app/bloodlink-loadtest

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/bloodlink-api"]
