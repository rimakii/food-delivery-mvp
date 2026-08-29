.PHONY: up down lint lint-docker test test-integration test-e2e migrate-up migrate-down

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

test-integration:
	docker compose -f docker-compose.e2e.yaml up -d postgres-e2e
	docker compose -f docker-compose.e2e.yaml run --rm migrate-e2e
	TEST_DATABASE_URL="postgres://avito:avito@localhost:5433/avito_kitchen_e2e?sslmode=disable" go test -tags=integration -count=1 ./...; \
	status=$$?; \
	docker compose -f docker-compose.e2e.yaml down -v; \
	exit $$status

test-e2e:
	docker compose -f docker-compose.e2e.yaml up -d --build api-e2e
	E2E_API_URL="http://localhost:8090" go test -tags=e2e -count=1 ./e2e/...; \
	status=$$?; \
	docker compose -f docker-compose.e2e.yaml down -v; \
	exit $$status

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1
