$ErrorActionPreference = 'Stop'

$packageName    = 'muxic'
$url64          = 'https://github.com/punkscience/muxic/releases/download/v1.0.0/muxic_1.0.0_windows_amd64.zip'
$url64arm       = 'https://github.com/punkscience/muxic/releases/download/v1.0.0/muxic_1.0.0_windows_arm64.zip'
$checksum64     = ''
$checksum64arm  = ''
$checksumType   = 'sha256'

# Select arch from the environment (fast, no WMI dependency).
if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
    $url        = $url64arm
    $checksum   = $checksum64arm
} else {
    $url        = $url64
    $checksum   = $checksum64
}

$toolsDir = "$(Split-Path -Parent $MyInvocation.MyCommand.Definition)"

$packageArgs = @{
    packageName   = $packageName
    unzipLocation = $toolsDir
    url64bit      = $url
    checksum64    = $checksum
    checksumType64= $checksumType
}

Install-ChocolateyZipPackage @packageArgs
