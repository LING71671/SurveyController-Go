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

$repoRoot = Split-Path -Parent $PSScriptRoot
$invariantCulture = [System.Globalization.CultureInfo]::InvariantCulture
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
    if ($MinThroughput -gt 0) {
        $surveyctlArgs += @("--min-throughput", $MinThroughput.ToString($invariantCulture))
    }
    if ($MaxHeapDelta -gt 0) {
        $surveyctlArgs += @("--max-heap-delta", $MaxHeapDelta.ToString($invariantCulture))
    }
    if ($MaxGoroutines -gt 0) {
        $surveyctlArgs += @("--max-goroutines", $MaxGoroutines)
    }
    if ($ExpectFailureThreshold -ne "") {
        $surveyctlArgs += @("--expect-failure-threshold", $ExpectFailureThreshold)
    }
    $surveyctlArgs += "--json"

    $raw = & go run ./cmd/surveyctl @surveyctlArgs
    $exitCode = $LASTEXITCODE
    $summary = $null
    if ($raw) {
        try {
            $summary = $raw | ConvertFrom-Json
        }
        catch {
            $raw
            exit $exitCode
        }
    }

    if ($Json) {
        if ($raw) {
            $raw
        }
    }
    elseif ($null -ne $summary) {
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
    elseif ($raw) {
        $raw
    }

    if ($exitCode -ne 0) {
        exit $exitCode
    }
    exit 0
}
finally {
    Pop-Location
}
