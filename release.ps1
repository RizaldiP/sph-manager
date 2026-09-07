param(
    [string]$Notes = ""
)

$ErrorActionPreference = "Stop"
$root = $PSScriptRoot
Set-Location $root

Write-Host "Membangun aplikasi (wails build)..."
wails build
if ($LASTEXITCODE -ne 0) {
    throw "wails build gagal (exit code $LASTEXITCODE)"
}

$appGo = Get-Content "$root\app.go" -Raw
$m = [regex]::Match($appGo, 'const\s+appVersion\s*=\s*"([^"]+)"')
if (-not $m.Success) {
    throw "Versi tidak ditemukan di app.go"
}
$version = $m.Groups[1].Value

$exe = "$root\build\bin\SPHManager.exe"
$outDir = "$root\build\release"
$out = "$outDir\SPHManager.exe"
New-Item -ItemType Directory -Force -Path $outDir | Out-Null
Copy-Item $exe $out -Force

# Sisipkan trailer metadata versi di ekor file (dibaca aplikasi saat update).
$meta = @{ version = $version; notes = $Notes } | ConvertTo-Json -Compress
$block = "`nSPHMANAGER-META-BEGIN`n$meta`nSPHMANAGER-META-END`n"
$bytes = [System.Text.Encoding]::UTF8.GetBytes($block)
$fs = [System.IO.File]::Open($out, [System.IO.FileMode]::Append)
try {
    $fs.Write($bytes, 0, $bytes.Length)
} finally {
    $fs.Close()
}

Write-Host "Release $version siap: $out"
Write-Host "Upload/unggah file ini via Google Drive -> Manage versions -> Upload new version untuk mengganti versi (link tidak berubah)."