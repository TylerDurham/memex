# List available recipes
default:
    @just --list

# Build the memex binary into ./bin
build:
    go build -o bin/memex ./cmd

# Run the test suite
test:
    go test ./...

# Run the test suite with verbose output
test-v:
    go test -v ./...

# gofmt and go vet, then run tests
check: fmt vet test

# Report files that aren't gofmt-clean
fmt:
    gofmt -l .

# Run go vet
vet:
    go vet ./...

# Remove build artifacts
clean:
    rm -rf bin
