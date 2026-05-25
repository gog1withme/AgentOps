param(
    [Parameter(Position = 0)]
    [ValidateSet("build", "test", "dev", "lint", "qa", "dashboard-install", "dashboard-build", "all", "clean")]
    [string]$Target = "build"
)

$ErrorActionPreference = "Stop"
$Root = $PSScriptRoot
$Binary = Join-Path $Root "bin\agentops.exe"
$Dashboard = Join-Path $Root "dashboard"
$Version = "1.0.0"
$LdFlags = "-ldflags `"-X github.com/gog1withme/AgentOps/cli/version.Version=$Version`""

function Ensure-Go {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        $goPath = "C:\Program Files\Go\bin\go.exe"
        if (Test-Path $goPath) {
            $env:Path = "C:\Program Files\Go\bin;" + $env:Path
        } else {
            throw "Go not found. Install from https://go.dev/dl/ or run: winget install GoLang.Go"
        }
    }
}

switch ($Target) {
    "build" {
        Ensure-Go
        New-Item -ItemType Directory -Force -Path (Join-Path $Root "bin") | Out-Null
        Push-Location $Root
        go build $LdFlags -o $Binary ./cli
        Pop-Location
        Write-Host "Built $Binary"
    }
    "test" {
        Ensure-Go
        Push-Location $Root
        go test ./...
        Pop-Location
    }
    "lint" {
        Ensure-Go
        Push-Location $Root
        go vet ./...
        Pop-Location
    }
    "qa" {
        Ensure-Go
        Push-Location $Root
        go run scripts/generate-qa.go
        Pop-Location
    }
    "dev" {
        & $PSScriptRoot\build.ps1 build
        & $Binary dev
    }
    "dashboard-install" {
        Push-Location $Dashboard
        npm install
        Pop-Location
    }
    "dashboard-build" {
        Push-Location $Dashboard
        npm run build
        Pop-Location
    }
    "all" {
        & $PSScriptRoot\build.ps1 build
        & $PSScriptRoot\build.ps1 dashboard-build
    }
    "clean" {
        Remove-Item -Recurse -Force -ErrorAction SilentlyContinue `
            (Join-Path $Root "bin"),
            (Join-Path $Dashboard ".next"),
            (Join-Path $Dashboard "out")
    }
}
