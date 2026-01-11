# Build VPSentinel Agent binaries for all platforms (Windows PowerShell)
# Run: powershell -ExecutionPolicy Bypass -File build-binaries.ps1

# Ensure we're in the agent directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

Write-Host "Building VPSentinel Agent binaries..." -ForegroundColor Green
Write-Host "Working directory: $(Get-Location)" -ForegroundColor Cyan

# Check if Go is installed
$goVersion = go version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}
Write-Host "Go version: $goVersion" -ForegroundColor Cyan

# Create dist directory
$distPath = Join-Path (Get-Location) "dist"
if (-not (Test-Path $distPath)) {
    New-Item -ItemType Directory -Path $distPath | Out-Null
    Write-Host "Created dist directory" -ForegroundColor Green
}

# Download dependencies and generate go.sum
Write-Host ""
Write-Host "Downloading Go dependencies and generating go.sum..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to run go mod tidy" -ForegroundColor Red
    exit 1
}
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to download dependencies" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Dependencies downloaded and go.sum generated" -ForegroundColor Green

# Build Linux AMD64
Write-Host ""
Write-Host "Building Linux AMD64..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
$outputPath = Join-Path $distPath "vpsentinel-agent-linux-amd64.exe"
Write-Host "Output path: $outputPath" -ForegroundColor Cyan
$buildOutput = go build -ldflags "-X main.Version=dev" -o $outputPath .\main.go 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build Linux AMD64" -ForegroundColor Red
    Write-Host "Error output: $buildOutput" -ForegroundColor Red
    Write-Host "Error code: $LASTEXITCODE" -ForegroundColor Red
    exit 1
}
if (Test-Path $outputPath) {
    Write-Host "[OK] Linux AMD64 built successfully" -ForegroundColor Green
    $fileInfo = Get-Item $outputPath
    Write-Host "  File size: $($fileInfo.Length) bytes" -ForegroundColor Gray
} else {
    Write-Host "[ERROR] Build succeeded but file not found at: $outputPath" -ForegroundColor Red
    exit 1
}

# Build Linux ARM64
Write-Host ""
Write-Host "Building Linux ARM64..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
$outputPath = Join-Path $distPath "vpsentinel-agent-linux-arm64.exe"
$buildOutput = go build -ldflags "-X main.Version=dev" -o $outputPath .\main.go 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build Linux ARM64" -ForegroundColor Red
    Write-Host "Error output: $buildOutput" -ForegroundColor Red
    exit 1
}
if (Test-Path $outputPath) {
    Write-Host "[OK] Linux ARM64 built successfully" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Build succeeded but file not found" -ForegroundColor Red
    exit 1
}

# Build macOS AMD64
Write-Host ""
Write-Host "Building macOS AMD64..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
$outputPath = Join-Path $distPath "vpsentinel-agent-darwin-amd64.exe"
$buildOutput = go build -ldflags "-X main.Version=dev" -o $outputPath .\main.go 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build macOS AMD64" -ForegroundColor Red
    Write-Host "Error output: $buildOutput" -ForegroundColor Red
    exit 1
}
if (Test-Path $outputPath) {
    Write-Host "[OK] macOS AMD64 built successfully" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Build succeeded but file not found" -ForegroundColor Red
    exit 1
}

# Build macOS ARM64 (Apple Silicon)
Write-Host ""
Write-Host "Building macOS ARM64..." -ForegroundColor Yellow
$env:GOOS = "darwin"
$env:GOARCH = "arm64"
$env:CGO_ENABLED = "0"
$outputPath = Join-Path $distPath "vpsentinel-agent-darwin-arm64.exe"
$buildOutput = go build -ldflags "-X main.Version=dev" -o $outputPath .\main.go 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build macOS ARM64" -ForegroundColor Red
    Write-Host "Error output: $buildOutput" -ForegroundColor Red
    exit 1
}
if (Test-Path $outputPath) {
    Write-Host "[OK] macOS ARM64 built successfully" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Build succeeded but file not found" -ForegroundColor Red
    exit 1
}

# Build Windows AMD64
Write-Host ""
Write-Host "Building Windows AMD64..." -ForegroundColor Yellow
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
$outputPath = Join-Path $distPath "vpsentinel-agent-windows-amd64.exe"
$buildOutput = go build -ldflags "-X main.Version=dev" -o $outputPath .\main.go 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to build Windows AMD64" -ForegroundColor Red
    Write-Host "Error output: $buildOutput" -ForegroundColor Red
    exit 1
}
if (Test-Path $outputPath) {
    Write-Host "[OK] Windows AMD64 built successfully" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Build succeeded but file not found" -ForegroundColor Red
    exit 1
}

# Remove .exe extension from Linux/macOS binaries (Windows adds .exe to all)
Write-Host ""
Write-Host "Removing .exe extension from Linux/macOS binaries..." -ForegroundColor Yellow
$linuxAmd64 = Join-Path $distPath "vpsentinel-agent-linux-amd64.exe"
$linuxAmd64New = Join-Path $distPath "vpsentinel-agent-linux-amd64"
if (Test-Path $linuxAmd64) {
    Move-Item $linuxAmd64 $linuxAmd64New -Force
    Write-Host "[OK] Renamed Linux AMD64" -ForegroundColor Green
}

$linuxArm64 = Join-Path $distPath "vpsentinel-agent-linux-arm64.exe"
$linuxArm64New = Join-Path $distPath "vpsentinel-agent-linux-arm64"
if (Test-Path $linuxArm64) {
    Move-Item $linuxArm64 $linuxArm64New -Force
    Write-Host "[OK] Renamed Linux ARM64" -ForegroundColor Green
}

$darwinAmd64 = Join-Path $distPath "vpsentinel-agent-darwin-amd64.exe"
$darwinAmd64New = Join-Path $distPath "vpsentinel-agent-darwin-amd64"
if (Test-Path $darwinAmd64) {
    Move-Item $darwinAmd64 $darwinAmd64New -Force
    Write-Host "[OK] Renamed macOS AMD64" -ForegroundColor Green
}

$darwinArm64 = Join-Path $distPath "vpsentinel-agent-darwin-arm64.exe"
$darwinArm64New = Join-Path $distPath "vpsentinel-agent-darwin-arm64"
if (Test-Path $darwinArm64) {
    Move-Item $darwinArm64 $darwinArm64New -Force
    Write-Host "[OK] Renamed macOS ARM64" -ForegroundColor Green
}

Write-Host ""
Write-Host "[SUCCESS] All binaries built successfully!" -ForegroundColor Green
Write-Host "Binaries are in: $distPath" -ForegroundColor Cyan
Write-Host ""
Write-Host "Files created:" -ForegroundColor Cyan
Get-ChildItem -Path $distPath | Format-Table Name, @{Label="Size (KB)"; Expression={[math]::Round($_.Length/1KB, 2)}}, LastWriteTime -AutoSize
