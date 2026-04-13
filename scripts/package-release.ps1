param(
    [string]$Version = (Get-Date -Format 'yyyyMMdd-HHmmss')
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$npmCommand = 'npm.cmd'

$root = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$releaseRoot = Join-Path $root 'release'
$packageName = "codex-blog-$Version"
$packageDir = Join-Path $releaseRoot $packageName
$zipPath = Join-Path $releaseRoot "$packageName.zip"
$backendDir = Join-Path $root 'backend'
$readerDir = Join-Path $root 'frontend\reader'
$adminDir = Join-Path $root 'frontend\admin'

if (Test-Path $packageDir) {
    Remove-Item -LiteralPath $packageDir -Recurse -Force
}

if (Test-Path $zipPath) {
    Remove-Item -LiteralPath $zipPath -Force
}

New-Item -ItemType Directory -Path $packageDir | Out-Null
New-Item -ItemType Directory -Path (Join-Path $packageDir 'backend\configs') | Out-Null
New-Item -ItemType Directory -Path (Join-Path $packageDir 'frontend\reader') | Out-Null
New-Item -ItemType Directory -Path (Join-Path $packageDir 'frontend\admin') | Out-Null
New-Item -ItemType Directory -Path (Join-Path $packageDir 'scripts') | Out-Null
New-Item -ItemType Directory -Path (Join-Path $packageDir 'docs') | Out-Null
New-Item -ItemType Directory -Path (Join-Path $packageDir 'database') | Out-Null

Write-Host "==> Building backend binaries"
Push-Location $backendDir
try {
    go build -o (Join-Path $packageDir 'backend\codex-blog-server.exe') .\cmd\server
    go build -o (Join-Path $packageDir 'backend\dailybriefing-fetcher.exe') .\cmd\dailybriefing_fetcher
} finally {
    Pop-Location
}

Write-Host "==> Building reader frontend"
Push-Location $readerDir
try {
    & $npmCommand run build
} finally {
    Pop-Location
}

Write-Host "==> Building admin frontend"
Push-Location $adminDir
try {
    & $npmCommand run build
} finally {
    Pop-Location
}

Write-Host "==> Copying deployment files"
Copy-Item -LiteralPath (Join-Path $backendDir 'configs\config.example.json') -Destination (Join-Path $packageDir 'backend\configs\config.example.json')
Copy-Item -LiteralPath (Join-Path $readerDir 'dist') -Destination (Join-Path $packageDir 'frontend\reader') -Recurse
Copy-Item -LiteralPath (Join-Path $adminDir 'dist') -Destination (Join-Path $packageDir 'frontend\admin') -Recurse
Copy-Item -LiteralPath (Join-Path $root 'scripts\fetch-daily-briefings.ps1') -Destination (Join-Path $packageDir 'scripts\fetch-daily-briefings.ps1')
Copy-Item -LiteralPath (Join-Path $root 'DEPLOYMENT.md') -Destination (Join-Path $packageDir 'docs\DEPLOYMENT.md')
Copy-Item -LiteralPath (Join-Path $root 'README.md') -Destination (Join-Path $packageDir 'docs\README.md')
Copy-Item -LiteralPath (Join-Path $root 'ai_blog.sql') -Destination (Join-Path $packageDir 'database\ai_blog.sql')

$manifest = @"
Package: $packageName
BuiltAt: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')
Contents:
- backend/codex-blog-server.exe
- backend/dailybriefing-fetcher.exe
- backend/configs/config.example.json
- frontend/reader/dist
- frontend/admin/dist
- scripts/fetch-daily-briefings.ps1
- docs/DEPLOYMENT.md
- docs/README.md
- database/ai_blog.sql
"@

Set-Content -LiteralPath (Join-Path $packageDir 'PACKAGE_INFO.txt') -Value $manifest -Encoding utf8

Write-Host "==> Creating zip archive"
Compress-Archive -Path (Join-Path $packageDir '*') -DestinationPath $zipPath -Force

Write-Host "Package directory: $packageDir"
Write-Host "Package archive:   $zipPath"
