run:
	go run ./cmd/server

up:
	docker compose up -d

down:
	docker compose down

test:
	go test -race ./...

generate:
	go generate  ./...

.PHONY: run up down test
