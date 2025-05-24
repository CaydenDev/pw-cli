$ErrorActionPreference = "Stop"

$version = "1.0.0"
$outputDir = "./dist"
$exeName = "pwvault.exe"

Write-Host "Building Password Vault v$version..."

if (-not (Test-Path $outputDir)) {
    New-Item -ItemType Directory -Path $outputDir | Out-Null
}

Write-Host "Compiling executable..."
try {
    go build -o "$outputDir/$exeName" -ldflags "-s -w" main.go
    if ($LASTEXITCODE -ne 0) { throw "Build failed" }
    
    Write-Host "\nBuild successful!"
    Write-Host "Executable: $outputDir\$exeName"
} catch {
    Write-Host "\nError: Build failed - $_" -ForegroundColor Red
    exit 1
}