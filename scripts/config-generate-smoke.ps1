param(
    [string]$WJXFixture = "internal/provider/wjx/testdata/survey.html",
    [string]$TencentFixture = "internal/provider/tencent/testdata/survey_api.json",
    [string]$CredamoFixture = "internal/provider/credamo/testdata/snapshot.json"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot

function Invoke-ConfigGenerateSmoke {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Provider,
        [Parameter(Mandatory = $true)]
        [string]$Fixture,
        [Parameter(Mandatory = $true)]
        [string]$URL,
        [Parameter(Mandatory = $true)]
        [string[]]$ExpectedPatterns
    )

    Write-Host "== config generate $Provider =="
    $output = & go run ./cmd/surveyctl config generate --provider $Provider --fixture $Fixture --url $URL
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }

    $text = $output -join "`n"
    foreach ($pattern in $ExpectedPatterns) {
        if ($text -notmatch $pattern) {
            Write-Error "config generate $Provider did not contain expected pattern '$pattern': $text"
        }
    }
}

Push-Location $repoRoot
try {
    Invoke-ConfigGenerateSmoke `
        -Provider "wjx" `
        -Fixture $WJXFixture `
        -URL "https://www.wjx.cn/vm/example.aspx" `
        -ExpectedPatterns @(
            "provider:\s*wjx",
            "kind:\s*text",
            "mode:\s*fixed",
            "sample answer",
            "kind:\s*matrix",
            "matrix_weights"
        )

    Invoke-ConfigGenerateSmoke `
        -Provider "tencent" `
        -Fixture $TencentFixture `
        -URL "https://wj.qq.com/s2/example" `
        -ExpectedPatterns @(
            "provider:\s*tencent",
            "kind:\s*dropdown",
            "kind:\s*matrix",
            "matrix_weights",
            "option_id:\s*q2_a"
        )

    Invoke-ConfigGenerateSmoke `
        -Provider "credamo" `
        -Fixture $CredamoFixture `
        -URL "https://www.credamo.com/answer.html#/s/demo" `
        -ExpectedPatterns @(
            "provider:\s*credamo",
            "id:\s*question-10",
            "kind:\s*text",
            "mode:\s*fixed",
            "sample answer",
            "option_id:\s*""1"""
        )

    Write-Host "config generate smoke passed"
}
finally {
    Pop-Location
}
