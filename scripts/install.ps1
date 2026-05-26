#Requires -Version 5.1
param(
    [string]$Version = "1.0.1",
    [string]$InstallDir = "$env:USERPROFILE\.agentops",
    [switch]$SkipInit
)

$ErrorActionPreference = "Stop"
$Repo = "gog1withme/AgentOps"

function Get-Platform {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
        return "windows_arm64"
    }
    return "windows_amd64"
}

function Install-AgentOps {
    $platform = Get-Platform
    if ($platform -eq "windows_arm64") {
        throw "Windows ARM64 is not supported in v1.0.0 releases."
    }

    $archive = "agentops_${Version}_windows_amd64.zip"
    $url = "https://github.com/$Repo/releases/download/v${Version}/$archive"
    $checksumUrl = "https://github.com/$Repo/releases/download/v${Version}/checksums.txt"
    $tmpdir = Join-Path $env:TEMP ("agentops-install-" + [guid]::NewGuid().ToString())

    New-Item -ItemType Directory -Force -Path $tmpdir | Out-Null
    try {
        Write-Host "Downloading $url..."
        Invoke-WebRequest -Uri $url -OutFile (Join-Path $tmpdir $archive) -UseBasicParsing

        Write-Host "Verifying checksum..."
        $checksums = (Invoke-WebRequest -Uri $checksumUrl -UseBasicParsing).Content
        $expected = ($checksums -split "`n" | Where-Object { $_ -match " $archive$" }) -replace ' .*', ''
        if (-not $expected) {
            throw "Checksum entry not found for $archive"
        }
        $actual = (Get-FileHash (Join-Path $tmpdir $archive) -Algorithm SHA256).Hash.ToLower()
        if ($expected.ToLower() -ne $actual) {
            throw "Checksum mismatch for $archive"
        }

        Expand-Archive -Path (Join-Path $tmpdir $archive) -DestinationPath $tmpdir -Force

        $binDir = Join-Path $InstallDir "bin"
        $dashDir = Join-Path $InstallDir "dashboard\out"
        New-Item -ItemType Directory -Force -Path $binDir | Out-Null
        New-Item -ItemType Directory -Force -Path (Join-Path $InstallDir "dashboard") | Out-Null

        Copy-Item -Path (Join-Path $tmpdir "agentops.exe") -Destination (Join-Path $binDir "agentops.exe") -Force
        if (Test-Path $dashDir) {
            Remove-Item -Recurse -Force $dashDir
        }
        Copy-Item -Recurse -Path (Join-Path $tmpdir "dashboard\out") -Destination $dashDir -Force

        Write-Host ""
        Write-Host "AgentOps $Version installed to $InstallDir"
        Write-Host ""
        Write-Host "Add to your PATH:"
        Write-Host "  `$env:Path = `"$binDir;`$env:Path`""
        Write-Host ""
        Write-Host "Then run:"
        Write-Host "  agentops init"
        Write-Host "  agentops env"
        Write-Host "  agentops dev"
        Write-Host ""

        if (-not $SkipInit) {
            $env:Path = "$binDir;$env:Path"
            & (Join-Path $binDir "agentops.exe") init
        }
    }
    finally {
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue $tmpdir
    }
}

Install-AgentOps
