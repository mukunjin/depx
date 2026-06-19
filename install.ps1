# depx installer for Windows
# Usage: powershell -ExecutionPolicy Bypass -File .\install.ps1
# Uninstall: powershell -ExecutionPolicy Bypass -File .\install.ps1 -Uninstall

param(
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

# Configuration
$InstallDir = "$env:LOCALAPPDATA\depx"
$ExeName = "depx.exe"
$SourceExe = Join-Path $PSScriptRoot $ExeName

# Helper function to refresh PATH in current session
function Refresh-Path {
    $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $env:Path = "$userPath;$machinePath"
}

# Helper function to check if directory is in PATH
function Test-InPath {
    param([string]$Directory)
    $currentPath = $env:Path -split ";" | Where-Object { $_ -ne "" }
    return $currentPath -contains $Directory
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  depx Installer for Windows" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# ---------- Uninstall ----------
if ($Uninstall) {
    Write-Host "Uninstalling depx..." -ForegroundColor Yellow
    Write-Host ""
    
    # Remove from PATH
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath) {
        $paths = $userPath -split ";" | Where-Object { $_ -ne $InstallDir -and $_ -ne "" }
        $newPath = $paths -join ";"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "[OK] Removed $InstallDir from user PATH" -ForegroundColor Green
    }
    
    # Remove install directory
    if (Test-Path $InstallDir) {
        Remove-Item -Recurse -Force $InstallDir
        Write-Host "[OK] Removed $InstallDir" -ForegroundColor Green
    }
    
    # Refresh PATH in current session
    Refresh-Path
    
    Write-Host ""
    Write-Host "depx uninstalled successfully." -ForegroundColor Green
    Write-Host ""
    exit 0
}

# ---------- Install ----------

# Check if depx.exe exists
if (-not (Test-Path $SourceExe)) {
    Write-Host "Error: $ExeName not found in $PSScriptRoot" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please build it first:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

Write-Host "Installing depx..." -ForegroundColor Yellow
Write-Host ""

# Create install directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Host "[OK] Created $InstallDir" -ForegroundColor Green
} else {
    Write-Host "[OK] $InstallDir already exists" -ForegroundColor Green
}

# Copy executable
$DestExe = Join-Path $InstallDir $ExeName
try {
    if (Test-Path $DestExe) {
        Write-Host "[OK] Updating existing installation..." -ForegroundColor Yellow
    }
    Copy-Item -Force $SourceExe $DestExe
    Write-Host "[OK] Copied $ExeName to $InstallDir" -ForegroundColor Green
} catch {
    Write-Host "Error: Failed to copy $ExeName" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    exit 1
}

# Verify copy succeeded
if (-not (Test-Path $DestExe)) {
    Write-Host "Error: Copy verification failed" -ForegroundColor Red
    exit 1
}

# Add to PATH if not already present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if (-not $userPath) {
    $userPath = ""
}

$pathDirs = $userPath -split ";" | Where-Object { $_ -ne "" }
if ($pathDirs -notcontains $InstallDir) {
    if ([string]::IsNullOrWhiteSpace($userPath)) {
        $newPath = $InstallDir
    } else {
        $newPath = "$userPath;$InstallDir"
    }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "[OK] Added $InstallDir to user PATH" -ForegroundColor Green
} else {
    Write-Host "[OK] $InstallDir already in PATH" -ForegroundColor Green
}

# Refresh PATH in current session immediately
Refresh-Path

# Verify installation
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  depx installed successfully!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Test if depx is now accessible
if (Test-InPath $InstallDir) {
    try {
        $version = & $DestExe --version 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "[OK] Verified: $version" -ForegroundColor Green
            Write-Host ""
            Write-Host "You can now run:" -ForegroundColor Cyan
            Write-Host "  depx scan" -ForegroundColor Cyan
            Write-Host ""
            exit 0
        }
    } catch {
        # Ignore errors during verification
    }
}

# Fallback message if verification failed
Write-Host "[!] Note: PATH updated. You may need to:" -ForegroundColor Yellow
Write-Host "    - Restart your terminal, OR" -ForegroundColor Yellow
Write-Host "    - Run: `$env:Path = [Environment]::GetEnvironmentVariable('Path','User') + ';' + [Environment]::GetEnvironmentVariable('Path','Machine')" -ForegroundColor Yellow
Write-Host ""
Write-Host "Then run:" -ForegroundColor Cyan
Write-Host "  depx scan" -ForegroundColor Cyan
Write-Host ""
