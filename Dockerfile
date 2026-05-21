# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Bog'liqliklarni cache qilish
COPY go.mod go.sum* ./
RUN go mod download

# Manba kodini ko'chirish
COPY . .

# Binary kompilatsiya (statik, kichik)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /app/bin/api ./cmd/api

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/bin/api .
COPY --from=builder /app/migrations ./migrations

USER app

EXPOSE 8000

CMD ["./api"]
