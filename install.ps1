# install.ps1 - one-line installer for Windows

#Requires -Version 5.1
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$Repo       = 'nodora-org/nodora'
$Binary     = 'nodora'
$InstallDir = Join-Path $env:LOCALAPPDATA 'nodora\bin'

function Write-Info  { param([string]$Message) Write-Host "==> " -ForegroundColor Blue -NoNewline; Write-Host $Message }
function Write-Fail  { param([string]$Message) Write-Host "error: " -ForegroundColor Red -NoNewline; Write-Host $Message; exit 1 }

function Get-Arch {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        'AMD64' { return 'amd64' }
        'ARM64' { return 'arm64' }
        'x86'   {
            # 32-bit shell on a 64-bit OS still reports x86 here
            if ($env:PROCESSOR_ARCHITEW6432 -eq 'AMD64') { return 'amd64' }
            if ($env:PROCESSOR_ARCHITEW6432 -eq 'ARM64') { return 'arm64' }
            Write-Fail "unsupported architecture: x86 (32-bit)"
        }
        default { Write-Fail "unsupported architecture: $env:PROCESSOR_ARCHITECTURE" }
    }
}

function Main {
    $os   = 'windows'
    $arch = Get-Arch

    Write-Info "Detected platform: $os/$arch"

    # fetch latest release
    Write-Info "Fetching latest release..."
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" `
            -Headers @{ 'User-Agent' = 'nodora-windows-installer' }
    } catch {
        Write-Fail "could not determine latest release: $($_.Exception.Message)"
    }

    $latest = $release.tag_name
    if ([string]::IsNullOrEmpty($latest)) {
        Write-Fail "could not determine latest release"
    }

    Write-Info "Latest version: $latest"

    $filename    = "$Binary-$latest-$os-$arch.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$latest/$filename"

    # download to a temp directory
    $tmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ("nodora-install-" + [System.Guid]::NewGuid().ToString('N'))
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        Write-Info "Downloading $filename..."
        $zipPath = Join-Path $tmpDir $filename
        try {
            Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing `
                -Headers @{ 'User-Agent' = 'nodora-windows-installer' }
        } catch {
            Write-Fail "download failed - check that a release exists for $os/$arch"
        }

        # extract
        Write-Info "Extracting..."
        Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

        # find the binary
        $binaryPath = Get-ChildItem -Path $tmpDir -Filter "$Binary.exe" -Recurse -File |
            Select-Object -First 1
        if ($null -eq $binaryPath) {
            Write-Fail "could not find '$Binary.exe' in archive"
        }

        # install
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        $dest = Join-Path $InstallDir "$Binary.exe"
        Copy-Item -Path $binaryPath.FullName -Destination $dest -Force

        Write-Info "Installed $Binary $latest to $dest"
    } finally {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }

    # ensure install dir is on the user PATH
    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    $onPath = ($userPath -split ';') -contains $InstallDir
    if (-not $onPath) {
        Write-Info "Adding $InstallDir to your user PATH"
        $newPath = if ([string]::IsNullOrEmpty($userPath)) { $InstallDir } else { "$userPath;$InstallDir" }
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
        $env:Path = "$env:Path;$InstallDir"
        Write-Info "Open a new terminal for the PATH change to take effect."
    }

    Write-Info "Run '$Binary --help' to get started."
}

Main
