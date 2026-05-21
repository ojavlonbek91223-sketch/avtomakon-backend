.PHONY: run build test migrate-up migrate-down migrate-create lint docker-up docker-down

# Asosiy buyruqlar
run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test -v -race -cover ./...

lint:
	golangci-lint run ./...

# Migratsiya (golang-migrate kerak: https://github.com/golang-migrate/migrate)
DB_URL := postgres://avtomakon:avtomakon_dev_password@localhost:5432/avtomakon?sslmode=disable

migrate-up:
	migrate -path ./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DB_URL)" down 1

migrate-create:
	@read -p "Migration nomi: " name; \
	migrate create -ext sql -dir ./migrations -seq $$name

# Docker
docker-up:
	docker-compose -f ../docker-compose.yml up -d

docker-down:
	docker-compose -f ../docker-compose.yml down

# Code generation (mocks, etc.)
generate:
	go generate ./...
