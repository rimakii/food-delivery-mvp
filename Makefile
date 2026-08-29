.PHONY: up down lint lint-docker test migrate-up migrate-down

up:
	docker compose up --build

down:
	docker compose down -v

lint:
	golangci-lint run ./...
	cd restaurant-service && golangci-lint run ./...

lint-docker:
	docker run --rm -v "$(CURDIR):/app" -w /app golangci/golangci-lint:latest golangci-lint run ./...
	docker run --rm -v "$(CURDIR)/restaurant-service:/app" -w /app golangci/golangci-lint:latest golangci-lint run ./...

test:
	go test ./...
	cd restaurant-service && go test ./...

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1
