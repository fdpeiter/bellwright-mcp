# Synthetic fixtures only; no game data or Docker deployment required.
$ErrorActionPreference = 'Stop'
$fixture = Join-Path ([IO.Path]::GetTempPath()) ('steward-preflight-' + [guid]::NewGuid())
$keys = @('SAVE_DIR', 'UE4SS_LOG_DIR', 'MAP_ASSETS_DIR', 'RESOURCE_ANALYSIS_DIR')
$previous = @{}
function Check([string]$EnvPath, [int]$Code, [string[]]$Expected) {
    $output = & pwsh -NoProfile -File (Join-Path $PSScriptRoot 'preflight.ps1') -EnvFile $EnvPath -SkipDocker
    if ($LASTEXITCODE -ne $Code) { throw "Expected exit $Code; got $LASTEXITCODE" }
    foreach ($text in $Expected) {
        if (!($output | Where-Object { $_.Contains($text) })) { throw "Missing diagnostic: $text; synthetic output: $output" }
    }
}
try {
    foreach ($key in $keys) {
        $previous[$key] = [Environment]::GetEnvironmentVariable($key)
        Remove-Item -LiteralPath "Env:$key" -ErrorAction SilentlyContinue
    }
    New-Item -ItemType Directory -Path $fixture | Out-Null
    Check (Join-Path $fixture 'missing.env') 1 @('[MISSING_INPUT]', '[DOCKER_UNCHECKED]')
    $lines = foreach ($key in $keys) {
        $dir = Join-Path $fixture "$key with spaces"
        New-Item -ItemType Directory -Path $dir | Out-Null
        "$key=`"$dir`""
    }
    $envPath = Join-Path $fixture 'fixture.env'
    $lines | Set-Content -LiteralPath $envPath
    $map = Join-Path $fixture 'MAP_ASSETS_DIR with spaces'
    New-Item -ItemType Directory -Path (Join-Path $map 'exports') | Out-Null
    foreach ($file in @('exports/Karvenia_08_map_overlay.png', 'exports/Karvenia_08_companion.png', 'karvenia-background.json')) {
        Set-Content -LiteralPath (Join-Path $map $file) -Value 'synthetic presence fixture'
    }
    foreach ($file in @('resource-baseline.json', 'foraging-placement-candidates.json')) {
        Set-Content -LiteralPath (Join-Path $fixture "RESOURCE_ANALYSIS_DIR with spaces/$file") -Value 'synthetic presence fixture'
    }
    Check $envPath 1 @('UE4SS_LOG_DIR/UE4SS.log', '[WAITING_FOR_SAVE]')
    Set-Content -LiteralPath (Join-Path $fixture 'UE4SS_LOG_DIR with spaces/UE4SS.log') -Value 'synthetic presence fixture'
    Check $envPath 0 @('[INPUTS_PRESENT]', '[WAITING_FOR_SAVE]', '[DOCKER_UNCHECKED]')
    $nested = Join-Path $fixture 'SAVE_DIR with spaces/character'
    New-Item -ItemType Directory -Path $nested | Out-Null
    Set-Content -LiteralPath (Join-Path $nested 'synthetic.sav') -Value 'synthetic presence fixture'
    $output = & pwsh -NoProfile -File (Join-Path $PSScriptRoot 'preflight.ps1') -EnvFile $envPath -SkipDocker
    if ($LASTEXITCODE -ne 0 -or ($output -join "`n").Contains('[WAITING_FOR_SAVE]')) { throw 'Nested save was not detected.' }
    $env:SAVE_DIR = 'relative/path'
    Check $envPath 1 @('SAVE_DIR must be a literal absolute host directory')
    $env:SAVE_DIR = 'C:drive-relative'
    Check $envPath 1 @('SAVE_DIR must be a literal absolute host directory')
    Write-Output 'PASS: absent configuration, missing log, complete quoted paths, empty/nested saves, environment precedence, drive-relative path rejection.'
} finally {
    foreach ($key in $keys) { [Environment]::SetEnvironmentVariable($key, $previous[$key]) }
    $resolved = [IO.Path]::GetFullPath($fixture)
    $tempRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
    if ($resolved.StartsWith($tempRoot, [StringComparison]::OrdinalIgnoreCase) -and (Split-Path $resolved -Leaf).StartsWith('steward-preflight-')) {
        Remove-Item -LiteralPath $resolved -Recurse -Force -ErrorAction SilentlyContinue
    }
}
