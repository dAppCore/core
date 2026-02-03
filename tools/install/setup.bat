@echo off
REM Core CLI installer for Windows
REM Usage: curl -fsSL https://core.io.in/setup.bat -o setup.bat && setup.bat

setlocal enabledelayedexpansion

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=latest"
set "REPO=host-uk/core"
set "BINARY=core"
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\core"

echo [94m>>>[0m Installing Core CLI for Windows...

REM Create install directory
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

REM Resolve latest version if needed
if "%VERSION%"=="latest" (
    echo [94m>>>[0m Fetching latest version...
    for /f "tokens=2 delims=:" %%a in ('curl -fsSL "https://api.github.com/repos/%REPO%/releases/latest" ^| findstr "tag_name"') do (
        set "VERSION=%%a"
        set "VERSION=!VERSION:"=!"
        set "VERSION=!VERSION: =!"
        set "VERSION=!VERSION:,=!"
    )
    if "!VERSION!"=="" (
        echo [91m>>>[0m Failed to fetch latest version
        exit /b 1
    )
    if "!VERSION!"=="latest" (
        echo [91m>>>[0m Failed to resolve version
        exit /b 1
    )
)

echo [94m>>>[0m Installing %BINARY% !VERSION!...

REM Download archive
set "ARCHIVE=%BINARY%-windows-amd64.zip"
set "DOWNLOAD_URL=https://github.com/%REPO%/releases/download/!VERSION!/%ARCHIVE%"
set "TMP_FILE=%TEMP%\%ARCHIVE%"

echo [94m>>>[0m Downloading %ARCHIVE%...
curl -fsSL "%DOWNLOAD_URL%" -o "%TMP_FILE%"
if errorlevel 1 (
    echo [91m>>>[0m Failed to download %DOWNLOAD_URL%
    exit /b 1
)

REM Extract
echo [94m>>>[0m Extracting...
powershell -Command "try { Expand-Archive -Force '%TMP_FILE%' '%INSTALL_DIR%' } catch { exit 1 }"
if errorlevel 1 (
    echo [91m>>>[0m Failed to extract archive
    del "%TMP_FILE%" 2>nul
    exit /b 1
)
del "%TMP_FILE%" 2>nul

REM Add to PATH using PowerShell (avoids setx 1024 char limit)
echo %PATH% | findstr /i /c:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    echo [94m>>>[0m Adding to PATH...
    powershell -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';%INSTALL_DIR%', 'User')"
    set "PATH=%PATH%;%INSTALL_DIR%"
)

REM Verify
if not exist "%INSTALL_DIR%\%BINARY%.exe" (
    echo [91m>>>[0m Installation failed - binary not found
    exit /b 1
)

"%INSTALL_DIR%\%BINARY%.exe" --version
if errorlevel 1 (
    echo [91m>>>[0m Installation verification failed
    exit /b 1
)

echo [92m>>>[0m Installed successfully!
echo.
echo [90mRestart your terminal to use '%BINARY%' command[0m

endlocal
