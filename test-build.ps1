# Quick test build script
# Run: powershell -ExecutionPolicy Bypass -File test-build.ps1

Write-Host "Testing Go build..." -ForegroundColor Green

# Check Go
$goVersion = go version 2>&1
Write-Host "Go version: $goVersion" -ForegroundColor Cyan

# Check if we're in the right directory
Write-Host "Current directory: $(Get-Location)" -ForegroundColor Cyan
Write-Host "main.go exists: $(Test-Path 'main.go')" -ForegroundColor Cyan
Write-Host "go.mod exists: $(Test-Path 'go.mod')" -ForegroundColor Cyan

# Create dist directory
if (-not (Test-Path "dist")) {
    New-Item -ItemType Directory -Path "dist" | Out-Null
    Write-Host "Created dist directory" -ForegroundColor Green
}

# Download dependencies and generate go.sum first
Write-Host "`nDownloading dependencies and generating go.sum..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "Failed to run go mod tidy" -ForegroundColor Red
    exit 1
}
go mod download
Write-Host "✓ Dependencies ready" -ForegroundColor Green

# Try building one binary
Write-Host "`nBuilding Linux AMD64 test binary..." -ForegroundColor Yellow
$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

# Try building
$output = go build -ldflags "-X main.Version=test" -o "dist\test-linux-amd64.exe" .\main.go 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Build successful!" -ForegroundColor Green
    Write-Host "Output file: dist\test-linux-amd64.exe" -ForegroundColor Cyan
    
    if (Test-Path "dist\test-linux-amd64.exe") {
        $fileInfo = Get-Item "dist\test-linux-amd64.exe"
        Write-Host "File size: $($fileInfo.Length) bytes" -ForegroundColor Green
        Write-Host "File created: $($fileInfo.CreationTime)" -ForegroundColor Green
    } else {
        Write-Host "❌ File was not created!" -ForegroundColor Red
    }
} else {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    Write-Host "Error output:" -ForegroundColor Red
    Write-Host $output -ForegroundColor Red
    Write-Host "Exit code: $LASTEXITCODE" -ForegroundColor Red
}

# List dist directory
Write-Host "`nFiles in dist directory:" -ForegroundColor Cyan
Get-ChildItem -Path "dist" -ErrorAction SilentlyContinue | Format-Table Name, @{Label="Size (KB)"; Expression={[math]::Round($_.Length/1KB, 2)}}, LastWriteTime -AutoSize
