.PHONY: build run test lint clean dev tui tui-build docker-build docker-run docker-down

BINARY_NAME=brainsentry
BUILD_DIR=./bin

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: build
	$(BUILD_DIR)/$(BINARY_NAME)

dev:
	go run ./cmd/server

tui:
	go run ./cmd/tui

tui-build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-tui ./cmd/tui

cli:
	go run ./cmd/cli

cli-build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-cli ./cmd/cli

test:
	go test ./... -v -race

test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-short:
	go test ./... -short

lint:
	golangci-lint run ./...

vet:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

migrate-up:
	migrate -path internal/repository/postgres/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path internal/repository/postgres/migrations -database "$$DATABASE_URL" down

docker-build:
	docker build -t brainsentry-go .

docker-run:
	docker compose up -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f brainsentry

infra-up:
	docker compose up -d postgres falkordb

infra-down:
	docker compose down postgres falkordb
