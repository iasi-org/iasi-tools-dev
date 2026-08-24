param(
    [Parameter(Mandatory = $false)]
    [string]$Path = ".\publish"
)

$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $Path -PathType Container)) {
    throw "Directory does not exist: $Path"
}

$ResolvedPath = (Resolve-Path -LiteralPath $Path).Path

Write-Host "Applying public ACL to: $ResolvedPath"

& icacls.exe `
    $ResolvedPath `
    "/inheritance:e" `
    "/grant" `
    "*S-1-5-32-545:(OI)(CI)M" `
    "/T" `
    "/C" `
    "/Q"

if ($LASTEXITCODE -ne 0) {
    throw "icacls failed with exit code $LASTEXITCODE"
}

Write-Host "ACL applied successfully."