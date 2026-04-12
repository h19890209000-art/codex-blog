Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
$backendDir = Join-Path $projectRoot "backend"

Push-Location $backendDir
try {
  go run ./cmd/dailybriefing_fetcher @args
} finally {
  Pop-Location
}
