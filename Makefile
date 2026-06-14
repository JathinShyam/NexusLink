.PHONY: run build test docker-up docker-down migrate-up migrate-down health

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

docker-up:
	docker compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker compose -f deploy/docker-compose.yml down

docker-logs:
	docker compose -f deploy/docker-compose.yml logs -f api

migrate-up:
	@chmod +x scripts/migrate.sh
	./scripts/migrate.sh up

migrate-down:
	@chmod +x scripts/migrate.sh
	./scripts/migrate.sh down

health:
	curl -sf http://localhost:8080/health
