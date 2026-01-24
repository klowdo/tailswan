#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOGO="${SCRIPT_DIR}/../logo.jpeg"
OUTPUT_DIR="${SCRIPT_DIR}"
SIZES=(72 96 128 144 152 192 384 512)

if [ ! -f "$LOGO" ]; then
    echo "Error: logo.jpeg not found at $LOGO"
    exit 1
fi

echo "Generating PWA app icons from $LOGO..."

for size in "${SIZES[@]}"; do
    output="${OUTPUT_DIR}/icon-${size}.png"
    echo "Creating icon-${size}.png (${size}x${size})..."
    convert "$LOGO" -resize "${size}x${size}" -background white -gravity center -extent "${size}x${size}" "$output"
done

echo ""
echo "Icon generation complete!"
echo "Generated icons:"
ls -lh "${OUTPUT_DIR}"/icon-*.png
