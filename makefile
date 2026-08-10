SHELL := /bin/sh

.PHONY: dev build start clean

dev:
	@if [ ! -f backend/cmd/dist/index.html ]; then \
		$(MAKE) build; \
	fi
	@set -e; \
	backend_pid=; frontend_pid=; \
	cleanup() { \
		[ -z "$$backend_pid" ] || pkill -TERM -P "$$backend_pid" 2>/dev/null || true; \
		[ -z "$$frontend_pid" ] || pkill -TERM -P "$$frontend_pid" 2>/dev/null || true; \
		[ -z "$$backend_pid" ] || kill -TERM "$$backend_pid" 2>/dev/null || true; \
		[ -z "$$frontend_pid" ] || kill -TERM "$$frontend_pid" 2>/dev/null || true; \
	}; \
	trap cleanup INT TERM EXIT; \
	(cd backend && go run ./cmd/main.go) & backend_pid=$$!; \
	(cd frontend && yarn dev --configLoader runner --host 127.0.0.1 --port 5173 --strictPort) & frontend_pid=$$!; \
	wait $$backend_pid $$frontend_pid

build:
	cd frontend && yarn build
	rm -rf backend/cmd/dist
	cp -r frontend/dist backend/cmd/dist
	cd backend && go build -ldflags="-s -w" -o ./bin/server ./cmd/main.go

start: build
	cd backend && go run ./cmd/main.go

clean:
	rm -rf frontend/dist backend/cmd/dist backend/bin
