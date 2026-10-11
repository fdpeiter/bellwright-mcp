# Command-vector regression; mocks Docker and preflight, never contacts either service.
$ErrorActionPreference = 'Stop'
$fixture = Join-Path ([IO.Path]::GetTempPath()) ('steward-startup-' + [guid]::NewGuid())
try {
    New-Item -ItemType Directory -Path $fixture | Out-Null
    $wrapper = Join-Path $fixture 'wrapper.ps1'
    @'
param([string]$Startup, [switch]$NoBuild)
function docker {
    $DockerArgs = @($args)
    $global:LASTEXITCODE = 0
    if ($DockerArgs[0] -eq 'info') { 'linux'; return }
    ConvertTo-Json -Compress -InputObject @($DockerArgs)
}
function pwsh { $global:LASTEXITCODE = 0 }
. $Startup -EnvFile synthetic.env -ProjectName bellwright-acceptance-test -NoBuild:$NoBuild
'@ | Set-Content -LiteralPath $wrapper
    foreach ($noBuild in @($false, $true)) {
        $argsForTest = @('-NoProfile', '-File', $wrapper, '-Startup', (Join-Path $PSScriptRoot 'start-local.ps1'))
        if ($noBuild) { $argsForTest += '-NoBuild' }
        $output = & pwsh @argsForTest
        if ($LASTEXITCODE -ne 0) { throw 'Mocked startup failed.' }
        $records = @($output | ForEach-Object { ,($_ | ConvertFrom-Json) })
        if ($records.Count -ne 3) { throw 'Expected config and two deployment commands.' }
        $flag = if ($noBuild) { '--no-build' } else { '--build' }
        foreach ($record in $records[1..2]) {
            if ($record -notcontains $flag -or $record -contains '-' -or $record -notcontains 'bellwright-acceptance-test') { throw 'Build/project arguments were not preserved.' }
        }
        if ($records[1][-1] -ne 'companion' -or $records[2][-1] -ne 'publisher' -or
            $records[2] -notcontains '--no-deps' -or $records[2] -notcontains '--force-recreate') { throw 'Dependency startup order lost.' }
    }
    Write-Output 'PASS: build/no-build flags remain whole arguments; isolated project and dependency order preserved.'
} finally {
    $resolved = [IO.Path]::GetFullPath($fixture)
    $tempRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
    if ($resolved.StartsWith($tempRoot, [StringComparison]::OrdinalIgnoreCase) -and (Split-Path $resolved -Leaf).StartsWith('steward-startup-')) {
        Remove-Item -LiteralPath $resolved -Recurse -Force -ErrorAction SilentlyContinue
    }
}
