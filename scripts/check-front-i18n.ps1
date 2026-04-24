param(
    [string]$Root = "frontend/src"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

if (-not (Test-Path -LiteralPath $Root)) {
    throw "path not found: $Root"
}

$quotedRegex = [regex]'["''][^"''\r\n]*[\p{IsCJKUnifiedIdeographs}][^"''\r\n]*["'']'
$templateRegex = [regex]'>\s*[^<{/\r\n]*[\p{IsCJKUnifiedIdeographs}][^<{/\r\n]*<'
$commentRegex = [regex]'^(//|/\*|\*|\*/|<!--|-->)'

$files = Get-ChildItem -LiteralPath $Root -Recurse -File |
    Where-Object {
        $_.Extension -in ".vue", ".ts" -and
        $_.FullName -notmatch '[\\/](i18n|dist|node_modules)[\\/]'
    }

$findings = [System.Collections.Generic.List[object]]::new()

foreach ($file in $files) {
    $lines = Get-Content -LiteralPath $file.FullName -Encoding UTF8
    $inHtmlComment = $false
    for ($idx = 0; $idx -lt $lines.Length; $idx++) {
        $line = $lines[$idx]
        $trimmed = $line.Trim()

        if ($inHtmlComment) {
            if ($line.Contains("-->")) {
                $inHtmlComment = $false
            }
            continue
        }

        if ($line.Contains("<!--")) {
            if (-not $line.Contains("-->")) {
                $inHtmlComment = $true
            }
            continue
        }

        if ($trimmed -eq "" -or $commentRegex.IsMatch($trimmed)) {
            continue
        }

        $kinds = [System.Collections.Generic.List[string]]::new()
        if ($quotedRegex.IsMatch($line)) {
            $kinds.Add("quoted")
        }
        if ($templateRegex.IsMatch($line)) {
            $kinds.Add("template")
        }
        if ($kinds.Count -eq 0) {
            continue
        }

        $relative = (Resolve-Path -LiteralPath $file.FullName -Relative) -replace '^[.][\\/]', ''
        $findings.Add([pscustomobject]@{
            Path = $relative
            Line = $idx + 1
            Kind = ($kinds -join ",")
            Text = $trimmed
        })
    }
}

if ($findings.Count -eq 0) {
    Write-Host "[check-front-i18n] no suspected hard-coded Chinese literals found."
    exit 0
}

$findings |
    Sort-Object Path, Line |
    Format-Table -AutoSize

Write-Host "[check-front-i18n] found $($findings.Count) suspected literals."
exit 1
