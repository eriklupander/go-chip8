all: fmt build test

build:
	go build -o bin/go-chip8 cmd/chip8/main.go

run: build
	./bin/go-chip8

test:
	go test ./... -count=1 -v

fmt:
	go fmt ./...