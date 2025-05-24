$ErrorActionPreference = "Stop"

$installDir = "$env:ProgramFiles\PasswordVault"
$exeName = "pwvault.exe"
$sourceExe = ".\dist\$exeName"

Write-Host "Installing Password Vault..."

if (-not (Test-Path $sourceExe)) {
    Write-Host "Error: Executable not found. Please run build.ps1 first." -ForegroundColor Red
    exit 1
}

try {
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir | Out-Null
    }

    Copy-Item $sourceExe -Destination $installDir -Force

    $envPath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if (-not $envPath.Contains($installDir)) {
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$envPath;$installDir",
            "Machine"
        )
    }

    Write-Host "\nInstallation successful!"
    Write-Host "Password Vault has been installed to: $installDir"
    Write-Host "You can now run 'pwvault' from any terminal."
} catch {
    Write-Host "\nError: Installation failed - $_" -ForegroundColor Red
    exit 1
}