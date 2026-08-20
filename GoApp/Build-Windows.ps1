#!/usr/bin/env pwsh
# Builds AtomicAWG.exe for Windows: a windowsgui-subsystem binary (no black
# console window), with the atom icon embedded via the committed
# rsrc_windows_amd64.syso (see assets/atomicawg.ico and the comment at the
# bottom of this file for how to regenerate it).
$ErrorActionPreference = "Stop"

$ProjectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$Version = if ($env:APP_VERSION) { $env:APP_VERSION } else { "3.5.0" }
$DistDir = Join-Path $ProjectDir "dist"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go toolchain was not found."
    exit 1
}

New-Item -ItemType Directory -Force -Path $DistDir | Out-Null
$BinaryPath = Join-Path $DistDir "AtomicAWG.exe"

Push-Location $ProjectDir
try {
    $env:CGO_ENABLED = "1"
    go build -trimpath `
        -ldflags "-X main.appVersion=$Version -H=windowsgui" `
        -o $BinaryPath `
        .
} finally {
    Pop-Location
}

Write-Host "Done: $BinaryPath"

# --- Regenerating rsrc_windows_amd64.syso (only needed if assets/atomicawg.ico changes) ---
# go run github.com/tc-hib/go-winres@latest simply --icon assets/atomicawg.ico --manifest gui --out rsrc_windows
# mv rsrc_windows_windows_amd64.syso rsrc_windows_amd64.syso
# rm -f rsrc_windows_windows_386.syso
