build:
	cd frontend && yarn build
	rm -rf backend/cmd/dist
	cp -r frontend/dist backend/cmd/dist
	cd backend && go build -ldflags="-s -w" -o ./bin/server ./cmd/main.go

start: build
	cd backend && go run ./cmd/main.go

clean:
	rm -rf frontend/dist backend/cmd/dist backend/bin