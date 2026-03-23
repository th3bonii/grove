@echo off
setlocal enabledelayedexpansion

echo ========================================
echo GROVE Installation Script (Windows)
echo ========================================
echo.

REM Check Go
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Go is not installed
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
echo [1/5] Building GROVE binaries...
cd /d "%PROJECT_DIR%"

if not exist "bin" mkdir "bin"

go build -ldflags "-X main.version=1.2.0" -o bin\grove-spec.exe .\cmd\grove-spec
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to build grove-spec
    pause
    exit /b 1
)

go build -ldflags "-X main.version=1.2.0" -o bin\grove-loop.exe .\cmd\grove-loop
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to build grove-loop
    pause
    exit /b 1
)

go build -ldflags "-X main.version=1.2.0" -o bin\grove-opti.exe .\cmd\grove-opti
if %ERRORLEVEL% neq 0 (
    echo [ERROR] Failed to build grove-opti
    pause
    exit /b 1
)

echo [OK] Binaries built

REM Install skills
echo.
echo [2/5] Installing GROVE skills to OpenCode...

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
echo [3/5] Installing binaries...

set BIN_DIR=%USERPROFILE%\go\bin
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

copy /Y "%PROJECT_DIR%\bin\grove-spec.exe" "%BIN_DIR%\grove-spec.exe"
copy /Y "%PROJECT_DIR%\bin\grove-loop.exe" "%BIN_DIR%\grove-loop.exe"
copy /Y "%PROJECT_DIR%\bin\grove-opti.exe" "%BIN_DIR%\grove-opti.exe"

echo [OK] Binaries installed to %BIN_DIR%

REM Configure OpenCode agents
echo.
echo [4/5] Configuring OpenCode agents...

set OPENCODE_CONFIG=%USERPROFILE%\.config\opencode\opencode.json

if not exist "%OPENCODE_CONFIG%" (
    echo [ERROR] opencode.json not found at %OPENCODE_CONFIG%
    echo Please ensure OpenCode is installed
    goto :skip_commands
)

REM Check if agents already exist
findstr /C:"\"grove-spec\"" "%OPENCODE_CONFIG%" >nul
if %ERRORLEVEL% equ 0 (
    echo [SKIP] grove-spec agent already configured
    goto :skip_grove_spec
)

REM Backup original config
copy /Y "%OPENCODE_CONFIG%" "%OPENCODE_CONFIG%.bak"

REM Add grove-spec agent using PowerShell
powershell -Command "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; $config = Get-Content '%OPENCODE_CONFIG%' -Raw | ConvertFrom-Json; if (-not $config.agent.PSObject.Properties.Name.Contains('grove-spec')) { $config.agent | Add-Member -NotePropertyName 'grove-spec' -NotePropertyValue ([ordered]@{description='Generate specifications from ideas'; mode='primary'; model='opencode/big-pickle'; prompt='You are GROVE Spec. Read your skill file at ~/\.config\/opencode\/skills\/grove-spec\/SKILL.md'; tools=@{bash=$true; edit=$true; read=$true; write=$true}; permission=@{bash='allow'; edit='allow'; read='allow'; write='allow'}}) | Out-Null; $config | ConvertTo-Json -Depth 10 | Set-Content '%OPENCODE_CONFIG%' -Encoding UTF8 }; Write-Host '[OK] grove-spec agent configured'"

:skip_grove_spec

findstr /C:"\"grove-opti\"" "%OPENCODE_CONFIG%" >nul
if %ERRORLEVEL% equ 0 (
    echo [SKIP] grove-opti agent already configured
    goto :skip_grove_opti
)

powershell -Command "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; $config = Get-Content '%OPENCODE_CONFIG%' -Raw | ConvertFrom-Json; if (-not $config.agent.PSObject.Properties.Name.Contains('grove-opti')) { $config.agent | Add-Member -NotePropertyName 'grove-opti' -NotePropertyValue ([ordered]@{description='Optimize prompts for AI'; mode='primary'; model='opencode/big-pickle'; prompt='You are GROVE Opti Prompt. Read your skill file at ~/\.config\/opencode\/skills\/grove-opti\/SKILL.md'; tools=@{bash=$true; edit=$true; read=$true; write=$true}; permission=@{bash='allow'; edit='allow'; read='allow'; write='allow'}}) | Out-Null; $config | ConvertTo-Json -Depth 10 | Set-Content '%OPENCODE_CONFIG%' -Encoding UTF8 }; Write-Host '[OK] grove-opti agent configured'"

:skip_grove_opti

findstr /C:"\"grove-loop\"" "%OPENCODE_CONFIG%" >nul
if %ERRORLEVEL% equ 0 (
    echo [SKIP] grove-loop agent already configured
    goto :skip_grove_loop
)

powershell -Command "[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; $config = Get-Content '%OPENCODE_CONFIG%' -Raw | ConvertFrom-Json; if (-not $config.agent.PSObject.Properties.Name.Contains('grove-loop')) { $config.agent | Add-Member -NotePropertyName 'grove-loop' -NotePropertyValue ([ordered]@{description='Execute autonomous build loops'; mode='primary'; model='opencode/big-pickle'; prompt='You are GROVE Ralph Loop. Read your skill file at ~/\.config\/opencode\/skills\/grove-loop\/SKILL.md'; tools=@{bash=$true; edit=$true; read=$true; write=$true}; permission=@{bash='allow'; edit='allow'; read='allow'; write='allow'}}) | Out-Null; $config | ConvertTo-Json -Depth 10 | Set-Content '%OPENCODE_CONFIG%' -Encoding UTF8 }; Write-Host '[OK] grove-loop agent configured'"

:skip_grove_loop

:skip_commands

REM Verify installation
echo.
echo [5/5] Verifying installation...

where grove-spec >nul 2>nul
if %ERRORLEVEL% equ 0 (
    echo [OK] grove-spec command available
) else (
    echo [WARN] grove-spec not in PATH
)

echo.
echo ========================================
echo GROVE installed successfully!
echo ========================================
echo.
echo Commands available in terminal:
echo   grove-spec --input ./my-ideas   :: Generate specifications
echo   grove-loop                       :: Build from specs
echo   grove-opti "add login"          :: Optimize prompts
echo.
echo OpenCode Commands (restart OpenCode to see):
echo   /grove-spec  - Generate specifications
echo   /grove-opti  - Optimize prompts
echo   /grove-loop  - Execute build loops
echo.
echo Quick test: grove-spec --help
echo.
pause
