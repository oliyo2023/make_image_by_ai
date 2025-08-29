@echo off
:: Simple Docker deployment script for AI Image Generator

set IMAGE_NAME=ai-image-generator
set CONTAINER_NAME=ai-image-generator
set PORT=8000

if "%1"=="" goto help
if "%1"=="setup" goto setup
if "%1"=="build" goto build
if "%1"=="run" goto run
if "%1"=="stop" goto stop
if "%1"=="logs" goto logs
if "%1"=="status" goto status
if "%1"=="clean" goto clean
goto help

:setup
echo Setting up environment...
if not exist .env (
    copy .env.example .env
    echo Created .env file - please edit it to set your API keys
) else (
    echo .env file already exists
)
if not exist public\static\images mkdir public\static\images
if not exist logs mkdir logs
echo Setup completed!
goto end

:build
echo Building Docker image...
docker build -t %IMAGE_NAME%:latest .
if errorlevel 1 (
    echo Build failed!
    goto end
)
echo Build completed!
goto end

:run
echo Starting container...
call :stop
docker run -d --name %CONTAINER_NAME% -p %PORT%:%PORT% -v "%cd%\public\static\images:/app/public/static/images" -v "%cd%\logs:/app/logs" --env-file .env --restart unless-stopped %IMAGE_NAME%:latest
if errorlevel 1 (
    echo Failed to start container!
    goto end
)
echo Container started successfully!
echo Service URL: http://localhost:%PORT%
echo Health check: http://localhost:%PORT%/health
goto end

:stop
echo Stopping container...
docker stop %CONTAINER_NAME% 2>nul
docker rm %CONTAINER_NAME% 2>nul
echo Container stopped
goto end

:logs
echo Showing container logs...
docker logs -f %CONTAINER_NAME%
goto end

:status
echo Docker Status:
echo.
echo === Images ===
docker images | findstr %IMAGE_NAME%
echo.
echo === Containers ===
docker ps -a | findstr %CONTAINER_NAME%
echo.
echo === Running Status ===
docker ps | findstr %CONTAINER_NAME% >nul
if errorlevel 1 (
    echo Container is NOT running
) else (
    echo Container is running - Service URL: http://localhost:%PORT%
)
goto end

:clean
echo Cleaning up...
call :stop
docker rmi %IMAGE_NAME%:latest 2>nul
docker system prune -f
echo Cleanup completed!
goto end

:help
echo AI Image Generator Docker Deploy Script
echo.
echo Usage: %~nx0 [command]
echo.
echo Available commands:
echo   setup   - Setup environment and config files
echo   build   - Build Docker image
echo   run     - Run container
echo   stop    - Stop container
echo   logs    - Show container logs
echo   status  - Show status
echo   clean   - Clean all resources
echo   help    - Show this help
echo.
echo Quick start:
echo   1. %~nx0 setup
echo   2. Edit .env file to set API keys
echo   3. %~nx0 build
echo   4. %~nx0 run

:end