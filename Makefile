APP_NAME := welyo-recruitment-task
PKG= ./cmd/server

GOFILES=$(shell find . -name "*.go")
FMT=$(GOFILES)

.PHONY: fmt fmt-check lint security vet check clean test build run docker-build docker-run

fmt:
	@echo "Running go fmt..."
	gofmt -w $(FMT)

fmt-check:
	@echo "Checking formatting..."
	@if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted:"; \
		gofmt -s -l .; \
		exit 1; \
	fi

lint:
	@echo "Running go lint..."
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	golangci-lint run ./...

security:
	@echo "Running security scan..."
	@if ! command -v gosec &> /dev/null; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...

vet:
	@echo "Running go vet..."
	go vet ./...

check: fmt-check vet lint security test
	@echo "All checks passed!"

clean:
	@echo "Cleaning up..."
	rm -f $(APP_NAME)
	
test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) $(PKG)

run:
	@echo "Running $(APP_NAME)..."
	SERVER_HELLO="HELLO!" PORT=8080 go run $(PKG)

docker-build:
	@echo "Building docker image..."
	docker build -t $(APP_NAME):local .

docker-run: docker-build
	@echo "Running docker container..."
	docker run -p 8082:8082 -e SERVER_HELLO="HELLO AGAIN!" -e PORT=8082 $(APP_NAME):local