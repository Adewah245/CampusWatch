param(
    [Parameter(Mandatory = $true)][string]$ServerUrl,
    [Parameter(Mandatory = $true)][string]$AgentId,
    [Parameter(Mandatory = $true)][string]$AgentCode,
    [Parameter(Mandatory = $true)][string]$Credential,
    [Parameter(Mandatory = $true)][string]$SystemId,
    [string]$Binary = ".\campuswatch-agent.exe"
)

$ErrorActionPreference = "Stop"

$InstallDir = Join-Path $env:ProgramFiles "CampusWatch\Agent"
$ConfigPath = Join-Path $InstallDir "agent.env"
$ServiceName = "CampusWatchAgent"

if (-not (Test-Path $Binary)) { throw "Agent binary not found: $Binary" }
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item $Binary (Join-Path $InstallDir "campuswatch-agent.exe") -Force

@"
CAMPUSWATCH_SERVER_URL=$ServerUrl
CAMPUSWATCH_AGENT_ID=$AgentId
CAMPUSWATCH_AGENT_CODE=$AgentCode
CAMPUSWATCH_AGENT_CREDENTIAL=$Credential
CAMPUSWATCH_SYSTEM_ID=$SystemId
"@ | Set-Content -Path $ConfigPath -Encoding ascii

$acl = Get-Acl $ConfigPath
$acl.SetAccessRuleProtection($true, $false)
$systemRule = New-Object System.Security.AccessControl.FileSystemAccessRule("SYSTEM", "FullControl", "Allow")
$adminsRule = New-Object System.Security.AccessControl.FileSystemAccessRule("Administrators", "FullControl", "Allow")
$acl.AddAccessRule($systemRule)
$acl.AddAccessRule($adminsRule)
Set-Acl $ConfigPath $acl

$binaryPath = Join-Path $InstallDir "campuswatch-agent.exe"
if (Get-Service -Name $ServiceName -ErrorAction SilentlyContinue) {
    Stop-Service $ServiceName -ErrorAction SilentlyContinue
    sc.exe delete $ServiceName | Out-Null
}
sc.exe create $ServiceName binPath= "`"$binaryPath`"" start= auto | Out-Null
sc.exe description $ServiceName "CampusWatch Monitoring Agent" | Out-Null
Start-Service $ServiceName
Write-Output "CampusWatch agent installed and started."