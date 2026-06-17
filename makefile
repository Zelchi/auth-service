.PHONY: dev build start coverage docker up down

coverage:
	cd backend && go test -v -cover ./...

build:
	cd frontend && yarn build
	rm -rf backend/cmd/dist
	cp -r frontend/dist backend/cmd/dist
	cd backend && go build -ldflags="-s -w" -o ./bin/server ./cmd/main.go

start:
	cd backend && go run ./cmd/main.go

docker:
	docker build -t auth-service .

up:
	docker compose up -d --build

down:
	docker compose down
