param(
  [string]$Version = $(if ($env:ERP_SKILL_VERSION) { $env:ERP_SKILL_VERSION } else { "0.1.1" }),
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
  if (-not $line) { throw "校验清单缺少当前安装包，已停止安装。" }
  $expected = ($line -split '\s+')[0].ToLowerInvariant()
  $actual = (Get-FileHash (Join-Path $temporary $archive) -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($expected -ne $actual) { throw "SHA-256 校验失败，已停止安装。" }
  Expand-Archive (Join-Path $temporary $archive) -DestinationPath $temporary
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
  Copy-Item (Join-Path $temporary "erp-agent-skill_${Version}_windows_amd64\erp-skill.exe") (Join-Path $InstallDir "erp-skill.exe") -Force
  Write-Host "已安装：$(Join-Path $InstallDir 'erp-skill.exe')"
  Write-Host ('当前终端加入 PATH：$env:Path = "{0};$env:Path"' -f $InstallDir)
  Write-Host "下一步：运行 erp-skill version，再按安装指南配置 Codex、Claude Code 或 WorkBuddy。"
} finally {
  Remove-Item -Recurse -Force $temporary -ErrorAction SilentlyContinue
}
