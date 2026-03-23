@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

echo.
echo  ╔═══════════════════════════════════════════════════════╗
echo  ║           GROVE - Instalador v1.3.0                 ║
echo  ╚═══════════════════════════════════════════════════════╝
echo.

REM Create temp directory
set TEMP_DIR=%TEMP%\grove-install
if exist "%TEMP_DIR%" rmdir /S /Q "%TEMP_DIR%"
mkdir "%TEMP_DIR%"

echo [1/4] Descargando GROVE...

REM Download using PowerShell
powershell -Command "[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; Invoke-WebRequest -Uri 'https://github.com/th3bonii/grove/releases/latest/download/grove-windows.zip' -OutFile '%TEMP_DIR%\grove-windows.zip'"

if not exist "%TEMP_DIR%\grove-windows.zip" (
    echo [ERROR] No se pudo descargar
    echo Descarga manual: https://github.com/th3bonii/grove/releases/latest
    pause
    exit /b 1
)

echo [2/4] Extrayendo...
powershell -Command "Expand-Archive -Path '%TEMP_DIR%\grove-windows.zip' -DestinationPath '%TEMP_DIR%\grove' -Force"

echo [3/4] Instalando...
set INSTALL_DIR=%USERPROFILE%\grove
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

copy /Y "%TEMP_DIR%\grove\grove-spec.exe" "%INSTALL_DIR%\" >nul
copy /Y "%TEMP_DIR%\grove\grove-loop.exe" "%INSTALL_DIR%\" >nul
copy /Y "%TEMP_DIR%\grove\grove-opti.exe" "%INSTALL_DIR%\" >nul
copy /Y "%TEMP_DIR%\grove\install.bat" "%INSTALL_DIR%\" >nul

REM Add to PATH
echo %PATH% | findstr /C:"%INSTALL_DIR%" >nul
if %ERRORLEVEL% neq 0 (
    setx PATH "%PATH%;%INSTALL_DIR%" >nul 2>&1
)

echo [4/4] Instalando skills...
set SKILLS_DIR=%USERPROFILE%\.config\opencode\skills
if not exist "%SKILLS_DIR%" mkdir "%SKILLS_DIR%"
if exist "%TEMP_DIR%\grove\skills" (
    xcopy /E /I /Y "%TEMP_DIR%\grove\skills\*" "%SKILLS_DIR%\" >nul
)

REM Cleanup
rmdir /S /Q "%TEMP_DIR%" >nul 2>&1

echo.
echo  ╔═══════════════════════════════════════════════════════╗
echo  ║           ✅ GROVE INSTALADO                         ║
echo  ╚═══════════════════════════════════════════════════════╝
echo.
echo  Cierra esta terminal y abre una nueva.
echo.
echo  Luego prueba:
echo    grove-spec --help
echo.
pause
