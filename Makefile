swagger:
	swag init -g cmd/api/main.go -o docs

build:
	go build -o bin/api cmd/api/main.go

run: build
	bin/api

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

seed: build
	bin/api -seed

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

.PHONY: swagger build run migrate-up migrate-down migrate-step-up migrate-step-down migrate-create migrate-status migrate-force seed docker-run docker-stop