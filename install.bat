@echo off
REM Core CLI unified installer (Windows)
REM Served via *.core.help with BunnyCDN edge transformation
REM
REM Usage:
REM   curl -fsSL setup.core.help -o install.bat && install.bat     # Interactive (default)
REM   curl -fsSL ci.core.help -o install.bat && install.bat        # CI/CD
REM   curl -fsSL dev.core.help -o install.bat && install.bat       # Full development
REM   curl -fsSL go.core.help -o install.bat && install.bat        # Go variant
REM   curl -fsSL php.core.help -o install.bat && install.bat       # PHP variant
REM   curl -fsSL agent.core.help -o install.bat && install.bat     # Agent variant
REM
setlocal enabledelayedexpansion

REM === BunnyCDN Edge Variables (transformed at edge based on subdomain) ===
set "MODE={{CORE_MODE}}"
set "VARIANT={{CORE_VARIANT}}"

REM === Fallback for local testing ===
if "!MODE!"=="{{CORE_MODE}}" (
    if defined CORE_MODE (set "MODE=!CORE_MODE!") else (set "MODE=setup")
)
if "!VARIANT!"=="{{CORE_VARIANT}}" (
    if defined CORE_VARIANT (set "VARIANT=!CORE_VARIANT!") else (set "VARIANT=")
)

REM === Configuration ===
set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=latest"
set "REPO=host-uk/core"
set "BINARY=core"
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\core"

REM === Resolve Version ===
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

REM === Create install directory ===
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

REM === Mode dispatch ===
if "%MODE%"=="ci" goto :install_ci
if "%MODE%"=="dev" goto :install_dev
if "%MODE%"=="variant" goto :install_variant
goto :install_setup

:install_setup
echo Installing %BINARY% !VERSION! for Windows...
call :find_archive "" ARCHIVE
call :download_and_extract
call :install_binary
call :verify_install
goto :done

:install_ci
echo Installing %BINARY% !VERSION! (CI)...
call :find_archive "" ARCHIVE

REM Download
curl -fsSL "https://github.com/%REPO%/releases/download/!VERSION!/!ARCHIVE!" -o "%TEMP%\!ARCHIVE!"
if errorlevel 1 (
    echo ERROR: Failed to download !ARCHIVE!
    exit /b 1
)

REM Extract - try System32 first (CI often has admin), else local
powershell -Command "try { Expand-Archive -Force '%TEMP%\!ARCHIVE!' '%TEMP%\core-extract' } catch { exit 1 }"
if errorlevel 1 (
    echo ERROR: Failed to extract archive
    del "%TEMP%\!ARCHIVE!" 2>nul
    exit /b 1
)

move /y "%TEMP%\core-extract\%BINARY%.exe" "C:\Windows\System32\%BINARY%.exe" >nul 2>&1
if errorlevel 1 (
    if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
    move /y "%TEMP%\core-extract\%BINARY%.exe" "%INSTALL_DIR%\%BINARY%.exe"
    echo %PATH% | findstr /i /c:"%INSTALL_DIR%" >nul
    if errorlevel 1 (
        powershell -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';%INSTALL_DIR%', 'User')"
        set "PATH=%PATH%;%INSTALL_DIR%"
    )
)
rmdir /s /q "%TEMP%\core-extract" 2>nul
del "%TEMP%\!ARCHIVE!" 2>nul

%BINARY% --version || exit /b 1
goto :done

:install_dev
echo Installing %BINARY% !VERSION! (full) for Windows...
call :find_archive "" ARCHIVE
call :download_and_extract
call :install_binary
call :verify_install
echo.
echo Full development variant installed. Available commands:
echo   core dev     - Multi-repo workflows
echo   core build   - Cross-platform builds
echo   core release - Build and publish releases
goto :done

:install_variant
echo Installing %BINARY% !VERSION! (%VARIANT% variant) for Windows...
call :find_archive "%VARIANT%" ARCHIVE
call :download_and_extract
call :install_binary
call :verify_install
goto :done

REM === Helper Functions ===

:find_archive
set "_variant=%~1"
set "_result=%~2"

REM Try variant-specific first, then full
if not "%_variant%"=="" (
    set "_try=%BINARY%-%_variant%-windows-amd64.zip"
    curl -fsSLI "https://github.com/%REPO%/releases/download/!VERSION!/!_try!" 2>nul | findstr /r "HTTP/.* [23]0[02]" >nul
    if not errorlevel 1 (
        set "%_result%=!_try!"
        exit /b 0
    )
    echo Using full variant ^(%_variant% variant not available^)
)

set "%_result%=%BINARY%-windows-amd64.zip"
exit /b 0

:download_and_extract
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
exit /b 0

:install_binary
REM Add to PATH using PowerShell (avoids setx 1024 char limit)
echo %PATH% | findstr /i /c:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    powershell -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';%INSTALL_DIR%', 'User')"
    set "PATH=%PATH%;%INSTALL_DIR%"
)
exit /b 0

:verify_install
if not exist "%INSTALL_DIR%\%BINARY%.exe" (
    echo ERROR: Installation failed - binary not found
    exit /b 1
)
"%INSTALL_DIR%\%BINARY%.exe" --version
if errorlevel 1 exit /b 1
exit /b 0

:done
endlocal
