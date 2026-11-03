# Config directory used by test recipes, so tests never touch your real memex config.
# Matches scripts/src-env-test.sh; source that script to use it in an interactive shell.
test_config_dir := justfile_directory() / "tmp/.config/memex"

# List available recipes
default:
    @just --list

# Run tests w/ENV vars
test:
    MEMEX_CONFIG_DIR={{ quote(test_config_dir) }} go test -v ./...
