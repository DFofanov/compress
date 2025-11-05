# PowerShell script for building Windows binaries
# Usage: .\scripts\build-windows.ps1

# Exit on error
$ErrorActionPreference = "Stop"

# Read version from VERSION file
$VERSION = Get-Content -Path "VERSION" -Raw
$VERSION = $VERSION.Trim()

$BINARY = "compress"
$OUTPUT_DIR = "dist"

Write-Host "Building compress $VERSION for Windows..." -ForegroundColor Green

# Create output directory if it doesn't exist
if (-Not (Test-Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
}

# Build for Windows amd64
Write-Host "`nBuilding Windows amd64..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags "-s -w" -o "$OUTPUT_DIR\${BINARY}-windows-amd64.exe" .\cmd
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build Windows amd64"
    exit 1
}

# Build for Windows arm64
Write-Host "Building Windows arm64..." -ForegroundColor Cyan
$env:GOOS = "windows"
$env:GOARCH = "arm64"
go build -ldflags "-s -w" -o "$OUTPUT_DIR\${BINARY}-windows-arm64.exe" .\cmd
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to build Windows arm64"
    exit 1
}

Write-Host "`nPackaging Windows releases..." -ForegroundColor Green

# Package Windows amd64
Write-Host "Packaging Windows amd64..." -ForegroundColor Cyan
$tempDir = New-Item -ItemType Directory -Path "$OUTPUT_DIR\temp-windows-amd64" -Force
Copy-Item "$OUTPUT_DIR\${BINARY}-windows-amd64.exe" "$tempDir\${BINARY}.exe"
Copy-Item "LICENSE" "$tempDir\" -ErrorAction SilentlyContinue
Copy-Item "README.md" "$tempDir\" -ErrorAction SilentlyContinue
Copy-Item "config.yaml.example" "$tempDir\" -ErrorAction SilentlyContinue
Compress-Archive -Path "$tempDir\*" -DestinationPath "$OUTPUT_DIR\${BINARY}-${VERSION}-windows-amd64.zip" -Force
Remove-Item $tempDir -Recurse -Force

# Package Windows arm64
Write-Host "Packaging Windows arm64..." -ForegroundColor Cyan
$tempDir = New-Item -ItemType Directory -Path "$OUTPUT_DIR\temp-windows-arm64" -Force
Copy-Item "$OUTPUT_DIR\${BINARY}-windows-arm64.exe" "$tempDir\${BINARY}.exe"
Copy-Item "LICENSE" "$tempDir\" -ErrorAction SilentlyContinue
Copy-Item "README.md" "$tempDir\" -ErrorAction SilentlyContinue
Copy-Item "config.yaml.example" "$tempDir\" -ErrorAction SilentlyContinue
Compress-Archive -Path "$tempDir\*" -DestinationPath "$OUTPUT_DIR\${BINARY}-${VERSION}-windows-arm64.zip" -Force
Remove-Item $tempDir -Recurse -Force

Write-Host "`nGenerating SHA256 checksums..." -ForegroundColor Green

# Generate SHA256 checksums
Get-FileHash "$OUTPUT_DIR\${BINARY}-${VERSION}-windows-amd64.zip" -Algorithm SHA256 | 
    Select-Object -ExpandProperty Hash | 
    Out-File "$OUTPUT_DIR\${BINARY}-${VERSION}-windows-amd64.zip.sha256" -Encoding ASCII

Get-FileHash "$OUTPUT_DIR\${BINARY}-${VERSION}-windows-arm64.zip" -Algorithm SHA256 | 
    Select-Object -ExpandProperty Hash | 
    Out-File "$OUTPUT_DIR\${BINARY}-${VERSION}-windows-arm64.zip.sha256" -Encoding ASCII

Write-Host "`nBuild complete! Artifacts:" -ForegroundColor Green
Write-Host "  - $OUTPUT_DIR\${BINARY}-${VERSION}-windows-amd64.zip" -ForegroundColor Yellow
Write-Host "  - $OUTPUT_DIR\${BINARY}-${VERSION}-windows-amd64.zip.sha256" -ForegroundColor Yellow
Write-Host "  - $OUTPUT_DIR\${BINARY}-${VERSION}-windows-arm64.zip" -ForegroundColor Yellow
Write-Host "  - $OUTPUT_DIR\${BINARY}-${VERSION}-windows-arm64.zip.sha256" -ForegroundColor Yellow
Write-Host "`nDone!" -ForegroundColor Green
