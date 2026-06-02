# install-bridge.ps1 — install thermalik-bridge on Windows
# Run with: iwr https://raw.githubusercontent.com/saefy/TicketJS/main/Bridge/install-bridge.ps1 -useb | iex

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Repo    = "saefy/TicketJS"
$Binary  = "thermalik-bridge"
$Asset   = "thermalik-bridge_windows_amd64.zip"
$InstallDir = Join-Path $env:LOCALAPPDATA "thermalik-bridge"

Write-Host ""
Write-Host "  Thermalik Bridge Installer" -ForegroundColor Cyan
Write-Host "  ────────────────────────────────────────────────"
Write-Host "  Platform: windows/amd64"
Write-Host "  Install dir: $InstallDir"
Write-Host ""

# ── fetch latest release ──────────────────────────────────────────────────────

$ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
$Release = Invoke-RestMethod -Uri $ApiUrl -Headers @{ "User-Agent" = "thermalik-installer" }

$DownloadUrl = ($Release.assets | Where-Object { $_.name -eq $Asset } | Select-Object -First 1).browser_download_url
if (-not $DownloadUrl) {
    Write-Error "Could not find release asset: $Asset"
    exit 1
}

$ChecksumUrl = ($Release.assets | Where-Object { $_.name -eq "checksums.txt" } | Select-Object -First 1).browser_download_url

# ── download ──────────────────────────────────────────────────────────────────

$TempDir = Join-Path $env:TEMP "thermalik-bridge-install"
New-Item -ItemType Directory -Force -Path $TempDir | Out-Null

$ZipPath = Join-Path $TempDir $Asset
Write-Host "  Downloading $Asset..."
Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath -UseBasicParsing

# ── verify checksum ───────────────────────────────────────────────────────────

if ($ChecksumUrl) {
    $ChecksumsPath = Join-Path $TempDir "checksums.txt"
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumsPath -UseBasicParsing

    $Expected = (Get-Content $ChecksumsPath | Where-Object { $_ -match $Asset }) -split "\s+" | Select-Object -First 1
    $Actual   = (Get-FileHash -Path $ZipPath -Algorithm SHA256).Hash.ToLower()

    if ($Actual -ne $Expected) {
        Write-Error "Checksum mismatch. Download may be corrupted."
        exit 1
    }
    Write-Host "  + Checksum verified" -ForegroundColor Green
}

# ── install ───────────────────────────────────────────────────────────────────

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

$BinaryPath = Join-Path $InstallDir "$Binary.exe"
Copy-Item -Path (Join-Path $TempDir "$Binary.exe") -Destination $BinaryPath -Force

# Add to user PATH if not already present.
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    $env:PATH = "$env:PATH;$InstallDir"
    Write-Host "  + Added $InstallDir to user PATH" -ForegroundColor Green
}

# Clean up.
Remove-Item -Recurse -Force $TempDir

Write-Host "  + Installed to $BinaryPath" -ForegroundColor Green

$Ver = & $BinaryPath version 2>$null
Write-Host "  + Version: $Ver"

Write-Host ""
Write-Host "Run the setup wizard to configure your printers and install as a service:" -ForegroundColor Cyan
Write-Host ""
Write-Host "  thermalik-bridge setup"
Write-Host ""
