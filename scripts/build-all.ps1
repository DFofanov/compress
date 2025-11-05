# Universal build script for all platforms
# Builds Windows and Linux binaries from Windows using cross-compilation
# Usage: .\scripts\build-all.ps1

# Exit on error
$ErrorActionPreference = "Stop"

# Read version from VERSION file
$VERSION = Get-Content -Path "VERSION" -Raw
$VERSION = $VERSION.Trim()

$BINARY = "compress"
$OUTPUT_DIR = "dist"

Write-Host "Building compress $VERSION for all platforms..." -ForegroundColor Green

# Create output directory if it doesn't exist
if (-Not (Test-Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR | Out-Null
}

# Build configurations
$platforms = @(
    @{OS="windows"; ARCH="amd64"; EXT=".exe"},
    @{OS="windows"; ARCH="arm64"; EXT=".exe"},
    @{OS="linux"; ARCH="amd64"; EXT=""},
    @{OS="linux"; ARCH="arm64"; EXT=""}
)

foreach ($platform in $platforms) {
    Write-Host "`nBuilding $($platform.OS) $($platform.ARCH)..." -ForegroundColor Cyan
    
    $env:CGO_ENABLED = "0"
    $env:GOOS = $platform.OS
    $env:GOARCH = $platform.ARCH
    
    $outputName = "$OUTPUT_DIR\${BINARY}-$($platform.OS)-$($platform.ARCH)$($platform.EXT)"
    
    go build -ldflags "-s -w" -o $outputName .\cmd
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build $($platform.OS) $($platform.ARCH)"
        exit 1
    }
    
    Write-Host "Success: $($platform.OS) $($platform.ARCH) built successfully" -ForegroundColor Green
}

# Clean up environment variables
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

Write-Host "`nPackaging releases..." -ForegroundColor Green

# Package Windows releases
foreach ($arch in @("amd64", "arm64")) {
    Write-Host "Packaging Windows $arch..." -ForegroundColor Cyan
    
    $tempDir = New-Item -ItemType Directory -Path "$OUTPUT_DIR\temp-windows-$arch" -Force
    Copy-Item "$OUTPUT_DIR\${BINARY}-windows-${arch}.exe" "$tempDir\${BINARY}.exe"
    Copy-Item "LICENSE" "$tempDir\" -ErrorAction SilentlyContinue
    Copy-Item "README.md" "$tempDir\" -ErrorAction SilentlyContinue
    Copy-Item "config.yaml.example" "$tempDir\" -ErrorAction SilentlyContinue
    
    $archiveName = "$OUTPUT_DIR\${BINARY}-${VERSION}-windows-${arch}.zip"
    Compress-Archive -Path "$tempDir\*" -DestinationPath $archiveName -Force
    Remove-Item $tempDir -Recurse -Force
    Remove-Item "$OUTPUT_DIR\${BINARY}-windows-${arch}.exe"
    
    # Generate SHA256
    $hash = Get-FileHash $archiveName -Algorithm SHA256 | Select-Object -ExpandProperty Hash
    $hash | Out-File "$archiveName.sha256" -Encoding ASCII
    
    Write-Host "Created: $archiveName" -ForegroundColor Yellow
}

# Package Linux releases
foreach ($arch in @("amd64", "arm64")) {
    Write-Host "Packaging Linux $arch..." -ForegroundColor Cyan
    
    $tempDir = New-Item -ItemType Directory -Path "$OUTPUT_DIR\temp-linux-$arch" -Force
    Copy-Item "$OUTPUT_DIR\${BINARY}-linux-${arch}" "$tempDir\${BINARY}"
    Copy-Item "LICENSE" "$tempDir\" -ErrorAction SilentlyContinue
    Copy-Item "README.md" "$tempDir\" -ErrorAction SilentlyContinue
    Copy-Item "config.yaml.example" "$tempDir\" -ErrorAction SilentlyContinue
    
    # Use tar via WSL or tar.exe if available
    $archiveName = "${BINARY}-${VERSION}-linux-${arch}.tar.gz"
    
    if (Get-Command "tar" -ErrorAction SilentlyContinue) {
        # Native Windows tar or WSL tar
        Push-Location $tempDir
        tar -czf "..\$archiveName" *
        Pop-Location
    } else {
        # Fallback: create zip for Linux (not ideal but works)
        Write-Warning "tar not found, creating .zip instead of .tar.gz for Linux"
        $archiveName = "${BINARY}-${VERSION}-linux-${arch}.zip"
        Compress-Archive -Path "$tempDir\*" -DestinationPath "$OUTPUT_DIR\$archiveName" -Force
    }
    
    Remove-Item $tempDir -Recurse -Force
    Remove-Item "$OUTPUT_DIR\${BINARY}-linux-${arch}"
    
    # Generate SHA256
    $hash = Get-FileHash "$OUTPUT_DIR\$archiveName" -Algorithm SHA256 | Select-Object -ExpandProperty Hash
    $hash | Out-File "$OUTPUT_DIR\$archiveName.sha256" -Encoding ASCII
    
    Write-Host "Created: $archiveName" -ForegroundColor Yellow
}

Write-Host "`nBuild complete! Artifacts:" -ForegroundColor Green
Get-ChildItem "$OUTPUT_DIR\*" -Include *.zip,*.tar.gz,*.sha256 | ForEach-Object {
    Write-Host "  - $($_.Name)" -ForegroundColor Yellow
}
Write-Host "`nDone!" -ForegroundColor Green
