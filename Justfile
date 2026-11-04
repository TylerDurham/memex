# List available recipes
default:
    @just --list

# BUILD

# TESTS
test:
    go test ./...

test-verbose:
    go test ./... -v
