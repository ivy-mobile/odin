@echo off
chcp 65001

setlocal enabledelayedexpansion

:: 配置项
set "TARGET_REPO=git@10.80.1.11:ivy-vs-backend/party-pop-idl.git"
set "TARGET_DIR=proto/game"
set "REPO_NAME=party-pop-idl"
set "TEMP_BRANCH_FILE=%~dp0temp_branches.txt"

:: 检查Git是否安装
where git >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 未找到Git命令。请安装Git并确保其在系统PATH中。
    pause
    exit /b 1
)

:: 获取远程仓库分支列表
echo [+] 获取远程仓库分支列表...
git ls-remote --heads "%TARGET_REPO%" > "%TEMP_BRANCH_FILE%"
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 无法获取远程分支列表。
    if exist "%TEMP_BRANCH_FILE%" del /q "%TEMP_BRANCH_FILE%"
    pause
    exit /b 1
)

:: 显示分支列表并让用户选择
echo. & echo 远程仓库分支列表：
set "branch_count=0"
for /f "tokens=2" %%b in ('git ls-remote --heads "%TARGET_REPO%" ^| find "refs/heads/"') do (
    set "branch_name=%%b"
    set "branch_name=!branch_name:refs/heads/=!"
    set /a "branch_count+=1"
    set "branch_!branch_count!=!branch_name!"
    echo !branch_count!. !branch_name!
)


:: 如果没有分支
echo.
if %branch_count% EQU 0 (
    echo 错误: 未找到任何远程分支。
    if exist "%TEMP_BRANCH_FILE%" del /q "%TEMP_BRANCH_FILE%"
    pause
    exit /b 1
)

:: 用户选择分支
set /p "branch_choice=请选择分支序号 [1-%branch_count%]: "

:: 验证用户输入
echo.
set "valid_choice=false"
for /l %%i in (1,1,%branch_count%) do (
    if "%branch_choice%" EQU "%%i" (
        set "valid_choice=true"
        set "BRANCH=!branch_%%i!"
    )
)

:: 如果输入无效
echo.
if not "%valid_choice%" == "true" (
    echo 错误: 无效的分支序号。
    if exist "%TEMP_BRANCH_FILE%" del /q "%TEMP_BRANCH_FILE%"
    pause
    exit /b 1
)

:: 显示选择的分支
echo 已选择分支: %BRANCH%

:: 清理临时文件
if exist "%TEMP_BRANCH_FILE%" del /q "%TEMP_BRANCH_FILE%"

:: 克隆目标仓库
echo [+] 克隆目标仓库...
git clone "%TARGET_REPO%" --branch "%BRANCH%" --depth 1
if %ERRORLEVEL% NEQ 0 (
    echo 错误: 克隆仓库失败。
    pause
    exit /b 1
)

:: 确保目标目录存在
if not exist "%REPO_NAME%\%TARGET_DIR%" (
    echo 创建目标目录...
    mkdir "%REPO_NAME%\%TARGET_DIR%"
    if %ERRORLEVEL% NEQ 0 (
        echo 错误: 无法创建目标目录。
        pause
        exit /b 1
    )
)

:: 复制proto文件
echo [+] 复制proto文件...
copy "*.proto" "%REPO_NAME%\%TARGET_DIR%\" /y
if %ERRORLEVEL% NEQ 0 (
    echo 警告: 复制proto文件可能不完全成功。
)

:: 提交并推送更改
cd "%REPO_NAME%"
echo [+] 检查更改...

git status
git add "%TARGET_DIR%\*.proto"

echo [+] 提交更改...
git commit -m "feat: update game envelope proto"
if %ERRORLEVEL% NEQ 0 (
    echo 警告: 没有需要提交的更改,直接清理临时目录...
) else (
    echo [+] 推送更改...
    git push origin "%BRANCH%"
    if %ERRORLEVEL% NEQ 0 (
        echo 错误: 推送更改失败。
    )
)

:: 清理
cd /d "%~dp0"
echo [+] 清理临时目录...
rmdir /s /q "%REPO_NAME%"
if %ERRORLEVEL% NEQ 0 (
    echo 警告: 无法删除临时目录。
    pause
    exit /b 1
)

echo 操作完成！
pause
endlocal