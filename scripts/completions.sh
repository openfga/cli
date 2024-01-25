#!/bin/sh
# scripts/completions.sh
set -e

OUTPUT_DIR=completions
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

for sh in bash zsh fish; do
	go run ./cmd/fga/main.go completion "$sh" >"${OUTPUT_DIR}/fga.$sh"
done
