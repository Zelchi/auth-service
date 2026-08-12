SHELL := /bin/sh

.PHONY: help install dev dev-backend dev-frontend typecheck build check start preview clean

help:
	@printf '%s\n' \
		'make install    Instala as dependências do frontend e backend' \
		'make dev        Inicia o backend e o frontend de desenvolvimento' \
		'make typecheck  Verifica os tipos do frontend e backend' \
		'make build      Gera o build de produção' \
		'make check      Executa typecheck e build' \
		'make start      Inicia o build de produção existente' \
		'make preview    Inicia o preview do frontend' \
		'make clean      Remove apenas os artefatos gerados'

install:
	yarn --cwd frontend install
	go -C backend mod download

dev: build
	+$(MAKE) --jobs=2 dev-backend dev-frontend

dev-backend:
	go -C backend run ./cmd/main.go

dev-frontend:
	yarn --cwd frontend dev --configLoader runner --host 127.0.0.1 --port 5173 --strictPort

typecheck:
	yarn --cwd frontend tsc -b
	go -C backend vet ./...

build:
	yarn --cwd frontend build
	rm -rf backend/cmd/dist
	cp -r frontend/dist backend/cmd/dist
	mkdir -p backend/bin
	go -C backend build -ldflags="-s -w" -o ./bin/server ./cmd/main.go

check: typecheck build

start:
	cd backend && ./bin/server

preview:
	yarn --cwd frontend preview

clean:
	rm -rf frontend/dist backend/cmd/dist backend/bin
