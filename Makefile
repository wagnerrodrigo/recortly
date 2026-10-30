.PHONY: setup dev api worker web build test clean db-reset

setup:
	@test -f .env || cp .env.example .env
	docker compose up -d
	cd api && go mod download
	cd web && npm install

dev:
	./dev.sh

api:
	cd api && go run ./cmd/api

worker:
	cd api && go run ./cmd/worker

web:
	cd web && npm run dev -- --port 3000

build:
	cd api && go build -o ../bin/api ./cmd/api && go build -o ../bin/worker ./cmd/worker
	cd web && npm run build

test:
	cd api && go test ./...
	cd web && npx nuxt build

clean:
	rm -rf bin web/.output web/.nuxt

db-reset:
	docker compose down -v
	docker compose up -d
