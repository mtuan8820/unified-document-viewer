# Same default as internal/config; override with DATABASE_URL=... make migrate
DATABASE_URL ?= postgres://app:app@localhost:5432/documents?sslmode=disable

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

migrate:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

.PHONY: run up down test generate mock-sales mock-service migrate migrate-down
