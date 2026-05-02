param(
    [switch]$IncludeFullStress,
    [switch]$IncludeWJXHTTPDryRunStress,
    [switch]$SkipGoChecks,
    [switch]$SkipStaticcheck,
    [switch]$SkipStress
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
. (Join-Path $PSScriptRoot "lib/powershell.ps1")
$powerShellCommand = Resolve-SurveyControllerPowerShell
Push-Location $repoRoot
try {
    if (-not $SkipGoChecks) {
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
    }

    if (-not $SkipGoChecks -and -not $SkipStaticcheck) {
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
        $commandArgs = New-SurveyControllerPowerShellFileArgs -Command $powerShellCommand -File "scripts/mock-stress-matrix.ps1" -Arguments $stressArgs
        & $powerShellCommand.Source @commandArgs
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    if ($IncludeWJXHTTPDryRunStress) {
        Write-Host "== wjx http dry-run stress matrix =="
        $wjxStressArgs = @()
        if (-not $IncludeFullStress) {
            $wjxStressArgs += "-SkipFull"
        }
        $commandArgs = New-SurveyControllerPowerShellFileArgs -Command $powerShellCommand -File "scripts/wjx-http-dryrun-stress-matrix.ps1" -Arguments $wjxStressArgs
        & $powerShellCommand.Source @commandArgs
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    Write-Host "local verification passed"
}
finally {
    Pop-Location
}
