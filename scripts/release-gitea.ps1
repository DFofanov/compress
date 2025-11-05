# PDF Compressor Release Generator for Gitea
# PowerShell version with Russian release описаниями
# Author: PDF Compressor Team
# Version: 1.0.0

param(
    [Parameter(Position=0)]
    [string]$Version,
    
    [Parameter()]
    [switch]$Help
)

# Ensure console uses UTF-8 to display Russian correctly
try { [Console]::OutputEncoding = [System.Text.Encoding]::UTF8 } catch {}
# Переменные конфигурации
$BINARY_NAME = "pdf-compressor"
$BUILD_DIR = "releases"
# Prefer environment variables; do not hardcode secrets
$GITEA_SERVER = $env:GITEA_SERVER
$GITEA_USER = $env:GITEA_USER
$GITEA_PASSWORD = $env:GITEA_PASSWORD
$GITEA_OWNER = $env:GITEA_OWNER
$GITEA_REPO = if ($env:GITEA_REPO) { $env:GITEA_REPO } else { "pdf-compressor" }

# Цвета для вывода
$Colors = @{
    Red = "Red"
    Green = "Green" 
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

# Функции вывода сообщений
function Write-Log {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Red
    exit 1
}

# Функция справки
function Show-Help {
    Write-Host "PDF Compressor Release Generator" -ForegroundColor $Colors.Blue
    Write-Host ""
    Write-Host "Usage: .\release-gitea.ps1 [version]"
    Write-Host ""
    Write-Host "Parameters:"
    Write-Host "  -Version    Release version (e.g.: v1.2.0)"
    Write-Host "              If not specified, uses VERSION file or latest git tag"
    Write-Host "  -Help       Show this help"
    Write-Host ""
    Write-Host "Environment variables:"
    Write-Host "  GITEA_SERVER    Gitea server URL"
    Write-Host "  GITEA_USER      Gitea username"
    Write-Host "  GITEA_PASSWORD  Gitea password"
    Write-Host "  GITEA_OWNER     Repository owner"
    Write-Host "  GITEA_REPO      Repository name"
    Write-Host "  .env            Automatically loaded from project root (KEY=VALUE)"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\release-gitea.ps1                    # Auto-detect version"
    Write-Host "  .\release-gitea.ps1 -Version v1.2.0   # Specific version"
    Write-Host ""
}

# Load variables from a .env file into the current process environment
function Load-DotEnv {
    param(
        [string]$Path = ".env",
        [switch]$Override
    )
    try {
        $candidates = @()
        # current working directory
        $candidates += (Join-Path -Path (Get-Location) -ChildPath $Path)
        # script directory
        if ($PSScriptRoot) {
            $candidates += (Join-Path -Path $PSScriptRoot -ChildPath $Path)
            # repository root (one level up from scripts)
            $candidates += (Join-Path -Path (Split-Path -Parent $PSScriptRoot) -ChildPath $Path)
        }
        $envFile = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
        if (-not $envFile) { return }
        Write-Log "Loading .env from $envFile"
        $lines = Get-Content -Path $envFile -Encoding UTF8 -ErrorAction Stop
        foreach ($raw in $lines) {
            $line = $raw.Trim()
            if (-not $line) { continue }
            if ($line.StartsWith('#') -or $line.StartsWith(';')) { continue }
            # Remove inline comments that start with # after a space
            $hashIdx = $line.IndexOf(' # ')
            if ($hashIdx -gt 0) { $line = $line.Substring(0, $hashIdx).TrimEnd() }
            # Support optional leading 'export '
            if ($line -like 'export *') { $line = $line.Substring(7).TrimStart() }
            $eq = $line.IndexOf('=')
            if ($eq -lt 1) { continue }
            $key = $line.Substring(0, $eq).Trim()
            $val = $line.Substring($eq + 1).Trim()
            if ($val.StartsWith('"') -and $val.EndsWith('"') -and $val.Length -ge 2) {
                $val = $val.Substring(1, $val.Length - 2)
                $val = $val -replace "\\n", "`n" -replace "\\r", "" -replace "\\t", "`t" -replace "\\\\", "\\"
            } elseif ($val.StartsWith("'") -and $val.EndsWith("'") -and $val.Length -ge 2) {
                $val = $val.Substring(1, $val.Length - 2)
            }
            $existing = [Environment]::GetEnvironmentVariable($key, 'Process')
            if ($Override -or [string]::IsNullOrEmpty($existing)) {
                [Environment]::SetEnvironmentVariable($key, $val, 'Process')
            }
        }
    } catch {
        Write-Warn "Failed to load .env: $($_.Exception.Message)"
    }
}

# Функция проверки зависимостей
function Test-Dependencies {
    Write-Log "Checking dependencies..."
    
    # Check Go
    if (!(Get-Command "go" -ErrorAction SilentlyContinue)) {
        Write-Error-Custom "Go is not installed"
    }
    
    # Check git
    if (!(Get-Command "git" -ErrorAction SilentlyContinue)) {
        Write-Error-Custom "Git is not installed"
    }
    
    Write-Log "All dependencies found"
}

# Функция проверки переменных окружения
function Test-Environment {
    Write-Log "Checking environment variables..."
    
    # Refresh from environment (after Load-DotEnv) so .env overrides take effect
    $script:GITEA_SERVER   = $env:GITEA_SERVER
    $script:GITEA_USER     = $env:GITEA_USER
    $script:GITEA_PASSWORD = $env:GITEA_PASSWORD
    $script:GITEA_OWNER    = $env:GITEA_OWNER
    if (-not $script:GITEA_REPO -and $env:GITEA_REPO) { $script:GITEA_REPO = $env:GITEA_REPO }
    
    if ([string]::IsNullOrEmpty($script:GITEA_SERVER)) { Write-Error-Custom "GITEA_SERVER is not set" }
    if ([string]::IsNullOrEmpty($script:GITEA_USER))   { Write-Error-Custom "GITEA_USER is not set" }
    if ([string]::IsNullOrEmpty($script:GITEA_PASSWORD)) { Write-Error-Custom "GITEA_PASSWORD is not set" }
    if ([string]::IsNullOrEmpty($script:GITEA_OWNER))  { Write-Error-Custom "GITEA_OWNER is not set" }

    # Normalize values (strip quotes/spaces, remove trailing slash)
    $script:GITEA_SERVER = ($script:GITEA_SERVER).ToString().Trim().Trim('"', "'").TrimEnd('/')
    $script:GITEA_USER   = ($script:GITEA_USER).ToString().Trim().Trim('"', "'")
    $script:GITEA_PASSWORD = ($script:GITEA_PASSWORD).ToString().Trim()
    $script:GITEA_OWNER  = ($script:GITEA_OWNER).ToString().Trim().Trim('"', "'")
    $script:GITEA_REPO   = ($script:GITEA_REPO).ToString().Trim().Trim('"', "'")
    
    Write-Log "Environment variables checked"
    Write-Log "Server: $($script:GITEA_SERVER) | Repo: $($script:GITEA_OWNER)/$($script:GITEA_REPO)"
}

# Quick preflight checks against Gitea API
function Test-GiteaApi {
    $apiBase = "$($script:GITEA_SERVER)/api/v1"
    Write-Log "API base: $apiBase"
    try {
        $v = Invoke-RestMethod -Uri "$apiBase/version" -Method Get -ErrorAction Stop
        Write-Log "Gitea version: $($v.version)"
    } catch {
        Write-Error-Custom "API check failed: $($_.Exception.Message). URL: $apiBase/version"
    }
    try {
        $auth = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("$($script:GITEA_USER):$($script:GITEA_PASSWORD)"))
        Invoke-RestMethod -Uri "$apiBase/repos/$($script:GITEA_OWNER)/$($script:GITEA_REPO)" -Method Get -Headers @{ Authorization = "Basic $auth" } -ErrorAction Stop | Out-Null
        Write-Log "Repo access OK"
    } catch {
        Write-Error-Custom "Repo check failed: $($_.Exception.Message). URL: $apiBase/repos/$($script:GITEA_OWNER)/$($script:GITEA_REPO)"
    }
}

