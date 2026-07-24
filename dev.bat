@echo off
rem DevKit one-click dev startup (Windows)
rem Loads .env, then starts server / web / www in separate windows.

setlocal
cd /d "%~dp0"

if not exist ".env" (
    echo [devkit] .env not found, copy .env.example to .env first.
    pause
    exit /b 1
)

rem Load environment variables from .env (skip comments and empty lines)
for /f "usebackq eol=# tokens=1,* delims==" %%a in (".env") do (
    if not "%%a"=="" set "%%a=%%b"
)

start "devkit-server" cmd /k "cd /d "%~dp0server" && go run ./cmd/server"
start "devkit-web"    cmd /k "cd /d "%~dp0web" && npm run dev"
start "devkit-www"    cmd /k "cd /d "%~dp0www" && npm run dev"

echo [devkit] Started server, web and www in separate windows.
echo [devkit] Close those windows to stop the services.
endlocal
