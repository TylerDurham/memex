# Default directory for zsh completions: zinit's completions dir, which is on $fpath
zsh_completions_dir := env("XDG_DATA_HOME", env("HOME") / ".local/share") / "zinit/completions"

# Config directory used by test recipes, so tests never touch your real memex config.
# Matches scripts/src-env-test.sh; source that script to use it in an interactive shell.
test_config_dir := justfile_directory() / "tmp/.config/memex"

# List available recipes
default:
    @just --list

# Build the memex binary into ./bin
build:
    go build -o bin/memex ./cmd

# Build memex and symlink ~/.local/bin/memex to ./bin/memex, so later builds update it in place
link: build
    #!/usr/bin/env bash
    set -euo pipefail
    src={{ quote(justfile_directory() / "bin" / "memex") }}
    dst="$HOME/.local/bin/memex"
    if [[ -e "$dst" && ! -L "$dst" ]]; then
        echo "error: $dst exists and isn't a symlink; not replacing it" >&2
        exit 1
    fi
    mkdir -p "$(dirname "$dst")"
    ln -sfn "$src" "$dst"
    echo "linked $dst -> $src"
    case ":$PATH:" in
        *":$HOME/.local/bin:"*) ;;
        *) echo "note: ~/.local/bin isn't on your PATH" >&2 ;;
    esac

# Remove the ~/.local/bin/memex symlink, if it points at this repo
unlink:
    #!/usr/bin/env bash
    set -euo pipefail
    src={{ quote(justfile_directory() / "bin" / "memex") }}
    dst="$HOME/.local/bin/memex"
    if [[ -L "$dst" && "$(readlink "$dst")" == "$src" ]]; then
        rm "$dst"
        echo "removed $dst"
    else
        echo "$dst isn't a symlink to $src; leaving it alone"
    fi

# Install zsh completions into a directory on $fpath (default: zinit's completions dir)
completions-zsh dir=zsh_completions_dir: build link
    #!/usr/bin/env bash
    set -euo pipefail
    dir={{ quote(dir) }}
    mkdir -p "$dir"
    ./bin/memex completion zsh > "$dir/_memex"
    echo "wrote $dir/_memex"
    # compinit caches completions in .zcompdump; remove it so _memex is picked up.
    rm -f "${ZDOTDIR:-$HOME}"/.zcompdump*
    echo "cleared .zcompdump; run 'exec zsh' to load the completions"

# Remove the zsh completions installed by completions-zsh
completions-zsh-clean dir=zsh_completions_dir: unlink
    #!/usr/bin/env bash
    set -euo pipefail
    file={{ quote(dir) }}/_memex
    if [[ ! -e "$file" ]]; then
        echo "$file doesn't exist; nothing to remove"
        exit 0
    fi
    # Only remove a file that is memex's completion script.
    if [[ "$(head -n 1 "$file")" != "#compdef memex" ]]; then
        echo "error: $file isn't a memex completion script; not removing it" >&2
        exit 1
    fi
    rm "$file"
    echo "removed $file"
    # Drop the cached completions so zsh forgets _memex.
    rm -f "${ZDOTDIR:-$HOME}"/.zcompdump*
    echo "cleared .zcompdump; run 'exec zsh' to reload completions"

# Run the test suite
test:
    MEMEX_CONFIG_DIR={{ quote(test_config_dir) }} go test ./...

# Run the test suite with verbose output
test-v:
    MEMEX_CONFIG_DIR={{ quote(test_config_dir) }} go test -v ./...

# Run the tests and write an HTML coverage report to coverage.html
cover:
    MEMEX_CONFIG_DIR={{ quote(test_config_dir) }} go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "wrote coverage.html"

# Report files that aren't gofmt-clean
fmt:
    gofmt -l .

# Run go vet
vet:
    go vet ./...

# gofmt check, go vet, then run tests
check: fmt vet test

# Remove build artifacts and coverage output
clean:
    rm -rf bin coverage.out coverage.html
