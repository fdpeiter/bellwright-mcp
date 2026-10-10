# Read-only Windows host checks. Never prints save contents, telemetry or resolved paths.
[CmdletBinding()]
param(
    [string]$EnvFile = (Join-Path $PSScriptRoot '../.env'),
    [switch]$SkipDocker
)
$ErrorActionPreference = 'Stop'
$failed = $false
function Report([string]$State, [string]$Message) { Write-Output "[$State] $Message" }
$config = @{}
if (Test-Path -LiteralPath $EnvFile -PathType Leaf) {
    foreach ($line in Get-Content -LiteralPath $EnvFile) {
        if ($line -match '^\s*(SAVE_DIR|UE4SS_LOG_DIR|MAP_ASSETS_DIR|RESOURCE_ANALYSIS_DIR)\s*=\s*(.*?)\s*$') {
            $key = $Matches[1]; $value = $Matches[2]
            if ($value.Length -ge 2 -and (($value.StartsWith('"') -and $value.EndsWith('"')) -or ($value.StartsWith("'") -and $value.EndsWith("'")))) {
                $value = $value.Substring(1, $value.Length - 2)
            }
            $config[$key] = $value
        }
    }
} else { Report 'MISSING_INPUT' 'Environment file absent; copy .env.example to .env.'; $failed = $true }
$requirements = @{
    SAVE_DIR = @()
    UE4SS_LOG_DIR = @('UE4SS.log')
    MAP_ASSETS_DIR = @('exports/Karvenia_08_map_overlay.png', 'exports/Karvenia_08_companion.png', 'karvenia-background.json')
    RESOURCE_ANALYSIS_DIR = @('resource-baseline.json', 'foraging-placement-candidates.json')
}
foreach ($key in ($requirements.Keys | Sort-Object)) {
    # Compose gives process environment precedence over --env-file.
    $override = [Environment]::GetEnvironmentVariable($key)
    $value = if ($null -ne $override) { $override } else { $config[$key] }
    if ([string]::IsNullOrWhiteSpace($value) -or $value.Contains('$') -or ![IO.Path]::IsPathFullyQualified($value)) {
        Report 'MISSING_INPUT' "$key must be a literal absolute host directory (no variable expansion)."
        $failed = $true; continue
    }
    if (!(Test-Path -LiteralPath $value -PathType Container)) {
        Report 'MISSING_INPUT' "$key directory absent or inaccessible."; $failed = $true; continue
    }
    foreach ($file in $requirements[$key]) {
        if (!(Test-Path -LiteralPath (Join-Path $value $file) -PathType Leaf)) {
            Report 'MISSING_INPUT' "$key/$file absent or inaccessible."; $failed = $true
        }
    }
    if ($key -eq 'SAVE_DIR' -and !(Get-ChildItem -LiteralPath $value -Filter '*.sav' -File | Select-Object -First 1)) {
        Report 'WAITING_FOR_SAVE' 'Save directory exists but contains no .sav; publisher cannot ingest yet.'
    }
}
if (!$failed) { Report 'INPUTS_PRESENT' 'Required private files exist; content validity is checked by service endpoints below.' }
if ($IsWindows -or $env:OS -eq 'Windows_NT') {
    if (Get-Process -Name 'BellwrightGame', 'BellwrightGame-Win64-Shipping' -ErrorAction SilentlyContinue) {
        Report 'GAME_RUNNING' 'Game process exists; this does not prove a loaded session or fresh probe.'
    } else { Report 'GAME_INACTIVE' 'No Bellwright game process detected. Existing files are not fresh gameplay.' }
} else { Report 'GAME_UNKNOWN' 'Game process checks require the Windows gaming host.' }
if ($SkipDocker) { Report 'DOCKER_UNCHECKED' 'Docker/service checks skipped.'; if ($failed) { exit 1 }; exit 0 }
if ($failed) { exit 1 }
if (!(Get-Command docker -ErrorAction SilentlyContinue)) {
    Report 'DOCKER_UNAVAILABLE' 'Docker CLI not found.'; exit 2
}
$os = & docker info --format '{{.OSType}}' 2>$null
if ($LASTEXITCODE -ne 0 -or $os -ne 'linux') {
    Report 'DOCKER_UNAVAILABLE' 'Start Docker Desktop in Linux-container mode.'; exit 2
}
$repo = Split-Path $PSScriptRoot -Parent
& docker compose --project-directory $repo --env-file $EnvFile config --quiet 2>$null
if ($LASTEXITCODE -ne 0) { Report 'COMPOSE_INVALID' 'Compose configuration invalid; check .env and compose.yaml.'; exit 2 }
# Check live-map from inside companion because port 18126 is intentionally not published.
# Emit only classifications, never response bodies or private coordinates/IDs.
$check = (Get-Content -LiteralPath (Join-Path $PSScriptRoot 'check-services.mjs') -Raw) + "`nprocess.exitCode = await checkServices();"
& docker compose --project-directory $repo --env-file $EnvFile exec -T companion node --input-type=module -e $check 2>$null
if ($LASTEXITCODE -ne 0) { Report 'SERVICES_UNVERIFIED' 'Start services or repair dependent network attachment; see DOCKER-LOCAL.md.'; exit 2 }
exit 0
