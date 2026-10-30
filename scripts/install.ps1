param(
  [string]$Version = $(if ($env:ERP_SKILL_VERSION) { $env:ERP_SKILL_VERSION } else { "0.1.0" }),
  [string]$BaseUrl = $env:ERP_SKILL_RELEASE_BASE_URL,
  [string]$InstallDir = $(if ($env:ERP_SKILL_INSTALL_DIR) { $env:ERP_SKILL_INSTALL_DIR } else { Join-Path $HOME ".local\bin" })
)

$ErrorActionPreference = "Stop"
if (-not $BaseUrl) { throw "Set ERP_SKILL_RELEASE_BASE_URL to the release directory URL" }
$archive = "erp-agent-skill_${Version}_windows_amd64.zip"
$temporary = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $temporary | Out-Null
try {
  Invoke-WebRequest "$BaseUrl/$archive" -OutFile (Join-Path $temporary $archive)
  Invoke-WebRequest "$BaseUrl/checksums.txt" -OutFile (Join-Path $temporary "checksums.txt")
  $line = Get-Content (Join-Path $temporary "checksums.txt") | Where-Object { $_ -match [regex]::Escape($archive) } | Select-Object -First 1
  if (-not $line) { throw "Archive checksum is missing" }
  $expected = ($line -split '\s+')[0].ToLowerInvariant()
  $actual = (Get-FileHash (Join-Path $temporary $archive) -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($expected -ne $actual) { throw "Checksum verification failed" }
  Expand-Archive (Join-Path $temporary $archive) -DestinationPath $temporary
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  Copy-Item (Join-Path $temporary "erp-agent-skill_${Version}_windows_amd64\erp-skill.exe") (Join-Path $InstallDir "erp-skill.exe") -Force
  Write-Host "Installed $(Join-Path $InstallDir 'erp-skill.exe')"
  Write-Host "Next: pipe your personal key to erp-skill install codex --url <MCP_URL> --key-stdin"
} finally {
  Remove-Item -Recurse -Force $temporary -ErrorAction SilentlyContinue
}
