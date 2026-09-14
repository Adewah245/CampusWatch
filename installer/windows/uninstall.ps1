$ErrorActionPreference = "Stop"
$ServiceName = "CampusWatchAgent"
$InstallDir = Join-Path $env:ProgramFiles "CampusWatch\Agent"

Stop-Service $ServiceName -ErrorAction SilentlyContinue
sc.exe delete $ServiceName | Out-Null
if (Test-Path $InstallDir) { Remove-Item $InstallDir -Recurse -Force }
Write-Output "CampusWatch agent removed."