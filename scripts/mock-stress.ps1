param(
    [string]$Config = "examples/mock-run.yaml",
    [int]$Target = 1000,
    [int]$Concurrency = 1000,
    [int]$Seed = 7,
    [int]$FailEvery = 0,
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
    if ($Json) {
        $surveyctlArgs += "--json"
    }

    & go run ./cmd/surveyctl @surveyctlArgs
    exit $LASTEXITCODE
}
finally {
    Pop-Location
}
