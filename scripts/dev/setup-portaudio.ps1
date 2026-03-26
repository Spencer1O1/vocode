param(
  [string]$Msys2Root = "C:\tools\msys64",
  [switch]$InstallMsys2
)

function Assert-CommandExists {
  param([Parameter(Mandatory = $true)][string]$Command)

  $resolved = Get-Command $Command -ErrorAction SilentlyContinue
  if (-not $resolved) {
    throw "Command '$Command' not found in PATH."
  }
}

function Invoke-Mingw64 {
  param(
    [Parameter(Mandatory = $true)][string]$Mingw64Exe,
    [Parameter(Mandatory = $true)][string]$Command
  )

  # mingw64.exe runs inside MSYS2 and accepts: -c "<command>"
  & $Mingw64Exe -c $Command
  if ($LASTEXITCODE -ne 0) {
    throw "MSYS2 command failed (exit code: $LASTEXITCODE). Command: $Command"
  }
}

if ($InstallMsys2) {
  Assert-CommandExists -Command "choco"
  Write-Host "Installing MSYS2 via Chocolatey..."
  choco install msys2 -y | Out-Null
}

if (-not (Test-Path $Msys2Root)) {
  throw "MSYS2 root not found: $Msys2Root"
}

$mingw64Exe = Join-Path $Msys2Root "mingw64.exe"
if (-not (Test-Path $mingw64Exe)) {
  throw "mingw64.exe not found under MSYS2 root: $mingw64Exe"
}

Write-Host "Updating package database (pacman -Syu)..."
Invoke-Mingw64 -Mingw64Exe $mingw64Exe -Command "pacman -Syu --noconfirm"

$pkgs = @(
  "mingw-w64-x86_64-gcc",
  "mingw-w64-x86_64-portaudio",
  "mingw-w64-x86_64-pkg-config"
)

$pkgArgs = ($pkgs | ForEach-Object { $_ }) -join " "
Write-Host "Installing: $pkgArgs"
Invoke-Mingw64 -Mingw64Exe $mingw64Exe -Command "pacman -S --needed $pkgArgs --noconfirm"

Write-Host "Verifying pkg-config can find portaudio-2.0..."

# Use the MinGW pkg-config in the mingw64 environment.
Invoke-Mingw64 -Mingw64Exe $mingw64Exe -Command "pkg-config --modversion portaudio-2.0"

Write-Host ""
Write-Host "PortAudio + pkg-config setup complete."
Write-Host "Next: pnpm --filter @vocode/voice build"

