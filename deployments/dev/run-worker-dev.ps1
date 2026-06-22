$ErrorActionPreference = 'Stop'
. "$PSScriptRoot\shared.ps1"

$root = Get-LerosRepoRoot
$env:LEROS_DEV = 'true'

Set-Location $root
Write-Host '[Leros][Worker] Starting on http://localhost:8081' -ForegroundColor Cyan
& "$root\bundles\leros.exe" worker --worker-id 1 --config "$root\deployments\dev\worker.config.yaml" --listen-addr ':8081' --workspace-root "$root\.leros-workspace\1\1\workspace"
