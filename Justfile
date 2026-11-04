# List available recipes
default:
    @just --list

# BUILD

# TESTS
test:
    gotestsum --format testdox ./...

test-verbose:
    gotestsum --format standard-verbose ./...

test-byname:
    gotestsum --format testname ./... -v
