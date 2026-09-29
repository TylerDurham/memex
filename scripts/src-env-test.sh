#!/usr/bin/env bash

# SETS UP THE LOCAL ENV FOR TESTING

PROJECT_ROOT_DIR=$(git rev-parse --show-toplevel)

export MEMEX_CONFIG_DIR="$PROJECT_ROOT_DIR/tmp/.config/memex"

echo "TEST ENVIRONMENT INITIALIZED:"
echo ""
echo " - ENV[MEMEX_CONFIG_DIR] = '$MEMEX_CONFIG_DIR'"
echo ""