# Функция определения версии
function Get-ReleaseVersion {
    param([string]$InputVersion)
    
    if (![string]::IsNullOrEmpty($InputVersion)) {
        $script:Version = $InputVersion
    } elseif (Test-Path "VERSION") {
        $script:Version = (Get-Content "VERSION" -Raw).Trim()
    } else {
        try {
            $script:Version = git describe --tags --abbrev=0 2>$null
            if ([string]::IsNullOrEmpty($script:Version)) { $script:Version = "v1.0.0" }
        } catch { $script:Version = "v1.0.0" }
    }
    
    if (!$script:Version.StartsWith("v")) { $script:Version = "v$($script:Version)" }
    Write-Log "Release version: $($script:Version)"
}

# Проверка статуса git
function Test-GitStatus {
    Write-Log "Checking git status..."
    try { git rev-parse --git-dir | Out-Null } catch { Write-Error-Custom "Git repository not found" }
    
    $status = git status --porcelain
    if (![string]::IsNullOrEmpty($status)) {
        Write-Warn "There are uncommitted changes"
        $response = Read-Host "Continue? (y/N)"
        if ($response -notin @('y','Y')) { exit 1 }
    }
    
    $currentBranch = git branch --show-current
    if ($currentBranch -notin @('master','main')) {
        Write-Warn "You are not on master/main branch (current: $currentBranch)"
        $response = Read-Host "Continue? (y/N)"
        if ($response -notin @('y','Y')) { exit 1 }
    }
}

