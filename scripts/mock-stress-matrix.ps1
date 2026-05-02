param(
    [string]$Config = "examples/mock-run.yaml",
    [switch]$SkipFull
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$stressScript = Join-Path $PSScriptRoot "mock-stress.ps1"
. (Join-Path $PSScriptRoot "lib/powershell.ps1")
$powerShellCommand = Resolve-SurveyControllerPowerShell

$profiles = @(
    @{
        Name = "smoke"
        Args = @("-Config", $Config, "-Target", "10", "-Concurrency", "2", "-MaxGoroutines", "1", "-Json")
    },
    @{
        Name = "failure-threshold"
        Args = @("-Config", $Config, "-Target", "5", "-Concurrency", "1", "-FailEvery", "2", "-ExpectFailureThreshold", "true", "-Json")
    }
)

if (-not $SkipFull) {
    $profiles += @{
        Name = "1000x1000"
        Args = @("-Config", $Config, "-Target", "1000", "-Concurrency", "1000", "-MinThroughput", "1", "-MaxGoroutines", "1", "-Json")
    }
}

Push-Location $repoRoot
try {
    $rows = @()
    foreach ($profile in $profiles) {
        $commandArgs = New-SurveyControllerPowerShellFileArgs -Command $powerShellCommand -File $stressScript -Arguments $profile.Args
        $output = & $powerShellCommand.Source @commandArgs
        if ($LASTEXITCODE -ne 0) {
            throw "profile $($profile.Name) failed with exit code $LASTEXITCODE"
        }
        $summary = $output | ConvertFrom-Json
        $rows += [pscustomobject]@{
            profile = $profile.Name
            target = $summary.target
            concurrency = $summary.concurrency
            successes = $summary.successes
            failures = $summary.failures
            throughput_per_second = $summary.throughput_per_second
            heap_alloc_delta_bytes = $summary.heap_alloc_delta_bytes
            goroutines = $summary.goroutines
            failure_threshold_reached = $summary.failure_threshold_reached
        }
    }
    $rows | Format-Table -AutoSize
}
finally {
    Pop-Location
}
