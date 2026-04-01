$ErrorActionPreference = 'Stop'

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$OutputDir = Join-Path $ScriptDir 'envelope'

try {
    if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
        throw '错误: 未找到 Docker 命令。请先安装 Docker 并确保其在系统 PATH 中。'
    }

    $ProtoFiles = Get-ChildItem -Path $ScriptDir -Filter '*.proto' -File
    if ($ProtoFiles.Count -eq 0) {
        throw '错误: 当前目录下未找到 .proto 文件。'
    }

    Write-Host '[+] 使用 Docker 生成 Go 代码...'
    docker run --rm -v "${ScriptDir}:/defs" rvolosatovs/protoc --proto_path=/defs --go_out=/defs /defs/*.proto
    if ($LASTEXITCODE -ne 0) {
        throw '错误: Docker 生成 Go 代码失败。'
    }

    if (-not (Test-Path $OutputDir)) {
        throw "错误: 未找到生成目录 $OutputDir 。"
    }

    $GeneratedFiles = Get-ChildItem -Path $OutputDir -Filter '*.go' -File
    if ($GeneratedFiles.Count -eq 0) {
        throw "错误: 生成目录 $OutputDir 下未找到 .go 文件。"
    }

    Write-Host '[+] 移动生成文件到当前目录...'
    Move-Item -Path $GeneratedFiles.FullName -Destination $ScriptDir -Force

    Write-Host '[+] 清理临时目录...'
    Remove-Item -Path $OutputDir -Recurse -Force

    Write-Host '操作完成！'
}
catch {
    Write-Error $_.Exception.Message
    exit 1
}