# Запуск тестов
function Invoke-Tests {
    Write-Log "Running tests..."
    $result = go test ./...
    if ($LASTEXITCODE -ne 0) { Write-Error-Custom "Tests failed" }
    Write-Log "All tests passed successfully"
}

# Создание тега
function New-GitTag {
    Write-Log "Creating tag $($script:Version)..."
    $existingTag = git tag -l $script:Version
    if (![string]::IsNullOrEmpty($existingTag)) {
        Write-Warn "Tag $($script:Version) already exists locally"
        $response = Read-Host "Overwrite? (y/N)"
        if ($response -in @('y','Y')) {
            git tag -d $script:Version
            Write-Log "Deleted local tag $($script:Version)"
        } else { exit 1 }
    }

    $releaseNotes = @"
Release $($script:Version)

New Features:
- Interface updates and improvements
- Performance optimization

Bug Fixes:
- Various fixes and stability improvements

Supported Platforms:
- Windows (64-bit)
- Linux (64-bit, ARM64)
- macOS (Intel 64-bit, Apple Silicon ARM64)
"@;

    git tag -a $script:Version -m $releaseNotes
    git push origin $script:Version --force
    Write-Log "Tag $($script:Version) created and pushed"
}

# Сборка бинарников
function Build-Binaries {
    Write-Log "Building binaries for different platforms..."
    $releaseDir = "$BUILD_DIR\$($script:Version)"
    New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null

    $platforms = @(
        @{GOOS="windows"; GOARCH="amd64"},
        @{GOOS="linux"; GOARCH="amd64"},
        @{GOOS="linux"; GOARCH="arm64"},
        @{GOOS="darwin"; GOARCH="amd64"},
        @{GOOS="darwin"; GOARCH="arm64"}
    )

    foreach ($platform in $platforms) {
        $output = "$releaseDir\$BINARY_NAME-$($script:Version)-$($platform.GOOS)-$($platform.GOARCH)"
        if ($platform.GOOS -eq "windows") { $output += ".exe" }
        Write-Log "Building for $($platform.GOOS)/$($platform.GOARCH)"
        $env:GOOS = $platform.GOOS; $env:GOARCH = $platform.GOARCH
        $buildTime = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"
        $ldflags = "-s -w -X main.version=$($script:Version) -X main.buildTime=$buildTime"
        go build -ldflags="$ldflags" -o $output cmd\main.go
        if ($LASTEXITCODE -ne 0) { Write-Error-Custom "Error: Build failed for $($platform.GOOS)/$($platform.GOARCH)" }
        Write-Log "Success: $($platform.GOOS)/$($platform.GOARCH) built successfully"
    }
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
}

# Создание архивов
function New-Archives {
    Write-Log "Creating archives..."
    Push-Location "$BUILD_DIR\$($script:Version)"
    try {
        Get-ChildItem "*windows*.exe" | ForEach-Object {
            $archive = $_.Name -replace '\.exe$', '.zip'
            Compress-Archive -Path $_.Name -DestinationPath $archive -Force
            Remove-Item $_.Name
            Write-Log "Created archive: $archive"
        }
        Get-ChildItem "*linux*", "*darwin*" | Where-Object { $_.Extension -ne ".zip" -and $_.Extension -ne ".gz" } | ForEach-Object {
            $archive = "$($_.Name).zip"
            Compress-Archive -Path $_.Name -DestinationPath $archive -Force
            Remove-Item $_.Name
            Write-Log "Created archive: $archive"
        }
    } finally { Pop-Location }
}

