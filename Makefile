run:
	go run ./cmd/server

up:
	docker compose up -d

down:
	docker compose down

test:
	go test -race ./...

.PHONY: run up down test
