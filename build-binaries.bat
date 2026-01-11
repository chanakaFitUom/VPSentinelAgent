@echo off
REM Build VPSentinel Agent binaries for all platforms (Windows Batch)
REM Run: build-binaries.bat

echo Building VPSentinel Agent binaries...

REM Create dist directory
if not exist "dist" mkdir dist

REM Build Linux AMD64
echo Building Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-X main.Version=dev" -o dist\vpsentinel-agent-linux-amd64.exe main.go
if errorlevel 1 (
    echo Failed to build Linux AMD64
    exit /b 1
)

REM Build Linux ARM64
echo Building Linux ARM64...
set GOOS=linux
set GOARCH=arm64
go build -ldflags "-X main.Version=dev" -o dist\vpsentinel-agent-linux-arm64.exe main.go
if errorlevel 1 (
    echo Failed to build Linux ARM64
    exit /b 1
)

REM Build macOS AMD64
echo Building macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags "-X main.Version=dev" -o dist\vpsentinel-agent-darwin-amd64.exe main.go
if errorlevel 1 (
    echo Failed to build macOS AMD64
    exit /b 1
)

REM Build macOS ARM64 (Apple Silicon)
echo Building macOS ARM64...
set GOOS=darwin
set GOARCH=arm64
go build -ldflags "-X main.Version=dev" -o dist\vpsentinel-agent-darwin-arm64.exe main.go
if errorlevel 1 (
    echo Failed to build macOS ARM64
    exit /b 1
)

REM Build Windows AMD64
echo Building Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-X main.Version=dev" -o dist\vpsentinel-agent-windows-amd64.exe main.go
if errorlevel 1 (
    echo Failed to build Windows AMD64
    exit /b 1
)

REM Remove .exe extension from Linux/macOS binaries (Windows adds .exe to all)
echo.
echo Removing .exe extension from Linux/macOS binaries...
if exist "dist\vpsentinel-agent-linux-amd64.exe" (
    ren "dist\vpsentinel-agent-linux-amd64.exe" "vpsentinel-agent-linux-amd64"
)
if exist "dist\vpsentinel-agent-linux-arm64.exe" (
    ren "dist\vpsentinel-agent-linux-arm64.exe" "vpsentinel-agent-linux-arm64"
)
if exist "dist\vpsentinel-agent-darwin-amd64.exe" (
    ren "dist\vpsentinel-agent-darwin-amd64.exe" "vpsentinel-agent-darwin-amd64"
)
if exist "dist\vpsentinel-agent-darwin-arm64.exe" (
    ren "dist\vpsentinel-agent-darwin-arm64.exe" "vpsentinel-agent-darwin-arm64"
)

echo.
echo ✅ All binaries built successfully!
echo Binaries are in: dist\
dir dist
