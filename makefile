dev: 
	cd backend && air run -c .air.toml
coverage:
	cd backend && go test ./...
build:
	cd backend && go build -o ./bin ./cmd/main.go
start: 
	cd backend && go run ./cmd/main.go