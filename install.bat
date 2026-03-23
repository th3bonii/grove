@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo.
echo  ╔═══════════════════════════════════════════════════════╗
echo  ║           GROVE - Instalador v1.3.0                 ║
echo  ║     Intelligent development tools for OpenCode       ║
echo  ╚═══════════════════════════════════════════════════════╝
echo.

REM Get script directory
set SCRIPT_DIR=%~dp0
set GROVE_DIR=%SCRIPT_DIR%

REM Check if binaries exist
if not exist "%GROVE_DIR%release\grove-spec.exe" (
    echo [ERROR] Binaries not found in release folder
    echo.
    echo Please download from: https://github.com/th3bonii/grove/releases
    echo Or run: go build ./cmd/grove-spec
    pause
    exit /b 1
)

REM Create installation directory
set INSTALL_DIR=%USERPROFILE%\grove
echo [1/5] Creating installation directory...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

REM Copy binaries
echo [2/5] Installing binaries...
copy /Y "%GROVE_DIR%release\grove-spec.exe" "%INSTALL_DIR%\grove-spec.exe" >nul
copy /Y "%GROVE_DIR%release\grove-loop.exe" "%INSTALL_DIR%\grove-loop.exe" >nul
copy /Y "%GROVE_DIR%release\grove-opti.exe" "%INSTALL_DIR%\grove-opti.exe" >nul
echo       ✓ grove-spec.exe
echo       ✓ grove-loop.exe
echo       ✓ grove-opti.exe

REM Add to PATH
echo [3/5] Adding to PATH...
echo %PATH% | findstr /C:"%INSTALL_DIR%" >nul
if %ERRORLEVEL% neq 0 (
    setx PATH "%PATH%;%INSTALL_DIR%" >nul 2>&1
    echo       ✓ Added to PATH
) else (
    echo       ✓ Already in PATH
)

REM Install skills
echo [4/5] Installing OpenCode skills...
set SKILLS_DIR=%USERPROFILE%\.config\opencode\skills
if not exist "%SKILLS_DIR%" mkdir "%SKILLS_DIR%"

if exist "%GROVE_DIR%skills\grove-spec" (
    xcopy /E /I /Y "%GROVE_DIR%skills\grove-spec" "%SKILLS_DIR%\grove-spec" >nul
    echo       ✓ grove-spec skill
)
if exist "%GROVE_DIR%skills\grove-loop" (
    xcopy /E /I /Y "%GROVE_DIR%skills\grove-loop" "%SKILLS_DIR%\grove-loop" >nul
    echo       ✓ grove-loop skill
)
if exist "%GROVE_DIR%skills\grove-opti" (
    xcopy /E /I /Y "%GROVE_DIR%skills\grove-opti" "%SKILLS_DIR%\grove-opti" >nul
    echo       ✓ grove-opti skill
)

REM Create shortcuts
echo [5/5] Creating shortcuts...
(
echo @echo off
echo "%INSTALL_DIR%\grove-spec.exe" %%*
) > "%INSTALL_DIR%\grove.bat"
echo       ✓ grove.bat shortcut

echo.
echo  ╔═══════════════════════════════════════════════════════╗
echo  ║           ✅ INSTALACIÓN COMPLETADA                  ║
echo  ╚═══════════════════════════════════════════════════════╝
echo.
echo  Ubicación: %INSTALL_DIR%
echo.
echo  COMANDOS:
echo  ─────────────────────────────────────────────────────
echo  grove-spec --input ./ideas     Ideas → Especificaciones
echo  grove-loop                     Especificaciones → Código
echo  grove-opti "prompt"            Optimizar prompts
echo.
echo  EJEMPLO RÁPIDO:
echo  ─────────────────────────────────────────────────────
echo  mkdir mi-proyecto\ideas
echo  echo # Mi App > mi-proyecto\ideas\README.md
echo  cd mi-proyecto
echo  grove-spec --input ./ideas
echo.
echo  ⚠ IMPORTANTE: Cierra y abre una nueva terminal
echo    para que los cambios en PATH surtan efecto
echo.
pause
