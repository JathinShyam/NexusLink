.PHONY: run build test docker-up docker-down migrate-up migrate-down health demo
.PHONY: logs logs-api logs-postgres logs-migrate logs-all
.PHONY: logs-tail logs-tail-api logs-tail-postgres logs-tail-migrate logs-tail-all

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

demo:
	@chmod +x scripts/demo-shorten.sh
	./scripts/demo-shorten.sh

logs: logs-api

logs-api:
	@chmod +x scripts/tail-logs.sh
	@FOLLOW=0 ./scripts/tail-logs.sh api 50

logs-postgres:
	@chmod +x scripts/tail-logs.sh
	@FOLLOW=0 ./scripts/tail-logs.sh postgres 50

logs-migrate:
	@chmod +x scripts/tail-logs.sh
	@FOLLOW=0 ./scripts/tail-logs.sh migrate 50

logs-all:
	@chmod +x scripts/tail-logs.sh
	@FOLLOW=0 ./scripts/tail-logs.sh all 50

logs-tail: logs-tail-api

logs-tail-api:
	@chmod +x scripts/tail-logs.sh
	@./scripts/tail-logs.sh api 50

logs-tail-postgres:
	@chmod +x scripts/tail-logs.sh
	@./scripts/tail-logs.sh postgres 50

logs-tail-migrate:
	@chmod +x scripts/tail-logs.sh
	@./scripts/tail-logs.sh migrate 50

logs-tail-all:
	@chmod +x scripts/tail-logs.sh
	@./scripts/tail-logs.sh all 50
