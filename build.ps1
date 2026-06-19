# depx build script for Windows
# Automatically gets version from Git tag

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  depx Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Get version from Git tag
$version = git describe --tags --abbrev=0 2>$null
if (-not $version) {
    Write-Host "[!] No Git tag found, using default version dev" -ForegroundColor Yellow
    $version = "dev"
} else {
    Write-Host "[OK] Git tag: $version" -ForegroundColor Green
}

# Build with version injection
Write-Host ""
Write-Host "Building depx.exe..." -ForegroundColor Yellow

$ldflags = "-s -w -X github.com/mukunjin/depx/cmd.Version=$version"
go build "-ldflags=$ldflags" .

if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Build failed" -ForegroundColor Red
    exit 1
}

Write-Host "[OK] Build successful" -ForegroundColor Green
Write-Host ""
Write-Host "Version: $version" -ForegroundColor Cyan
Write-Host ""
