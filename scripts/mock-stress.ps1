param(
    [string]$Config = "examples/mock-run.yaml",
    [int]$Target = 1000,
    [int]$Concurrency = 1000,
    [int]$Seed = 7,
    [int]$FailEvery = 0,
    [double]$MinThroughput = 0,
    [UInt64]$MaxHeapDelta = 0,
    [int]$MaxGoroutines = 0,
    [ValidateSet("", "true", "false")]
    [string]$ExpectFailureThreshold = "",
    [switch]$Json
)

$ErrorActionPreference = "Stop"

if ($Target -le 0) {
    throw "Target must be greater than 0."
}
if ($Concurrency -le 0) {
    throw "Concurrency must be greater than 0."
}
if ($FailEvery -lt 0) {
    throw "FailEvery must not be negative."
}
if ($MinThroughput -lt 0) {
    throw "MinThroughput must not be negative."
}
if ($MaxGoroutines -lt 0) {
    throw "MaxGoroutines must not be negative."
}
$expectFailureThresholdValue = $null
if ($ExpectFailureThreshold -ne "") {
    $expectFailureThresholdValue = [bool]::Parse($ExpectFailureThreshold)
}

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    $surveyctlArgs = @(
        "run",
        "--mock",
        $Config,
        "--target",
        $Target,
        "--concurrency",
        $Concurrency,
        "--seed",
        $Seed
    )

    if ($FailEvery -gt 0) {
        $surveyctlArgs += @("--mock-fail-every", $FailEvery)
    }
    $surveyctlArgs += "--json"

    $raw = & go run ./cmd/surveyctl @surveyctlArgs
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    $summary = $raw | ConvertFrom-Json
    $failures = @()

    if ($MinThroughput -gt 0 -and [double]$summary.throughput_per_second -lt $MinThroughput) {
        $failures += "throughput_per_second $($summary.throughput_per_second) is below $MinThroughput"
    }
    if ($MaxHeapDelta -gt 0 -and [UInt64]$summary.heap_alloc_delta_bytes -gt $MaxHeapDelta) {
        $failures += "heap_alloc_delta_bytes $($summary.heap_alloc_delta_bytes) is above $MaxHeapDelta"
    }
    if ($MaxGoroutines -gt 0 -and [int]$summary.goroutines -gt $MaxGoroutines) {
        $failures += "goroutines $($summary.goroutines) is above $MaxGoroutines"
    }
    if ($null -ne $expectFailureThresholdValue -and [bool]$summary.failure_threshold_reached -ne $expectFailureThresholdValue) {
        $failures += "failure_threshold_reached $($summary.failure_threshold_reached) does not match $expectFailureThresholdValue"
    }

    if ($Json) {
        $raw
    }
    else {
        Write-Host "mock stress:"
        Write-Host "  target: $($summary.target)"
        Write-Host "  concurrency: $($summary.concurrency)"
        Write-Host "  successes: $($summary.successes)"
        Write-Host "  failures: $($summary.failures)"
        Write-Host "  completed: $($summary.completed)"
        Write-Host "  throughput_per_second: $($summary.throughput_per_second)"
        Write-Host "  heap_alloc_delta_bytes: $($summary.heap_alloc_delta_bytes)"
        Write-Host "  total_alloc_delta_bytes: $($summary.total_alloc_delta_bytes)"
        Write-Host "  goroutines: $($summary.goroutines)"
        Write-Host "  failure_threshold_reached: $($summary.failure_threshold_reached)"
    }

    if ($failures.Count -gt 0) {
        foreach ($failure in $failures) {
            Write-Error $failure
        }
        exit 1
    }
    exit 0
}
finally {
    Pop-Location
}
