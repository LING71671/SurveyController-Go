param(
    [switch]$IncludeFullStress,
    [switch]$SkipStaticcheck,
    [switch]$SkipStress
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
Push-Location $repoRoot
try {
    Write-Host "== go test ./... =="
    & go test ./...
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    Write-Host "== go vet ./... =="
    & go vet ./...
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    if (-not $SkipStaticcheck) {
        Write-Host "== staticcheck ./... =="
        & go run honnef.co/go/tools/cmd/staticcheck@latest ./...
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    if (-not $SkipStress) {
        Write-Host "== mock stress matrix =="
        $stressArgs = @()
        if (-not $IncludeFullStress) {
            $stressArgs += "-SkipFull"
        }
        & powershell -ExecutionPolicy Bypass -File scripts/mock-stress-matrix.ps1 @stressArgs
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    Write-Host "local verification passed"
}
finally {
    Pop-Location
}
