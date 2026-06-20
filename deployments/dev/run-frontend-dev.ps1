$ErrorActionPreference = 'Stop'
. "$PSScriptRoot\shared.ps1"

$root = Get-LerosRepoRoot
$pnpmExe = Get-PnpmExe

Set-Location "$root\frontend"
Write-Host '[Leros][Frontend] Starting on http://localhost:3005' -ForegroundColor Cyan
& $pnpmExe run dev:web
