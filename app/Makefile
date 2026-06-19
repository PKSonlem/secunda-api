.PHONY: up down build test test-integration cover lint tidy

up:
	docker compose up -d

down:
	docker compose down -v

build:
	go build ./cmd/server

test:
	go test ./tests/service/... ./tests/handler/... ./internal/...

test-integration:
	go test -tags=integration ./tests/integration/... -v -timeout 5m

cover:
	go test -coverpkg=./internal/... ./tests/service/... ./tests/handler/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
