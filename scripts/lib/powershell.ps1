function Resolve-SurveyControllerPowerShell {
    $command = Get-Command pwsh -ErrorAction SilentlyContinue
    if ($null -eq $command) {
        $command = Get-Command powershell -ErrorAction SilentlyContinue
    }
    if ($null -eq $command) {
        throw "PowerShell executable was not found."
    }
    return $command
}

function New-SurveyControllerPowerShellFileArgs {
    param(
        [Parameter(Mandatory = $true)]
        [object]$Command,
        [Parameter(Mandatory = $true)]
        [string]$File,
        [object[]]$Arguments = @()
    )

    $commandArgs = @("-NoProfile")
    if ($Command.Name -like "powershell*") {
        $commandArgs += @("-ExecutionPolicy", "Bypass")
    }
    $commandArgs += @("-File", $File)
    $commandArgs += $Arguments
    return $commandArgs
}
