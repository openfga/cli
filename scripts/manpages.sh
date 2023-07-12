#!/bin/sh
# script/manpages.sh
set -e

OUTPUT_DIR=manpages
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

go run . man | gzip -c -9 > "$OUTPUT_DIR"/fga.1.gz