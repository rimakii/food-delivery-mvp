.PHONY: up down lint test migrate-up migrate-down

up:
	docker compose up --build

down:
	docker compose down -v

lint:
	golangci-lint run ./...

test:
	go test ./...

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1
