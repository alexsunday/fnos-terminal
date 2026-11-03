@echo off
setlocal enabledelayedexpansion

rem ===================================================
rem  fnOS Terminal Packaging Script (Windows)
rem  Usage: build.bat [version]
rem  Example: build.bat 1.0.1
rem ===================================================

rem Ensure clean host Go environment
set "GOOS="
set "GOARCH="

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
if !errorlevel! neq 0 (
    echo [ERROR] Frontend build failed!
    cd ..
    exit /b !errorlevel!
)
cd ..

echo.
echo [2/4] Syncing UI Dist and Verifying Scripts...
if not exist "pty\dist" mkdir "pty\dist"
xcopy /E /I /Y "ui\dist" "pty\dist" > nul
if !errorlevel! neq 0 (
    echo [ERROR] Failed to copy ui dist to pty/dist!
    exit /b !errorlevel!
)
go run scripts\write_cmd_main.go
if !errorlevel! neq 0 (
    echo [ERROR] Failed to write package/cmd/main!
    exit /b !errorlevel!
)

if not exist "output" mkdir "output"
if not exist "package\app\bin" mkdir "package\app\bin"

echo.
echo [3/4] Building and Packaging AMD64 [x86]...
set "GOOS="
set "GOARCH="
go run scripts\set_platform.go x86
if !errorlevel! neq 0 (
    echo [ERROR] Failed to set platform to x86!
    exit /b !errorlevel!
)

cd pty
set "CGO_ENABLED=0"
set "GOOS=linux"
set "GOARCH=amd64"
go build -ldflags="-s -w" -o "..\package\app\bin\terminal" main.go
set "GOOS="
set "GOARCH="
if !errorlevel! neq 0 (
    echo [ERROR] Go cross-compilation amd64 failed!
    cd ..
    exit /b !errorlevel!
)
cd ..

if exist "terminal.fpk" del /f /q "terminal.fpk"
fnpack.exe build -d package
if !errorlevel! neq 0 (
    echo [ERROR] fnpack build amd64 failed!
    exit /b !errorlevel!
)

set "OUTPUT_AMD64=output\terminal-%APP_VERSION%-amd64.fpk"
if exist "%OUTPUT_AMD64%" del /f /q "%OUTPUT_AMD64%"
move /Y "terminal.fpk" "%OUTPUT_AMD64%" > nul

echo.
echo [4/4] Building and Packaging ARM64 [arm]...
set "GOOS="
set "GOARCH="
go run scripts\set_platform.go arm
if !errorlevel! neq 0 (
    echo [ERROR] Failed to set platform to arm!
    exit /b !errorlevel!
)

cd pty
set "CGO_ENABLED=0"
set "GOOS=linux"
set "GOARCH=arm64"
go build -ldflags="-s -w" -o "..\package\app\bin\terminal" main.go
set "GOOS="
set "GOARCH="
if !errorlevel! neq 0 (
    echo [ERROR] Go cross-compilation arm64 failed!
    cd ..
    exit /b !errorlevel!
)
cd ..

if exist "terminal.fpk" del /f /q "terminal.fpk"
fnpack.exe build -d package
if !errorlevel! neq 0 (
    echo [ERROR] fnpack build arm64 failed!
    exit /b !errorlevel!
)

set "OUTPUT_ARM64=output\terminal-%APP_VERSION%-arm64.fpk"
if exist "%OUTPUT_ARM64%" del /f /q "%OUTPUT_ARM64%"
move /Y "terminal.fpk" "%OUTPUT_ARM64%" > nul

rem Reset platform to default x86 in manifest
set "GOOS="
set "GOARCH="
go run scripts\set_platform.go x86 > nul 2>&1

rem Clean up any root terminal.fpk if left behind
if exist "terminal.fpk" del /f /q "terminal.fpk"

echo.
echo ===================================================
echo [SUCCESS] Multi-Arch Packaging Complete!
echo Output packages generated:
echo   - %OUTPUT_AMD64%
echo   - %OUTPUT_ARM64%
echo ===================================================
dir output\terminal-%APP_VERSION%-*.fpk
