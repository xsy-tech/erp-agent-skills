param(
  [string]$Version = $(if ($env:ERP_SKILL_VERSION) { $env:ERP_SKILL_VERSION } else { "0.1.2" }),
  [string]$BaseUrl = $env:ERP_SKILL_RELEASE_BASE_URL,
  [string]$InstallDir = $(if ($env:ERP_SKILL_INSTALL_DIR) { $env:ERP_SKILL_INSTALL_DIR } else { Join-Path $HOME ".local\bin" })
)

$ErrorActionPreference = "Stop"
if (-not $BaseUrl) { $BaseUrl = "https://github.com/xsy-tech/erp-agent-skills/releases/download/v$Version" }
$archive = "erp-agent-skill_${Version}_windows_amd64.zip"
$temporary = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $temporary | Out-Null
try {
  Invoke-WebRequest "$BaseUrl/$archive" -UseBasicParsing -OutFile (Join-Path $temporary $archive)
  Invoke-WebRequest "$BaseUrl/checksums.txt" -UseBasicParsing -OutFile (Join-Path $temporary "checksums.txt")
  $line = Get-Content (Join-Path $temporary "checksums.txt") | Where-Object { $_ -match [regex]::Escape($archive) } | Select-Object -First 1
  if (-not $line) { throw "校验清单缺少当前安装包，已停止安装。 / Archive checksum missing; installation stopped." }
  $expected = ($line -split '\s+')[0].ToLowerInvariant()
  $actual = (Get-FileHash (Join-Path $temporary $archive) -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($expected -ne $actual) { throw "SHA-256 校验失败，已停止安装。 / SHA-256 verification failed; installation stopped." }
  Expand-Archive (Join-Path $temporary $archive) -DestinationPath $temporary
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  Copy-Item (Join-Path $temporary "erp-agent-skill_${Version}_windows_amd64\erp-skill.exe") (Join-Path $InstallDir "erp-skill.exe") -Force
  Write-Host "已安装 / Installed: $(Join-Path $InstallDir 'erp-skill.exe')"
  Write-Host ('当前终端加入 PATH / Add to this terminal PATH: $env:Path = "{0};$env:Path"' -f $InstallDir)
  Write-Host "下一步 / Next: erp-skill version；然后按安装指南配置客户端 / then follow the installation guide to configure your agent."
} finally {
  Remove-Item -Recurse -Force $temporary -ErrorAction SilentlyContinue
}
