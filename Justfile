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

# Build the Obsidian plugin (installs npm dependencies on first run)
[working-directory: 'obsidian-plugin']
plugin-build:
    @[ -d node_modules ] || npm ci --no-audit --no-fund
    npm run build

# Run the Obsidian plugin's tests
[working-directory: 'obsidian-plugin']
plugin-test:
    @[ -d node_modules ] || npm ci --no-audit --no-fund
    npm test

# Build the Obsidian plugin and symlink it into a vault (default: $OBSIDIAN_VAULT)
plugin-install vault=env("OBSIDIAN_VAULT", ""):
    #!/usr/bin/env bash
    set -euo pipefail
    vault={{ quote(vault) }}
    vault="${vault/#\~/$HOME}"
    if [[ -z "$vault" ]]; then
        echo "usage: just plugin-install <vault dir>   (or set OBSIDIAN_VAULT)" >&2
        exit 1
    fi
    if [[ ! -d "$vault/.obsidian" ]]; then
        echo "error: '$vault' is not an Obsidian vault (no .obsidian folder)" >&2
        exit 1
    fi
    just --justfile {{ quote(justfile()) }} plugin-build
    src={{ quote(justfile_directory() / "obsidian-plugin") }}
    dir="$vault/.obsidian/plugins/memex-search"
    mkdir -p "$dir"
    for f in main.js manifest.json styles.css; do
        ln -sfn "$src/$f" "$dir/$f"
    done
    echo "linked memex-search into $dir"
    echo "enable it under Settings → Community plugins (or reload Obsidian if it's already enabled)"
