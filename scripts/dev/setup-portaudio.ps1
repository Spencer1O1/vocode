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
    [Parameter(Mandatory = $true)][string]$Command
  )

  # MSYS2's `-c` belongs to bash, not `mingw64.exe`. Use bash with MSYSTEM=MINGW64
  # so pacman/pkg-config operate on the correct MinGW64 environment.
  $bashExe = Join-Path $Msys2Root "usr\bin\bash.exe"
  if (-not (Test-Path $bashExe)) {
    throw "bash.exe not found under MSYS2 root: $bashExe"
  }

  $msysRootUnix = $Msys2Root -replace "\\", "/"
  $mingw64BinUnix = "$msysRootUnix/mingw64/bin"
  $mingw64PkgConfigUnix = "$msysRootUnix/mingw64/lib/pkgconfig"

  # `$PATH` is a bash variable; escape the `$` in PowerShell with a backtick.
  & $bashExe -lc "export MSYSTEM=MINGW64; export PATH='$mingw64BinUnix':`$PATH; export PKG_CONFIG_PATH='$mingw64PkgConfigUnix'; $Command"
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

Write-Host "Updating package database (pacman -Syu)..."
Invoke-Mingw64 -Command "pacman -Syu --noconfirm"

$pkgs = @(
  "mingw-w64-x86_64-gcc",
  "mingw-w64-x86_64-portaudio",
  "mingw-w64-x86_64-pkg-config"
)

$pkgArgs = ($pkgs | ForEach-Object { $_ }) -join " "
Write-Host "Installing: $pkgArgs"
Invoke-Mingw64 -Command "pacman -S --needed $pkgArgs --noconfirm"

Write-Host "Verifying pkg-config can find portaudio-2.0..."

# Call pkg-config.exe directly (avoids PATH issues inside bash)
$pkgConfigExe = Join-Path $Msys2Root "mingw64\bin\pkg-config.exe"
$pkgConfigPath = Join-Path $Msys2Root "mingw64\lib\pkgconfig"
$env:PKG_CONFIG_PATH = $pkgConfigPath

& $pkgConfigExe --modversion portaudio-2.0
if ($LASTEXITCODE -ne 0) {
  throw "pkg-config failed to find portaudio-2.0 (exit code: $LASTEXITCODE)."
}

# Validate expected artifacts exist on disk (so CGO/cgo pkg-config can find them later).
$gccExe = Join-Path $Msys2Root "mingw64\bin\gcc.exe"
$pkgConfigExe = Join-Path $Msys2Root "mingw64\bin\pkg-config.exe"
$pcFile = Join-Path $Msys2Root "mingw64\lib\pkgconfig\portaudio-2.0.pc"

if (-not (Test-Path $gccExe)) {
  throw "gcc.exe not found: $gccExe"
}
if (-not (Test-Path $pkgConfigExe)) {
  throw "pkg-config.exe not found: $pkgConfigExe"
}
if (-not (Test-Path $pcFile)) {
  throw "portaudio-2.0.pc not found: $pcFile"
}

Write-Host ""
Write-Host "PortAudio + pkg-config setup complete."
Write-Host "Next: pnpm --filter @vocode/voice build"

