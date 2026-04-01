$ErrorActionPreference = 'Stop'

# 配置项
$TargetRepo = 'git@10.80.1.11:ivy-vs-backend/party-pop-idl.git'
$TargetDir = 'proto/game'
$RepoName = 'party-pop-idl'
$CommitMessage = 'feat: update game envelope proto'
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$TempDir = $null

try {
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        throw '错误: 未找到 Git 命令。请安装 Git 并确保其在系统 PATH 中。'
    }

    Write-Host '[+] 获取远程仓库分支列表...'
    $branchLines = git ls-remote --heads $TargetRepo
    if ($LASTEXITCODE -ne 0) {
        throw '错误: 无法获取远程分支列表。'
    }

    $branches = @(
        $branchLines |
            Where-Object { $_ -match 'refs/heads/' } |
            ForEach-Object {
                $parts = $_ -split '\s+'
                if ($parts.Length -ge 2) {
                    $parts[1] -replace '^refs/heads/', ''
                }
            } |
            Where-Object { -not [string]::IsNullOrWhiteSpace($_) }
    )

    if ($branches.Count -eq 0) {
        throw '错误: 未找到任何远程分支。'
    }

    Write-Host ''
    Write-Host '远程仓库分支列表：'
    for ($i = 0; $i -lt $branches.Count; $i++) {
        Write-Host ('{0}. {1}' -f ($i + 1), $branches[$i])
    }

    Write-Host ''
    $branchChoice = Read-Host "请选择分支序号 [1-$($branches.Count)]"
    [int]$branchIndex = 0
    if (-not [int]::TryParse($branchChoice, [ref]$branchIndex) -or $branchIndex -lt 1 -or $branchIndex -gt $branches.Count) {
        throw '错误: 无效的分支序号。'
    }

    $Branch = $branches[$branchIndex - 1]
    Write-Host "已选择分支: $Branch"

    $TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ('{0}.{1}' -f $RepoName, [System.Guid]::NewGuid().ToString('N'))
    $CloneDir = Join-Path $TempDir $RepoName
    New-Item -ItemType Directory -Path $TempDir | Out-Null

    Write-Host '[+] 克隆目标仓库...'
    git clone $TargetRepo --branch $Branch --depth 1 $CloneDir
    if ($LASTEXITCODE -ne 0) {
        throw '错误: 克隆仓库失败。'
    }

    $DestDir = Join-Path $CloneDir $TargetDir
    if (-not (Test-Path $DestDir)) {
        Write-Host '创建目标目录...'
        New-Item -ItemType Directory -Path $DestDir -Force | Out-Null
    }

    $ProtoFiles = Get-ChildItem -Path $ScriptDir -Filter '*.proto' -File
    if ($ProtoFiles.Count -eq 0) {
        throw '错误: 未找到需要上传的 proto 文件。'
    }

    Write-Host '[+] 复制proto文件...'
    Copy-Item -Path $ProtoFiles.FullName -Destination $DestDir -Force

    Push-Location $CloneDir
    try {
        Write-Host '[+] 检查更改...'
        git status

        git add $TargetDir
        if ($LASTEXITCODE -ne 0) {
            throw '错误: 添加变更失败。'
        }

        $HasChanges = $true
        git diff --cached --quiet
        if ($LASTEXITCODE -eq 0) {
            Write-Host '警告: 没有需要提交的更改，直接清理临时目录...'
            $HasChanges = $false
        }
        elseif ($LASTEXITCODE -ne 1) {
            throw '错误: 检查暂存区变更失败。'
        }

        if ($HasChanges) {
            Write-Host '[+] 提交更改...'
            git commit -m $CommitMessage
            if ($LASTEXITCODE -ne 0) {
                throw '错误: 提交更改失败。'
            }

            Write-Host '[+] 推送更改...'
            git push origin $Branch
            if ($LASTEXITCODE -ne 0) {
                throw '错误: 推送更改失败。'
            }
        }
    }
    finally {
        Pop-Location
    }

    Write-Host '操作完成！'
}
catch {
    Write-Error $_.Exception.Message
    exit 1
}
finally {
    if ($TempDir -and (Test-Path $TempDir)) {
        Write-Host '[+] 清理临时目录...'
        Remove-Item -Path $TempDir -Recurse -Force
    }
}
