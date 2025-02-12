.PHONY: test coverage build clean lint

# Build the application
build:
	go build -o bin/urlsluice

# Run tests
test:
	go test -race ./...

# Run tests with coverage
coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Run linter
lint:
	golangci-lint run

# Install development dependencies
dev-deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest 