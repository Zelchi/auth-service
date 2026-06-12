dev: 
	air run -c .air.toml
coverage:
	go test ./...
build:
	go build -o ./bin ./cmd/main.go
start: 
	go run ./cmd/main.go