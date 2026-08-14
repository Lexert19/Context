$ErrorActionPreference = "Stop"

$binDir = Join-Path $env:USERPROFILE "bin"
$exeName = "context.exe"
$targetPath = Join-Path $binDir $exeName

Write-Host "==> Building '$exeName'..." -ForegroundColor Cyan

if (-not (Test-Path -Path $binDir)) {
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    Write-Host "Created directory: $binDir" -ForegroundColor DarkGray
}

try {
    go build -o $targetPath .
    Write-Host "[OK] Successfully built: $targetPath" -ForegroundColor Green
}
catch {
    Write-Host "[ERROR] Go build failed!" -ForegroundColor Red
    exit 1
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$paths = $userPath -split ';' | Where-Object { $_ -ne "" }

if ($paths -notcontains $binDir) {
    Write-Host "Adding $binDir to user PATH..." -ForegroundColor Yellow
    $newUserPath = if ($userPath) { "$userPath;$binDir" } else { $binDir }
    [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
    
    $env:PATH = "$binDir;$env:PATH"
    Write-Host "[OK] Added to PATH!" -ForegroundColor Green
} else {
    if (($env:PATH -split ';') -notcontains $binDir) {
        $env:PATH = "$binDir;$env:PATH"
    }
}

Write-Host "`nDone! You can now run 'context' in PowerShell." -ForegroundColor Cyan