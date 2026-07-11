$ErrorActionPreference = 'Stop'

# Install-ChocolateyZipPackage auto-shims every .exe it unzips. Chocolatey removes
# those shims automatically on uninstall, and the unzipped tools dir is purged with
# the package — so no manual file removal is needed here. This script exists so the
# package has an explicit, auditable uninstall entry point; extend it if the install
# script ever creates state outside the tools directory.

$packageName = 'muxic'

Write-Host "$packageName uninstalled. Shims and unpacked files are removed by Chocolatey."
