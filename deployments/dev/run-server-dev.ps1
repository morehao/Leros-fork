$ErrorActionPreference = 'Stop'
. "$PSScriptRoot\shared.ps1"

$root = Get-LerosRepoRoot
$env:LEROS_STORAGE_LOCAL_DIR = "$root\leros-storage"

if (-not (Test-Path $env:LEROS_STORAGE_LOCAL_DIR)) {
    New-Item -ItemType Directory -Path $env:LEROS_STORAGE_LOCAL_DIR | Out-Null
}

Set-Location $root
Write-Host '[Leros][Server] Starting on http://localhost:8080' -ForegroundColor Cyan
& "$root\bundles\leros.exe" server --config "$root\deployments\dev\server.config.yaml" --workspace-root "$root\.leros-workspace\1\1\workspace"
