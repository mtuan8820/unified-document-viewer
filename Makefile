run:
	go run ./cmd/server

up:
	docker compose up -d

down:
	docker compose down

test:
	go test -race ./...

generate:
	go generate ./...

mock-sales:
	go run ./mocks/sales

mock-service:
	go run ./mocks/service

.PHONY: run up down test generate mock-sales mock-service
