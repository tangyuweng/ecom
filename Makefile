# ==================== Build Targets ====================
.PHONY:	swagger	build	run

swagger:
	swag init -g cmd/api/main.go -o docs

build:
	go build -o bin/api cmd/api/main.go

run: build
	bin/api

# ==================== Migration Targets ====================
.PHONY: migrate-up migrate-down migrate-step-up migrate-step-down
.PHONY: migrate-create migrate-status migrate-force

migrate-up: build
	bin/api -migrate-up

migrate-down: build
	bin/api -migrate-down

migrate-step-up: build
	@read -p "How many steps to migrate up? " steps; \
	bin/api -migrate-steps=$$steps

migrate-step-down: build
	@read -p "How many steps to rollback? " steps; \
	bin/api -migrate-steps=-$$steps

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir internal/infrastructure/mysql/migrations -seq $$name

migrate-status: build
	bin/api -migrate-status

migrate-force: build
	@read -p "Force migration to version (current dirty version): " version; \
	bin/api -migrate-force=$$version

# ==================== Data Targets ====================
.PHONY: seed

seed: build
	bin/api -seed

# ==================== Docker Targets ====================
.PHONY: docker-run docker-stop test-db-up test-db-down

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

test-db-up:
	docker-compose --profile test up -d mysql-test

test-db-down:
	docker-compose --profile test down

# ==================== Test Targets ====================
.PHONY: test test-unit test-integration test-coverage test-coverage-report

test:
	go test -v ./...

test-unit:
	go test -v ./internal/application/...

test-integration:
	go test -v ./internal/infrastructure/...

test-coverage:
	go test -cover ./...

test-coverage-report:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out