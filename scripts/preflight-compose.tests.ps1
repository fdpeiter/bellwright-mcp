# Opt-in Compose configuration check plus host startup gate. Never deploys a stack.
$ErrorActionPreference = 'Stop'
$repo = Split-Path $PSScriptRoot -Parent
$fixture = Join-Path ([IO.Path]::GetTempPath()) ('steward-compose-' + [guid]::NewGuid())
$project = 'steward-preflight-' + [guid]::NewGuid().ToString('N')
$keys = @('SAVE_DIR', 'UE4SS_LOG_DIR', 'MAP_ASSETS_DIR', 'RESOURCE_ANALYSIS_DIR')
$previous = @{}
try {
    foreach ($key in $keys) {
        $previous[$key] = [Environment]::GetEnvironmentVariable($key)
        Remove-Item -LiteralPath "Env:$key" -ErrorAction SilentlyContinue
    }
    New-Item -ItemType Directory -Path $fixture | Out-Null
    $envPath = Join-Path $fixture 'fixture.env'
    $keys | ForEach-Object { "$($_)=`"$($fixture.Replace('\','/'))/$($_) missing`"" } | Set-Content -LiteralPath $envPath
    $baseArgs = @('compose', '--project-directory', $repo, '--env-file', $envPath, '-p', $project, '-f', (Join-Path $repo 'compose.yaml'))
    $json = & docker @baseArgs config --format json 2>$null
    if ($LASTEXITCODE -ne 0) { throw 'Synthetic owner Compose config failed.' }
    $config = $json | ConvertFrom-Json
    # Some Compose versions omit false booleans from JSON serialization.
    $source = Get-Content -LiteralPath (Join-Path $repo 'compose.yaml') -Raw
    if ([regex]::Matches($source, 'create_host_path:\s*false').Count -ne 5) { throw 'Expected five explicit host-creation guards.' }
    $count = 0
    foreach ($service in $config.services.PSObject.Properties.Value) {
        foreach ($mount in $service.volumes) {
            if ($mount.target -in @('/map-assets', '/game', '/assets', '/resources', '/saves')) {
                if (!$mount.read_only -or $mount.bind.create_host_path -eq $true) { throw 'Private bind safety flags lost.' }
                $count++
            }
        }
    }
    if ($count -ne 5) { throw 'Expected five private directory mounts.' }
    # Missing inputs stop the documented entry point before its first Docker action.
    $output = & pwsh -NoProfile -File (Join-Path $PSScriptRoot 'start-local.ps1') -EnvFile $envPath
    if ($LASTEXITCODE -ne 1 -or ($output -join "`n") -notmatch '\[MISSING_INPUT\]') { throw 'Startup gate did not reject missing inputs.' }
    if (Test-Path -LiteralPath (Join-Path $fixture 'MAP_ASSETS_DIR missing')) { throw 'Compose created a private host directory.' }
    Write-Output 'PASS: clean-owner quoted config, five read-only binds, startup gate rejected missing host input without directory creation.'
} finally {
    foreach ($key in $keys) {
        if ($null -eq $previous[$key]) { Remove-Item -LiteralPath "Env:$key" -ErrorAction SilentlyContinue }
        else { [Environment]::SetEnvironmentVariable($key, $previous[$key]) }
    }
    $resolved = [IO.Path]::GetFullPath($fixture)
    $tempRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
    if ($resolved.StartsWith($tempRoot, [StringComparison]::OrdinalIgnoreCase) -and (Split-Path $resolved -Leaf).StartsWith('steward-compose-')) {
        Remove-Item -LiteralPath $resolved -Recurse -Force -ErrorAction SilentlyContinue
    }
}
