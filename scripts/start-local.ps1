# Explicit deployment command. Private-file checks must pass before any Docker action.
[CmdletBinding()]
param(
    [string]$EnvFile = (Join-Path $PSScriptRoot '../.env'),
    [switch]$NoBuild
)
$ErrorActionPreference = 'Stop'
$repo = Split-Path $PSScriptRoot -Parent
& pwsh -NoProfile -File (Join-Path $PSScriptRoot 'preflight.ps1') -EnvFile $EnvFile -SkipDocker
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
if (!(Get-Command docker -ErrorAction SilentlyContinue)) { Write-Output '[DOCKER_UNAVAILABLE] Docker CLI not found.'; exit 2 }
$os = & docker info --format '{{.OSType}}' 2>$null
if ($LASTEXITCODE -ne 0 -or $os -ne 'linux') { Write-Output '[DOCKER_UNAVAILABLE] Start Docker Desktop in Linux-container mode.'; exit 2 }
$compose = @('compose', '--project-directory', $repo, '--env-file', $EnvFile)
& docker @compose config --quiet
if ($LASTEXITCODE -ne 0) { exit 2 }
$build = if ($NoBuild) { @('--no-build') } else { @('--build') }
# Complete companion updates before attaching/recreating its shared-network dependents.
& docker @compose up -d @build companion
if ($LASTEXITCODE -ne 0) { exit 2 }
& docker @compose up -d @build --no-deps --force-recreate live-map publisher
if ($LASTEXITCODE -ne 0) { exit 2 }
for ($attempt = 1; $attempt -le 3; $attempt++) {
    & pwsh -NoProfile -File (Join-Path $PSScriptRoot 'preflight.ps1') -EnvFile $EnvFile
    $result = $LASTEXITCODE
    if ($result -eq 0 -or $result -eq 1) { exit $result }
    if ($attempt -lt 3) { Start-Sleep -Seconds 2 }
}
exit $result
