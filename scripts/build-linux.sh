#!/bin/bash
# Bash script for building Linux binaries
# Usage: ./scripts/build-linux.sh

set -e  # Exit on error

# Read version from VERSION file
VERSION=$(cat VERSION | tr -d '[:space:]')

BINARY="compress"
OUTPUT_DIR="dist"

echo "Building compress $VERSION for Linux..."

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Build for Linux amd64
echo ""
echo "Building Linux amd64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o "$OUTPUT_DIR/${BINARY}-linux-amd64" ./cmd
if [ $? -ne 0 ]; then
    echo "Failed to build Linux amd64"
    exit 1
fi
chmod +x "$OUTPUT_DIR/${BINARY}-linux-amd64"

# Build for Linux arm64
echo "Building Linux arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o "$OUTPUT_DIR/${BINARY}-linux-arm64" ./cmd
if [ $? -ne 0 ]; then
    echo "Failed to build Linux arm64"
    exit 1
fi
chmod +x "$OUTPUT_DIR/${BINARY}-linux-arm64"

echo ""
echo "Packaging Linux releases..."

# Package Linux amd64
echo "Packaging Linux amd64..."
TEMP_DIR="$OUTPUT_DIR/temp-linux-amd64"
mkdir -p "$TEMP_DIR"
cp "$OUTPUT_DIR/${BINARY}-linux-amd64" "$TEMP_DIR/${BINARY}"
[ -f "LICENSE" ] && cp "LICENSE" "$TEMP_DIR/"
[ -f "README.md" ] && cp "README.md" "$TEMP_DIR/"
[ -f "config.yaml.example" ] && cp "config.yaml.example" "$TEMP_DIR/"
tar -czf "$OUTPUT_DIR/${BINARY}-${VERSION}-linux-amd64.tar.gz" -C "$TEMP_DIR" .
rm -rf "$TEMP_DIR"

# Package Linux arm64
echo "Packaging Linux arm64..."
TEMP_DIR="$OUTPUT_DIR/temp-linux-arm64"
mkdir -p "$TEMP_DIR"
cp "$OUTPUT_DIR/${BINARY}-linux-arm64" "$TEMP_DIR/${BINARY}"
[ -f "LICENSE" ] && cp "LICENSE" "$TEMP_DIR/"
[ -f "README.md" ] && cp "README.md" "$TEMP_DIR/"
[ -f "config.yaml.example" ] && cp "config.yaml.example" "$TEMP_DIR/"
tar -czf "$OUTPUT_DIR/${BINARY}-${VERSION}-linux-arm64.tar.gz" -C "$TEMP_DIR" .
rm -rf "$TEMP_DIR"

echo ""
echo "Generating SHA256 checksums..."

# Generate SHA256 checksums
sha256sum "$OUTPUT_DIR/${BINARY}-${VERSION}-linux-amd64.tar.gz" | awk '{print $1}' > "$OUTPUT_DIR/${BINARY}-${VERSION}-linux-amd64.tar.gz.sha256"
sha256sum "$OUTPUT_DIR/${BINARY}-${VERSION}-linux-arm64.tar.gz" | awk '{print $1}' > "$OUTPUT_DIR/${BINARY}-${VERSION}-linux-arm64.tar.gz.sha256"

echo ""
echo "Build complete! Artifacts:"
echo "  - $OUTPUT_DIR/${BINARY}-${VERSION}-linux-amd64.tar.gz"
echo "  - $OUTPUT_DIR/${BINARY}-${VERSION}-linux-amd64.tar.gz.sha256"
echo "  - $OUTPUT_DIR/${BINARY}-${VERSION}-linux-arm64.tar.gz"
echo "  - $OUTPUT_DIR/${BINARY}-${VERSION}-linux-arm64.tar.gz.sha256"
echo ""
echo "Done!"
