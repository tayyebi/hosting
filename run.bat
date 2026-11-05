@echo off
setlocal
cd /d "%~dp0"

echo [Platform] Stopping any running instance...
taskkill /F /IM platform.exe >nul 2>&1

echo [Platform] Cleaning old binary...
if exist platform.exe del /f /q platform.exe

echo [Platform] Building...
go build -o platform.exe .
if %errorlevel% neq 0 (
    echo [Platform] Build failed.
    exit /b 1
)

if exist .env (
    echo [Platform] Loading .env...
    for /f "usebackq tokens=* delims=" %%A in (".env") do (
        set "_line=%%A"
        call :loadenv
    )
)

echo [Platform] Starting on :3000
.\platform.exe
goto :eof

:loadenv
if "%_line:~0,1%"=="#" goto :eof
if "%_line%"=="" goto :eof
set "%_line%"
goto :eof
