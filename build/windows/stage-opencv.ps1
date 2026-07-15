param(
    [Parameter(Mandatory = $true)]
    [string]$Source
)

$destination = Join-Path $PSScriptRoot "..\..\third_party\opencv\bin"

if (-not (Test-Path -LiteralPath $Source -PathType Container)) {
    throw "OpenCV runtime directory does not exist: $Source"
}

New-Item -ItemType Directory -Force -Path $destination | Out-Null
Get-ChildItem -LiteralPath $destination -Filter *.dll -File | Remove-Item -Force
Get-ChildItem -LiteralPath $Source -Filter *.dll -File | Copy-Item -Destination $destination -Force
