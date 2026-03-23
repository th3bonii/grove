@echo off
setlocal enabledelayedexpansion

echo ========================================
echo GROVE Installation Script
echo ========================================
echo.

REM Check Go
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed
    echo Please install Go 1.23 from: https://go.dev/dl/
    pause
    exit /b 1
)

REM Get script directory
set SCRIPT_DIR=%~dp0
set PROJECT_DIR=%SCRIPT_DIR%..

echo Project: %PROJECT_DIR%
echo.

REM Build binaries
echo Building GROVE...
cd /d "%PROJECT_DIR%"

if not exist "bin" mkdir "bin"

go build -ldflags "-X main.version=1.0.0" -o bin\grove-spec.exe .\cmd\grove-spec
if %ERRORLEVEL% neq 0 (
    echo Error: Failed to build grove-spec
    pause
    exit /b 1
)

go build -ldflags "-X main.version=1.0.0" -o bin\grove-loop.exe .\cmd\grove-loop
if %ERRORLEVEL% neq 0 (
    echo Error: Failed to build grove-loop
    pause
    exit /b 1
)

go build -ldflags "-X main.version=1.0.0" -o bin\grove-opti.exe .\cmd\grove-opti
if %ERRORLEVEL% neq 0 (
    echo Error: Failed to build grove-opti
    pause
    exit /b 1
)

echo [OK] Binaries built

REM Install skills
echo.
echo Installing GROVE skills to OpenCode...

set SKILLS_DIR=%USERPROFILE%\.config\opencode\skills
if not exist "%SKILLS_DIR%" mkdir "%SKILLS_DIR%"

if exist "%PROJECT_DIR%\skills\grove-spec" (
    xcopy /E /I /Y "%PROJECT_DIR%\skills\grove-spec" "%SKILLS_DIR%\grove-spec"
    echo [OK] grove-spec skill installed
)

if exist "%PROJECT_DIR%\skills\grove-loop" (
    xcopy /E /I /Y "%PROJECT_DIR%\skills\grove-loop" "%SKILLS_DIR%\grove-loop"
    echo [OK] grove-loop skill installed
)

if exist "%PROJECT_DIR%\skills\grove-opti" (
    xcopy /E /I /Y "%PROJECT_DIR%\skills\grove-opti" "%SKILLS_DIR%\grove-opti"
    echo [OK] grove-opti skill installed
)

REM Install binaries
echo.
echo Installing binaries...

set BIN_DIR=%USERPROFILE%\bin
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

copy /Y "%PROJECT_DIR%\bin\grove-spec.exe" "%BIN_DIR%\grove-spec.exe"
copy /Y "%PROJECT_DIR%\bin\grove-loop.exe" "%BIN_DIR%\grove-loop.exe"
copy /Y "%PROJECT_DIR%\bin\grove-opti.exe" "%BIN_DIR%\grove-opti.exe"

echo [OK] Binaries installed to %BIN_DIR%

REM Add to PATH if not already
echo %PATH% | findstr /C:"%BIN_DIR%" >nul
if %ERRORLEVEL% neq 0 (
    echo.
    echo Adding %BIN_DIR% to PATH...
    setx PATH "%PATH%;%BIN_DIR%"
    echo [OK] PATH updated (restart terminal to apply)
)

echo.
echo ========================================
echo GROVE installed successfully!
echo ========================================
echo.
echo Quick start:
echo   grove-spec --input ./my-ideas    :: Generate specs
echo   grove-loop                       :: Build from specs
echo   grove-opti "add login button"    :: Optimize prompts
echo.
pause
