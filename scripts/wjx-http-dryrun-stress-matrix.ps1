param(
    [string]$Config = "examples/wjx-http-preview.yaml",
    [string]$Fixture = "internal/provider/wjx/testdata/survey.html",
    [switch]$SkipFull
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$stressScript = Join-Path $PSScriptRoot "wjx-http-dryrun-stress.ps1"
. (Join-Path $PSScriptRoot "lib/powershell.ps1")
$powerShellCommand = Resolve-SurveyControllerPowerShell

$profiles = @(
    @{
        Name = "smoke"
        Args = @("-Config", $Config, "-Fixture", $Fixture, "-Target", "10", "-Concurrency", "2", "-MaxGoroutines", "8", "-Json")
    },
    @{
        Name = "budget"
        Args = @("-Config", $Config, "-Fixture", $Fixture, "-Target", "25", "-Concurrency", "5", "-MaxGoroutines", "8", "-ExpectFailureThreshold", "false", "-Json")
    }
)

if (-not $SkipFull) {
    $profiles += @{
        Name = "1000x1000"
        Args = @("-Config", $Config, "-Fixture", $Fixture, "-Target", "1000", "-Concurrency", "1000", "-MinThroughput", "1", "-MaxGoroutines", "8", "-ExpectFailureThreshold", "false", "-Json")
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
            completed = $summary.completed
            throughput_per_second = $summary.throughput_per_second
            heap_alloc_delta_bytes = $summary.heap_alloc_delta_bytes
            total_alloc_delta_bytes = $summary.total_alloc_delta_bytes
            goroutines = $summary.goroutines
            failure_threshold_reached = $summary.failure_threshold_reached
            draft_count = $summary.draft_count
            network = $summary.network
        }
    }
    $rows | Format-Table -AutoSize
}
finally {
    Pop-Location
}
