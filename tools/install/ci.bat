@echo off
REM Core CLI installer for Windows CI environments
REM Usage: curl -fsSL https://core.io.in/ci.bat -o ci.bat && ci.bat
setlocal enabledelayedexpansion

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=latest"
set "REPO=host-uk/core"
set "BINARY=core"

if "%VERSION%"=="latest" (
    for /f "tokens=2 delims=:" %%a in ('curl -fsSL "https://api.github.com/repos/%REPO%/releases/latest" ^| findstr "tag_name"') do (
        set "VERSION=%%a"
        set "VERSION=!VERSION:"=!"
        set "VERSION=!VERSION: =!"
        set "VERSION=!VERSION:,=!"
    )
    if "!VERSION!"=="latest" (
        echo ERROR: Failed to fetch latest version from GitHub API
        exit /b 1
    )
    if "!VERSION!"=="" (
        echo ERROR: Failed to fetch latest version from GitHub API
        exit /b 1
    )
)

echo Installing %BINARY% !VERSION!...

set "ARCHIVE=%BINARY%-windows-amd64.zip"
curl -fsSL "https://github.com/%REPO%/releases/download/!VERSION!/%ARCHIVE%" -o "%TEMP%\%ARCHIVE%"
if errorlevel 1 (
    echo ERROR: Failed to download %ARCHIVE%
    exit /b 1
)

powershell -Command "try { Expand-Archive -Force '%TEMP%\%ARCHIVE%' '%TEMP%\core-extract' } catch { exit 1 }"
if errorlevel 1 (
    echo ERROR: Failed to extract archive
    del "%TEMP%\%ARCHIVE%" 2>nul
    exit /b 1
)

REM Try System32 first (CI runners often have admin), else use local programs
move /y "%TEMP%\core-extract\%BINARY%.exe" "C:\Windows\System32\%BINARY%.exe" >nul 2>&1
if errorlevel 1 (
    if not exist "%LOCALAPPDATA%\Programs" mkdir "%LOCALAPPDATA%\Programs"
    move /y "%TEMP%\core-extract\%BINARY%.exe" "%LOCALAPPDATA%\Programs\%BINARY%.exe"
    set "PATH=%LOCALAPPDATA%\Programs;%PATH%"
    echo NOTE: Installed to %LOCALAPPDATA%\Programs
)
rmdir /s /q "%TEMP%\core-extract" 2>nul
del "%TEMP%\%ARCHIVE%" 2>nul

%BINARY% --version || exit /b 1

endlocal
