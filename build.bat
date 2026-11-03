@echo off
setlocal enabledelayedexpansion

rem ===================================================
rem  fnOS Terminal Packaging Script (Windows)
rem  Usage: build.bat [version]
rem  Example: build.bat 1.0.1
rem ===================================================

rem 1. Determine and update version
set "TARGET_VERSION=%~1"
for /f "delims=" %%v in ('go run scripts\set_version.go %TARGET_VERSION%') do (
    set "APP_VERSION=%%v"
)

if "%APP_VERSION%"=="" (
    set "APP_VERSION=1.0.1"
)

echo ===================================================
echo  Packaging fnOS Terminal (Version: %APP_VERSION%)
echo ===================================================

echo.
echo [1/4] Building Frontend UI with pnpm...
cd ui
call pnpm run build
if %errorlevel% neq 0 (
    echo [ERROR] Frontend build failed!
    cd ..
    exit /b %errorlevel%
)
cd ..

echo.
echo [2/4] Syncing UI Dist and Verifying Scripts...
if not exist "pty\dist" mkdir "pty\dist"
xcopy /E /I /Y "ui\dist" "pty\dist" > nul
if %errorlevel% neq 0 (
    echo [ERROR] Failed to copy ui dist to pty/dist!
    exit /b %errorlevel%
)
go run scripts\write_cmd_main.go
if %errorlevel% neq 0 (
    echo [ERROR] Failed to write package/cmd/main!
    exit /b %errorlevel%
)

echo.
echo [3/4] Cross-Compiling Go Backend (Linux x86_64)...
cd pty
if not exist "..\package\app\bin" mkdir "..\package\app\bin"
set "CGO_ENABLED=0"
set "GOOS=linux"
set "GOARCH=amd64"
go build -ldflags="-s -w" -o "..\package\app\bin\terminal" main.go
if %errorlevel% neq 0 (
    echo [ERROR] Go cross-compilation failed!
    cd ..
    exit /b %errorlevel%
)
cd ..

echo.
echo [4/4] Packaging with fnpack...
if exist "terminal.fpk" del /f /q "terminal.fpk"
fnpack.exe build -d package
if %errorlevel% neq 0 (
    echo [ERROR] fnpack build failed!
    exit /b %errorlevel%
)

set "OUTPUT_FPK=terminal-%APP_VERSION%.fpk"
if exist "%OUTPUT_FPK%" del /f /q "%OUTPUT_FPK%"
copy /Y "terminal.fpk" "%OUTPUT_FPK%" > nul

echo.
echo ===================================================
echo [SUCCESS] Build Complete!
echo Output file: %OUTPUT_FPK%
echo Legacy alias: terminal.fpk
echo ===================================================
dir %OUTPUT_FPK%
