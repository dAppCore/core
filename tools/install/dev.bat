@echo off
REM Core CLI installer - Multi-repo development variant (Windows)
REM Usage: curl -fsSL https://core.io.in/dev.bat -o dev.bat && dev.bat
setlocal enabledelayedexpansion

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=latest"
set "REPO=host-uk/core"
set "BINARY=core"
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\core"

if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

if "%VERSION%"=="latest" (
    for /f "tokens=2 delims=:" %%a in ('curl -fsSL "https://api.github.com/repos/%REPO%/releases/latest" ^| findstr "tag_name"') do (
        set "VERSION=%%a"
        set "VERSION=!VERSION:"=!"
        set "VERSION=!VERSION: =!"
        set "VERSION=!VERSION:,=!"
    )
    if "!VERSION!"=="" (
        echo ERROR: Failed to fetch latest version
        exit /b 1
    )
    if "!VERSION!"=="latest" (
        echo ERROR: Failed to resolve version
        exit /b 1
    )
)

echo Installing %BINARY% !VERSION! (full) for Windows...

set "ARCHIVE=%BINARY%-windows-amd64.zip"
curl -fsSL "https://github.com/%REPO%/releases/download/!VERSION!/%ARCHIVE%" -o "%TEMP%\%ARCHIVE%"
if errorlevel 1 (
    echo ERROR: Failed to download %ARCHIVE%
    exit /b 1
)

powershell -Command "try { Expand-Archive -Force '%TEMP%\%ARCHIVE%' '%INSTALL_DIR%' } catch { exit 1 }"
if errorlevel 1 (
    echo ERROR: Failed to extract archive
    del "%TEMP%\%ARCHIVE%" 2>nul
    exit /b 1
)
del "%TEMP%\%ARCHIVE%" 2>nul

REM Add to PATH using PowerShell (avoids setx 1024 char limit)
echo %PATH% | findstr /i /c:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    powershell -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';%INSTALL_DIR%', 'User')"
    set "PATH=%PATH%;%INSTALL_DIR%"
)

if not exist "%INSTALL_DIR%\%BINARY%.exe" (
    echo ERROR: Installation failed - binary not found
    exit /b 1
)

"%INSTALL_DIR%\%BINARY%.exe" --version
if errorlevel 1 exit /b 1

echo.
echo Full development variant installed. Available commands:
echo   core dev     - Multi-repo workflows
echo   core build   - Cross-platform builds
echo   core release - Build and publish releases

endlocal
