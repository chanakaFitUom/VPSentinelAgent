# Building VPSentinel Agent

## Prerequisites

- Go 1.22 or higher
- Internet connection for `go mod download`

## Quick Build

```bash
cd agent
go mod download
go build -o vpsentinel-agent ./main.go
```

## Build Options

### Standard Build
```bash
go build -o vpsentinel-agent ./main.go
```

### Build with Version
```bash
go build -ldflags "-X main.Version=1.0.0" -o vpsentinel-agent ./main.go
```

### Optimized Build (Smaller Binary)
```bash
go build -ldflags "-s -w -X main.Version=1.0.0" -o vpsentinel-agent ./main.go
```

The `-s` flag strips symbol table and debug info, and `-w` omits DWARF symbol table.

### Static Binary (No Dependencies)
```bash
CGO_ENABLED=0 go build -ldflags "-s -w -X main.Version=1.0.0" -o vpsentinel-agent ./main.go
```

Note: gopsutil may have some dependencies that require CGO on certain platforms.

## Cross-Platform Builds

### Linux (amd64)
```bash
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o vpsentinel-agent-linux-amd64 ./main.go
```

### Linux (arm64)
```bash
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o vpsentinel-agent-linux-arm64 ./main.go
```

### macOS (amd64)
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o vpsentinel-agent-darwin-amd64 ./main.go
```

### macOS (arm64 - Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o vpsentinel-agent-darwin-arm64 ./main.go
```

### Windows (amd64)
```bash
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o vpsentinel-agent-windows-amd64.exe ./main.go
```

## Testing Build

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## Development Build

For development with race detection:
```bash
go build -race -o vpsentinel-agent ./main.go
```

## Release Build Script

Create a `build-release.sh` script:

```bash
#!/bin/bash
VERSION=${1:-"dev"}
OUTPUT_DIR="release"

mkdir -p $OUTPUT_DIR

# Linux
GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$VERSION" -o $OUTPUT_DIR/vpsentinel-agent-linux-amd64 ./main.go

# macOS
GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$VERSION" -o $OUTPUT_DIR/vpsentinel-agent-darwin-amd64 ./main.go
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.Version=$VERSION" -o $OUTPUT_DIR/vpsentinel-agent-darwin-arm64 ./main.go

# Windows
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.Version=$VERSION" -o $OUTPUT_DIR/vpsentinel-agent-windows-amd64.exe ./main.go

echo "Builds complete in $OUTPUT_DIR/"
```

## Binary Size Targets

- **With debug info**: < 20MB
- **Stripped (-s -w)**: < 10MB
- **Static binary**: Depends on platform

## Verification

After building, verify the binary:

```bash
# Check binary info
file vpsentinel-agent

# Check dependencies (Linux/macOS)
ldd vpsentinel-agent  # Linux
otool -L vpsentinel-agent  # macOS

# Test run (will fail without config, but shows version)
./vpsentinel-agent --version
```

## Distribution Checklist

- [ ] Build for all target platforms
- [ ] Test each binary on target platform
- [ ] Verify binary size is acceptable
- [ ] Test with sample config.json
- [ ] Verify version string is correct
- [ ] Create checksums (SHA256)
- [ ] Package with sample config and README

## Checksums

Generate checksums for distribution:

```bash
# Linux
sha256sum vpsentinel-agent-linux-amd64 > vpsentinel-agent-linux-amd64.sha256

# macOS
shasum -a 256 vpsentinel-agent-darwin-amd64 > vpsentinel-agent-darwin-amd64.sha256

# Windows
certutil -hashfile vpsentinel-agent-windows-amd64.exe SHA256 > vpsentinel-agent-windows-amd64.exe.sha256
```
