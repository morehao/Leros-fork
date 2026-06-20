$ErrorActionPreference = 'Stop'
. "$PSScriptRoot\shared.ps1"

$root = Get-LerosRepoRoot
Set-OptionalGoBuildEnvironment
$goExe = Get-GoExe

if (-not (Test-Path "$root\bundles")) {
    New-Item -ItemType Directory -Path "$root\bundles" | Out-Null
}

Set-Location $root
Write-Host '[Leros] Rebuilding backend...' -ForegroundColor Cyan
& $goExe build -v -o "$root\bundles\leros.exe" .\backend\cmd\leros\

if ($LASTEXITCODE -ne 0) {
    throw 'Backend rebuild failed.'
}

Write-Host '[Leros] Backend rebuild completed.' -ForegroundColor Green