# Создание релиза в Gitea
function New-GiteaRelease {
    Write-Log "Creating release in Gitea..."
    $apiBase = "$GITEA_SERVER/api/v1"
    # Load Russian body from external UTF-8 file to avoid PS source encoding issues
    $bodyTemplatePath = Join-Path $PSScriptRoot 'release-body-ru.md'
    if (-not (Test-Path $bodyTemplatePath)) { Write-Error-Custom "Release body template not found: $bodyTemplatePath" }
    $releaseBody = [System.IO.File]::ReadAllText($bodyTemplatePath, (New-Object System.Text.UTF8Encoding($false)))
    $releaseBody = $releaseBody -replace "{{VERSION}}", "$($script:Version)"

    # Авторизация
    $credentials = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("$($GITEA_USER):$($GITEA_PASSWORD)"))
    $headers = @{ "Authorization" = "Basic $credentials"; "Content-Type" = "application/json; charset=utf-8" }

    $releaseId = $null

    # Проверяем существующий релиз
    try {
        $existing = Invoke-RestMethod -Uri "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases/tags/$($script:Version)" -Method Get -Headers $headers
        if ($existing -and $existing.id) {
            Write-Log "Release for tag $($script:Version) already exists (ID: $($existing.id)). Will upload assets."
            $releaseId = $existing.id
            # Если описание короткое — обновим полным русским
            if (-not $existing.body -or $existing.body.Length -lt 100) {
                $updateJson = @{ name = "PDF Compressor $($script:Version)"; body = $releaseBody } | ConvertTo-Json -Depth 3
                $tempUpdate = "temp-update-$($script:Version).json"
                # $updateJson | Out-File -FilePath $tempUpdate -Encoding UTF8
                [System.IO.File]::WriteAllText($tempUpdate, $updateJson, (New-Object System.Text.UTF8Encoding($false)))
                if (Get-Command "curl.exe" -ErrorAction SilentlyContinue) {
                    & curl.exe -s -X PATCH -H "Authorization: Basic $credentials" -H "Content-Type: application/json; charset=utf-8" --data-binary "@$tempUpdate" "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases/$releaseId" | Out-Null
                } else {
                    $updateBytes = [System.Text.Encoding]::UTF8.GetBytes($updateJson)
                    Invoke-RestMethod -Uri "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases/$releaseId" -Method Patch -Body $updateBytes -Headers $headers | Out-Null
                }
                Remove-Item $tempUpdate -ErrorAction SilentlyContinue
            }
        }
    } catch { Write-Log "No existing release found for tag $($script:Version), will create new one." }

    # Создаём релиз
    if (-not $releaseId) {
        $releaseObj = @{ tag_name = $script:Version; name = "PDF Compressor $($script:Version)"; body = $releaseBody; draft = $false; prerelease = $false }
        $releaseJson = ($releaseObj | ConvertTo-Json -Depth 4)
        $tempJsonFile = "temp-release-$($script:Version).json"
        [System.IO.File]::WriteAllText($tempJsonFile, $releaseJson, (New-Object System.Text.UTF8Encoding($false)))
        Start-Sleep -Seconds 1
        if (Get-Command "curl.exe" -ErrorAction SilentlyContinue) {
            try {
                Write-Log "Creating release via curl..."
                $releaseUrl = "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases"
                $curlResult = & curl.exe -s -X POST -H "Authorization: Basic $credentials" -H "Content-Type: application/json; charset=utf-8" --data-binary "@$tempJsonFile" "$releaseUrl"
                if ($LASTEXITCODE -eq 0) {
                    $response = $curlResult | ConvertFrom-Json
                    $releaseId = $response.id
                    Write-Log "Release created with ID: $releaseId via curl"
                } else { throw "Curl failed with exit code $LASTEXITCODE (URL: $releaseUrl)" }
            } catch {
                Write-Warn "Curl method failed: $($_.Exception.Message)"
                $minimalJson = @{ tag_name = $script:Version; name = "PDF Compressor $($script:Version)"; body = "Release $($script:Version)" } | ConvertTo-Json -Depth 2
                $minimalBytes = [System.Text.Encoding]::UTF8.GetBytes($minimalJson)
                try {
                    $response = Invoke-RestMethod -Uri "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases" -Method Post -Body $minimalBytes -Headers $headers
                    $releaseId = $response.id
                    Write-Log "Minimal release created with ID: $releaseId"
                } catch { Write-Error-Custom "Failed to create release: $($_.Exception.Message)" }
            }
        } else {
            $releaseBytes = [System.Text.Encoding]::UTF8.GetBytes($releaseJson)
            try {
                $response = Invoke-RestMethod -Uri "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases" -Method Post -Body $releaseBytes -Headers $headers
                $releaseId = $response.id
                Write-Log "Release created with ID: $releaseId via PowerShell"
            } catch { Write-Error-Custom "Failed to create release: $($_.Exception.Message)" }
        }
        Remove-Item $tempJsonFile -ErrorAction SilentlyContinue
    }

    # Fallback: resolve release ID if creation didn't return it
    if (-not $releaseId) {
        try {
            $check = Invoke-RestMethod -Uri "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases/tags/$($script:Version)" -Method Get -Headers $headers
            if ($check -and $check.id) {
                $releaseId = $check.id
                Write-Log "Release ID resolved via GET: $releaseId"
            }
        } catch {
            Write-Warn "Could not resolve release ID after creation: $($_.Exception.Message)"
        }
    }
    if (-not $releaseId) { Write-Error-Custom "Release created but ID not found. Aborting uploads." }

    # Загрузка архивов
    Write-Log "Uploading archives..."
    Get-ChildItem "$BUILD_DIR\$($script:Version)\*" | ForEach-Object {
        Write-Log "Uploading file $($_.Name)..."
        try {
            $filePath = $_.FullName
            if (Get-Command "curl.exe" -ErrorAction SilentlyContinue) {
                Write-Log "Using curl for upload..."
                & curl.exe -s -X POST -H "Authorization: Basic $credentials" -F "attachment=@$filePath" "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases/$releaseId/assets" | Out-Null
                if ($LASTEXITCODE -eq 0) { Write-Log "Success: file $($_.Name) uploaded via curl" } else { Write-Warn "Curl upload failed for $($_.Name)" }
            } else {
                $boundary = [System.Guid]::NewGuid().ToString()
                $LF = "`r`n"
                $fileContent = [System.IO.File]::ReadAllBytes($filePath)
                # Build multipart body header with proper PowerShell escaping
                $bodyHeader = @(
                    "--$boundary$LF"
                    "Content-Disposition: form-data; name=`"attachment`"; filename=`"$($_.Name)`"$LF"
                    "Content-Type: application/octet-stream$LF$LF"
                ) -join ""
                $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes($bodyHeader)
                $endBytes  = [System.Text.Encoding]::UTF8.GetBytes("$LF--$boundary--$LF")
                # Concatenate bytes efficiently
                $fullBody = New-Object byte[] ($bodyBytes.Length + $fileContent.Length + $endBytes.Length)
                [Array]::Copy($bodyBytes, 0, $fullBody, 0, $bodyBytes.Length)
                [Array]::Copy($fileContent, 0, $fullBody, $bodyBytes.Length, $fileContent.Length)
                [Array]::Copy($endBytes, 0, $fullBody, $bodyBytes.Length + $fileContent.Length, $endBytes.Length)
                $uploadHeaders = @{ "Authorization" = "Basic $credentials"; "Content-Type" = "multipart/form-data; boundary=$boundary" }
                Invoke-RestMethod -Uri "$apiBase/repos/$GITEA_OWNER/$GITEA_REPO/releases/$releaseId/assets" -Method Post -Body $fullBody -Headers $uploadHeaders | Out-Null
                Write-Log "Success: file $($_.Name) uploaded via PowerShell"
            }
        } catch { Write-Warn "Error uploading file $($_.Name): $($_.Exception.Message)" }
    }
}

# Главная функция
function Main {
    Write-Host "PDF Compressor Release Generator" -ForegroundColor $Colors.Blue
    Write-Host ""
    if ($Help) { Show-Help; return }
    try {
        # Load variables from .env before validating environment
        Load-DotEnv -Override
        Test-Dependencies
        Test-Environment
        Test-GiteaApi
        Get-ReleaseVersion $Version
        Test-GitStatus
        Invoke-Tests
        New-GitTag
        Build-Binaries
        New-Archives
        New-GiteaRelease
        Write-Log "Release $($script:Version) successfully created!"
        Write-Host ""
        Write-Host "Release available at:" -ForegroundColor $Colors.Green
        Write-Host "$GITEA_SERVER/$GITEA_OWNER/$GITEA_REPO/releases/tag/$($script:Version)"
        Write-Host ""
        Write-Host "Done! Release published and ready to use." -ForegroundColor $Colors.Green
    } catch { Write-Error-Custom "An error occurred: $($_.Exception.Message)" }
}

# Запуск
Main