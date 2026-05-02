param(
    [string]$Config = "examples/wjx-http-preview.yaml",
    [string]$Fixture = "internal/provider/wjx/testdata/survey.html",
    [int]$Target = 1000,
    [int]$Concurrency = 1000,
    [int]$Seed = 7,
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
if ($Seed -eq 0) {
    throw "Seed must not be 0."
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
        "--wjx-http-dry-run",
        $Config,
        "--fixture",
        $Fixture,
        "--target",
        $Target,
        "--concurrency",
        $Concurrency,
        "--seed",
        $Seed
    )

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
            $result = $raw | ConvertFrom-Json
            $report = $result.report
            $firstDraft = $null
            if ($result.drafts.Count -gt 0) {
                $firstDraft = $result.drafts[0]
            }
            $firstDraftEndpoint = ""
            $firstDraftAnswerCount = 0
            if ($null -ne $firstDraft) {
                $firstDraftEndpoint = $firstDraft.endpoint
                $firstDraftAnswerCount = $firstDraft.answer_count
            }
            $throughput = 0
            $heapAllocDelta = 0
            $totalAllocDelta = 0
            if ($null -ne $report.throughput_per_second) {
                $throughput = $report.throughput_per_second
            }
            if ($null -ne $report.heap_alloc_delta_bytes) {
                $heapAllocDelta = $report.heap_alloc_delta_bytes
            }
            if ($null -ne $report.total_alloc_delta_bytes) {
                $totalAllocDelta = $report.total_alloc_delta_bytes
            }
            $summary = [pscustomobject]@{
                path = $result.path
                fixture = $result.fixture
                target = $report.target
                concurrency = $report.concurrency
                seed = $result.seed
                successes = $report.successes
                failures = $report.failures
                completed = $report.completed
                throughput_per_second = $throughput
                heap_alloc_delta_bytes = $heapAllocDelta
                total_alloc_delta_bytes = $totalAllocDelta
                goroutines = $report.goroutines
                failure_threshold_reached = $report.failure_threshold_reached
                draft_count = $result.draft_count
                first_draft_endpoint = $firstDraftEndpoint
                first_draft_answer_count = $firstDraftAnswerCount
                network = $result.network
            }
        }
        catch {
            $raw
            exit $exitCode
        }
    }

    if ($Json) {
        if ($null -ne $summary) {
            $summary | ConvertTo-Json -Depth 4
        }
    }
    elseif ($null -ne $summary) {
        Write-Host "wjx http dry-run stress:"
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
        Write-Host "  draft_count: $($summary.draft_count)"
        Write-Host "  first_draft_endpoint: $($summary.first_draft_endpoint)"
        Write-Host "  first_draft_answer_count: $($summary.first_draft_answer_count)"
        Write-Host "  network: $($summary.network)"
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
