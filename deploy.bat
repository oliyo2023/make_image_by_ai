@echo off
:: AI Image Generator Docker Deploy Script (Windows)
:: Usage: deploy.bat [command]

setlocal EnableDelayedExpansion

:: Project config
set PROJECT_NAME=ai-image-generator
set IMAGE_NAME=ai-image-generator
set CONTAINER_NAME=ai-image-generator
set PORT=8000

:: Check parameters
set COMMAND=%1
if "%COMMAND%"=="" set COMMAND=help

:: Main logic
goto %COMMAND% 2>nul || goto help

:setup
echo [INFO] Setting up Docker environment...
call :check_docker
if errorlevel 1 exit /b 1

:: Create environment file
if not exist .env (
    copy .env.example .env >nul
    echo [SUCCESS] Created .env file, please edit it to set your API keys
    echo [WARNING] Please set the following required environment variables in .env:
    echo   - MODEL_SCOPE_TOKEN
    echo   - OPENROUTER_API_KEY
) else (
    echo [INFO] .env file already exists
)

:: Create necessary directories
if not exist public\static\images mkdir public\static\images
if not exist logs mkdir logs
echo [SUCCESS] Created necessary directories

echo [SUCCESS] Environment setup completed
goto :eof

:build
echo [INFO] Building Docker image...
call :check_docker
if errorlevel 1 exit /b 1

docker build -t %IMAGE_NAME%:latest .
if errorlevel 1 (
    echo [ERROR] Failed to build image
    exit /b 1
)

echo [SUCCESS] Image build completed
goto :eof

:run
echo [INFO] Starting container...
call :check_docker
if errorlevel 1 exit /b 1

:: Stop existing container
call :stop_container

:: Start new container
docker run -d ^
    --name %CONTAINER_NAME% ^
    -p %PORT%:%PORT% ^
    -v "%cd%\public\static\images:/app/public/static/images" ^
    -v "%cd%\logs:/app/logs" ^
    --env-file .env ^
    --restart unless-stopped ^
    %IMAGE_NAME%:latest

if errorlevel 1 (
    echo [ERROR] Failed to start container
    exit /b 1
)

echo [SUCCESS] Container started successfully
echo [INFO] Service URL: http://localhost:%PORT%
echo [INFO] Health check: http://localhost:%PORT%/health

:: Wait for service to start
echo [INFO] Waiting for service to start...
timeout /t 5 /nobreak >nul

:: Check container status
docker ps | findstr %CONTAINER_NAME% >nul
if errorlevel 1 (
    echo [ERROR] Container failed to start, check logs:
    docker logs %CONTAINER_NAME%
    exit /b 1
) else (
    echo [SUCCESS] Container is running normally
)
goto :eof

:stop
echo [INFO] Stopping container...
call :check_docker
if errorlevel 1 exit /b 1

call :stop_container
goto :eof

:logs
echo [INFO] Showing container logs (Press Ctrl+C to exit)...
call :check_docker
if errorlevel 1 exit /b 1

docker ps | findstr %CONTAINER_NAME% >nul
if errorlevel 1 (
    echo [ERROR] Container is not running
    exit /b 1
)

docker logs -f %CONTAINER_NAME%
goto :eof

:status
echo [INFO] Docker Status:
call :check_docker
if errorlevel 1 exit /b 1

echo === Images ===
docker images | findstr %IMAGE_NAME% || echo No related images

echo === Containers ===
docker ps -a | findstr %CONTAINER_NAME% || echo No related containers

echo === Running Status ===
docker ps | findstr %CONTAINER_NAME% >nul
if errorlevel 1 (
    echo [WARNING] Container is not running
) else (
    echo [SUCCESS] Container is running
    echo Service URL: http://localhost:%PORT%
)
goto :eof

:clean
echo [INFO] Cleaning Docker resources...
call :check_docker
if errorlevel 1 exit /b 1

:: Stop container
call :stop_container

:: Remove image
docker images | findstr %IMAGE_NAME% >nul
if not errorlevel 1 (
    docker rmi %IMAGE_NAME%:latest
    echo [SUCCESS] Image removed
)

:: Clean system
docker system prune -f
echo [SUCCESS] Docker resource cleanup completed
goto :eof

:help
echo AI Image Generator Docker Deploy Script (Windows)
echo.
echo Usage:
echo   %~nx0 [command]
echo.
echo Available commands:
echo   setup   - Setup environment and config files
echo   build   - Build Docker image
echo   run     - Run container
echo   stop    - Stop container
echo   logs    - Show container logs
echo   status  - Show status
echo   clean   - Clean all resources
echo   help    - Show this help message
echo.
echo Quick start:
echo   1. %~nx0 setup    # Setup environment
echo   2. Edit .env file, set API keys
echo   3. %~nx0 build    # Build image
echo   4. %~nx0 run      # Run service
goto :eof

:: Helper functions
:check_docker
docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not installed, please install Docker first
    exit /b 1
)

docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker service is not running, please start Docker service
    exit /b 1
)

echo [SUCCESS] Docker check passed
goto :eof

:stop_container
docker ps -a | findstr %CONTAINER_NAME% >nul
if not errorlevel 1 (
    echo [INFO] Stopping existing container...
    docker stop %CONTAINER_NAME% >nul 2>&1
    docker rm %CONTAINER_NAME% >nul 2>&1
    echo [SUCCESS] Container stopped
)
goto :eof