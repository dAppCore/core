@echo off
REM Core CLI installer - Go development variant (Windows)
REM Usage: curl -fsSL https://core.io.in/go.bat -o go.bat && go.bat
setlocal enabledelayedexpansion

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=latest"
set "REPO=host-uk/core"
set "BINARY=core"
set "VARIANT=go"
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

echo Installing %BINARY% !VERSION! (%VARIANT% variant)...

set "ARCHIVE=%BINARY%-%VARIANT%-windows-amd64.zip"
set "URL=https://github.com/%REPO%/releases/download/!VERSION!/%ARCHIVE%"

curl -fsSLI "%URL%" 2>nul | findstr /r "HTTP/.* [23]0[02]" >nul
if errorlevel 1 (
    set "ARCHIVE=%BINARY%-windows-amd64.zip"
    echo Using full variant (%VARIANT% variant not available^)
)

curl -fsSL "https://github.com/%REPO%/releases/download/!VERSION!/!ARCHIVE!" -o "%TEMP%\!ARCHIVE!"
if errorlevel 1 (
    echo ERROR: Failed to download !ARCHIVE!
    exit /b 1
)

powershell -Command "try { Expand-Archive -Force '%TEMP%\!ARCHIVE!' '%INSTALL_DIR%' } catch { exit 1 }"
if errorlevel 1 (
    echo ERROR: Failed to extract archive
    del "%TEMP%\!ARCHIVE!" 2>nul
    exit /b 1
)
del "%TEMP%\!ARCHIVE!" 2>nul

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

endlocal
