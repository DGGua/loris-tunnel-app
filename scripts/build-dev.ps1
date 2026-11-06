# Dev build script for Loris Tunnel fork.
# Sets productVersion to <base>.dev<yyyymmdd> (today), runs wails build,
# then restores wails.json and copies the exe to an allowed folder.
#
# Usage:  powershell -File scripts\build-dev.ps1 [-BaseVersion 2.0.0] [-OutDir <folder>]
param(
    [string]$BaseVersion = "",
    [string]$OutDir = "$env:USERPROFILE\Downloads"
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$wailsJson = Join-Path $repo 'wails.json'
$envName = 'dev'

# --- locate toolchain -------------------------------------------------------
$tools = 'C:\Users\DG_Gu\code\tools'
$env:PATH = "$tools\gopath\bin;$tools\go\bin;$tools\shims;$env:LOCALAPPDATA\Microsoft\WinGet\Links;$env:PATH"
$env:GOPATH   = "$tools\gopath"
$env:GOBIN    = "$tools\gopath\bin"
$env:GOCACHE  = "$tools\gocache"
$env:GOTMPDIR = "$tools\gotmp"
$env:TMP      = "$tools\gotmp"
$env:TEMP     = "$tools\gotmp"
$env:GOPROXY  = 'https://goproxy.cn,direct'
$env:GOSUMDB  = 'sum.golang.google.cn'

# --- compute version --------------------------------------------------------
$json = Get-Content $wailsJson -Raw -Encoding UTF8 | ConvertFrom-Json
if (-not $BaseVersion) {
    $BaseVersion = ($json.info.productVersion -replace '\.dev\d+$', '')
}
$buildDate = Get-Date -Format 'yyyyMMdd'
$version = "$BaseVersion.$envName$buildDate"

# --- patch wails.json -------------------------------------------------------
$json.info.productVersion = $version
$jsonText = $json | ConvertTo-Json -Depth 10
[IO.File]::WriteAllText($wailsJson, $jsonText, (New-Object Text.UTF8Encoding($false)))
Write-Host "productVersion = $version"

# --- build ------------------------------------------------------------------
Push-Location $repo
try {
    wails build -platform windows/amd64
    if ($LASTEXITCODE -ne 0) { throw "wails build failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}

# --- restore wails.json (repo stays clean; exe already embedded the version)
git -C $repo checkout -- wails.json

# --- deliver ----------------------------------------------------------------
$srcExe = Join-Path $repo 'build\bin\loris-tunnel.exe'
$outExe = Join-Path $OutDir 'loris-tunnel-fork.exe'
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
Copy-Item $srcExe $outExe -Force

Write-Host ""
Write-Host "Build OK:  v$version"
Write-Host "Artifact:  $outExe"
