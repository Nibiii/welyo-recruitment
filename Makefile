APP_NAME := welyo-recruitment-task
PKG= ./cmd/server

GOFILES=$(shell find . -name "*.go" -not -path "*/vendor/*")
FMT=$(GOFILES)

.PHONY: fmt lint test build run

fmt:
	@echo "Running go fmt..."
	gofmt -w $(FMT)

lint:
	@echo "Running go lint..."
	golangci-lint run ./...
	
test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) $(PKG)

run:
	@echo "Running $(APP_NAME)..."
	SERVER_HELLO="HELLO!" PORT=8080 go run $(PKG)