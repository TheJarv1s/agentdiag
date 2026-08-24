[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$')]
    [string]$Version
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot

function Require-Match([string]$Path, [string]$Pattern, [string]$Label) {
    $body = Get-Content -Raw (Join-Path $repo $Path)
    if ($body -notmatch $Pattern) {
        throw "$Label does not declare release version $Version."
    }
}

Require-Match 'internal/engine/engine.go' ('const Version = "' + [regex]::Escape($Version) + '"') 'Source version'
Require-Match 'VERSION' ('^' + [regex]::Escape($Version) + '\r?\n?$') 'VERSION file'
Require-Match 'Makefile' ('(?m)^VERSION=' + [regex]::Escape($Version) + '\r?$') 'Release build version'
Require-Match 'README.md' ('\*\*Current version:\*\* ' + [regex]::Escape($Version)) 'English README'
Require-Match 'README.ru.md' ('\*\*Текущая версия:\*\* ' + [regex]::Escape($Version)) 'Russian README'
Require-Match 'CHANGELOG.md' ('(?m)^## \[' + [regex]::Escape($Version) + '\] - \d{4}-\d{2}-\d{2}\r?$') 'CHANGELOG release section'

foreach ($path in 'README.md', 'README.ru.md') {
    $body = Get-Content -Raw (Join-Path $repo $path)
    if ($body -match ('(?is)(candidate|planned|proposed|кандидат|планируем\w*|предлагаем\w*).{0,40}' + [regex]::Escape($Version) + '|' + [regex]::Escape($Version) + '.{0,40}(candidate|planned|proposed|кандидат|планируем\w*|предлагаем\w*)')) {
        throw "$path still marks $Version as a candidate."
    }
}

Write-Output "Release source/docs consistency passed for v$Version."
