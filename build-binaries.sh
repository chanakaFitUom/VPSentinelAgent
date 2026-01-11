#!/bin/bash
# Build VPSentinel Agent binaries for all platforms

set -e

echo "Building VPSentinel Agent binaries..."

# Create dist directory
mkdir -p dist

# Build Linux AMD64
echo "Building Linux AMD64..."
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo 'dev')" -o dist/vpsentinel-agent-linux-amd64 ./main.go

# Build Linux ARM64
echo "Building Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo 'dev')" -o dist/vpsentinel-agent-linux-arm64 ./main.go

# Build macOS AMD64
echo "Building macOS AMD64..."
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo 'dev')" -o dist/vpsentinel-agent-darwin-amd64 ./main.go

# Build macOS ARM64 (Apple Silicon)
echo "Building macOS ARM64..."
GOOS=darwin GOARCH=arm64 go build -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo 'dev')" -o dist/vpsentinel-agent-darwin-arm64 ./main.go

# Build Windows AMD64
echo "Building Windows AMD64..."
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.Version=$(git describe --tags --always 2>/dev/null || echo 'dev')" -o dist/vpsentinel-agent-windows-amd64.exe ./main.go

echo "✅ All binaries built successfully!"
echo "Binaries are in: dist/"
ls -lh dist/
