$ErrorActionPreference = 'Stop'

function Get-LerosRepoRoot {
    return (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
}

function Resolve-ToolPath {
    param(
        [Parameter(Mandatory = $true)]
        [string]$CommandName,

        [string[]]$FallbackPaths = @()
    )

    $command = Get-Command $CommandName -ErrorAction SilentlyContinue
    if ($command -and $command.Source) {
        return $command.Source
    }

    foreach ($path in $FallbackPaths) {
        if ($path -and (Test-Path $path)) {
            return $path
        }
    }

    throw "Required command not found: $CommandName"
}

function Get-DockerExe {
    return (Resolve-ToolPath -CommandName 'docker.exe' -FallbackPaths @(
        'E:\DevEnv\Docker\app\resources\bin\docker.exe'
    ))
}

function Get-GoExe {
    return (Resolve-ToolPath -CommandName 'go.exe' -FallbackPaths @(
        'E:\DevEnv\Go\goroot\bin\go.exe'
    ))
}

function Get-PnpmExe {
    return (Resolve-ToolPath -CommandName 'pnpm.cmd' -FallbackPaths @(
        'D:\nvm\nodejs\pnpm.cmd'
    ))
}

function Set-OptionalGoBuildEnvironment {
    $fallbackGoroot = 'E:\DevEnv\Go\goroot'
    $fallbackGopath = 'E:\DevEnv\Go\gopath'
    $fallbackGocache = 'E:\DevEnv\Go\cache'
    $fallbackGcc = 'E:\DevEnv\MSYS2\ucrt64\bin\gcc.exe'
    $fallbackGccDir = 'E:\DevEnv\MSYS2\ucrt64\bin'

    if (-not $env:GOROOT -and (Test-Path $fallbackGoroot)) {
        $env:GOROOT = $fallbackGoroot
    }
    if (-not $env:GOPATH -and (Test-Path $fallbackGopath)) {
        $env:GOPATH = $fallbackGopath
    }
    if (-not $env:GOCACHE -and (Test-Path $fallbackGocache)) {
        $env:GOCACHE = $fallbackGocache
    }
    if (-not $env:GOMODCACHE -and $env:GOPATH) {
        $env:GOMODCACHE = Join-Path $env:GOPATH 'pkg\mod'
    }
    if (-not $env:GOBIN -and $env:GOPATH) {
        $env:GOBIN = Join-Path $env:GOPATH 'bin'
    }

    $env:CGO_ENABLED = '1'

    if (-not $env:CC -and (Test-Path $fallbackGcc)) {
        $env:CC = $fallbackGcc
    }

    if ((Test-Path $fallbackGccDir) -and ($env:PATH -notlike "*$fallbackGccDir*")) {
        $env:PATH = "$fallbackGccDir;$env:PATH"
    }

    if ($env:GOROOT) {
        $goBinDir = Join-Path $env:GOROOT 'bin'
        if ((Test-Path $goBinDir) -and ($env:PATH -notlike "*$goBinDir*")) {
            $env:PATH = "$goBinDir;$env:PATH"
        }
    }
}

function Wait-DockerReady {
    $dockerExe = Get-DockerExe

    Write-Host '[Leros] Waiting for Docker engine...' -ForegroundColor Cyan
    for ($i = 0; $i -lt 30; $i++) {
        & $dockerExe info *> $null
        if ($LASTEXITCODE -eq 0) {
            return
        }

        Start-Sleep -Seconds 2
    }

    throw 'Docker engine did not become ready in time.'
}

function Get-IsAdministrator {
    $currentIdentity = [System.Security.Principal.WindowsIdentity]::GetCurrent()
    $currentPrincipal = New-Object System.Security.Principal.WindowsPrincipal($currentIdentity)
    return $currentPrincipal.IsInRole([System.Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Ensure-Administrator {
    param(
        [Parameter(Mandatory = $true)]
        [string]$ScriptPath
    )

    if (Get-IsAdministrator) {
        return $true
    }

    Write-Host '[Leros] Re-launching with administrator permission...' -ForegroundColor Yellow
    Start-Process -FilePath 'powershell.exe' -ArgumentList @(
        '-ExecutionPolicy', 'Bypass',
        '-File', $ScriptPath
    ) -Verb RunAs | Out-Null

    return $false
}

function Stop-DevProcessesByPorts {
    param(
        [int[]]$Ports
    )

    $repoRoot = Get-LerosRepoRoot
    $stoppedProcessIds = New-Object 'System.Collections.Generic.HashSet[int]'

    foreach ($port in $Ports) {
        $listenRows = netstat -ano -p tcp |
            Select-String -Pattern 'LISTENING\s+\d+$' |
            Where-Object { $_.ToString() -match "[:\.]$port\s" }

        foreach ($row in $listenRows) {
            $text = ($row.ToString() -replace '\s+', ' ').Trim()
            $parts = $text.Split(' ')
            if ($parts.Length -lt 5) {
                continue
            }

            $processId = $parts[-1]
            if ($processId -notmatch '^\d+$' -or $processId -eq '0') {
                continue
            }

            $pidValue = [int]$processId
            if ($stoppedProcessIds.Contains($pidValue)) {
                continue
            }

            $stoppedProcessIds.Add($pidValue) | Out-Null
            & taskkill /PID $pidValue /T /F *> $null

            if ($LASTEXITCODE -eq 0) {
                Write-Host "[Leros] Stopped process tree on port $port (PID: $pidValue)." -ForegroundColor Cyan
                continue
            }

            $proc = Get-CimInstance Win32_Process -Filter "ProcessId = $pidValue" -ErrorAction SilentlyContinue
            if ($proc -and $proc.CommandLine -and $proc.CommandLine -match [regex]::Escape($repoRoot)) {
                throw "Failed to stop process on port $port. Please run stop script as administrator."
            }
        }
    }
}
