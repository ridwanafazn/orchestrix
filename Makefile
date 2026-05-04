run-api:
	go run cmd/api/main.go

run-worker:
	go run cmd/worker/main.go

build:
	go build -o bin/api cmd/api/main.go
	go build -o bin/worker cmd/worker/main.go

test:
	go test -v ./...

tidy:
	go mod tidy
