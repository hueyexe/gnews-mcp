.PHONY: build test lint fmt run tidy

build:
	go build -o bin/gnews-mcp .

test:
	go test -race ./...

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...

run:
	go run .

tidy:
	go mod tidy