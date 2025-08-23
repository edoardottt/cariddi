@echo off
REM Build script for cariddi Docker image (Windows batch version)
REM This script builds the Docker image with proper metadata

setlocal enabledelayedexpansion

REM Default values
set IMAGE_NAME=edoardottt/cariddi
set TAG=latest
set PUSH=false
set MULTI_ARCH=false

REM Get build information
for /f "usebackq delims=" %%i in (`powershell -command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ' -AsUTC"`) do set BUILD_DATE=%%i
for /f "usebackq delims=" %%i in (`git rev-parse --short HEAD 2^>nul ^|^| echo unknown`) do set VCS_REF=%%i
for /f "usebackq delims=" %%i in (`git describe --tags --abbrev^=0 2^>nul ^|^| echo dev`) do set VERSION=%%i

REM Parse command line arguments
:parse
if "%~1"=="" goto build
if "%~1"=="-t" (
    set TAG=%~2
    shift
    shift
    goto parse
)
if "%~1"=="--tag" (
    set TAG=%~2
    shift
    shift
    goto parse
)
if "%~1"=="-p" (
    set PUSH=true
    shift
    goto parse
)
if "%~1"=="--push" (
    set PUSH=true
    shift
    goto parse
)
if "%~1"=="-m" (
    set MULTI_ARCH=true
    shift
    goto parse
)
if "%~1"=="--multi-arch" (
    set MULTI_ARCH=true
    shift
    goto parse
)
if "%~1"=="-v" (
    set VERSION=%~2
    shift
    shift
    goto parse
)
if "%~1"=="--version" (
    set VERSION=%~2
    shift
    shift
    goto parse
)
if "%~1"=="-h" goto usage
if "%~1"=="--help" goto usage

echo Unknown option: %~1
goto usage

:usage
echo Usage: %~nx0 [OPTIONS]
echo.
echo Options:
echo   -t, --tag TAG        Image tag (default: latest)
echo   -p, --push           Push image to registry after build
echo   -m, --multi-arch     Build multi-architecture image
echo   -v, --version VER    Set version (default: git tag or 'dev')
echo   -h, --help           Show this help message
echo.
echo Examples:
echo   %~nx0                   # Build local image
echo   %~nx0 -t v1.0.0 -p      # Build and push with tag v1.0.0
echo   %~nx0 -m -p             # Build multi-arch and push
exit /b 0

:build
echo Building cariddi Docker image...
echo Image: %IMAGE_NAME%:%TAG%
echo Version: %VERSION%
echo Build Date: %BUILD_DATE%
echo VCS Ref: %VCS_REF%

REM Build the Docker image
set BUILD_ARGS=--build-arg VERSION=%VERSION% --build-arg BUILD_DATE=%BUILD_DATE% --build-arg VCS_REF=%VCS_REF% --tag %IMAGE_NAME%:%TAG% --tag %IMAGE_NAME%:latest

if "%MULTI_ARCH%"=="true" (
    echo Building multi-architecture image...
    
    REM Create and use buildx builder if it doesn't exist
    docker buildx ls | findstr cariddi-builder >nul || (
        docker buildx create --name cariddi-builder --driver docker-container --bootstrap
    )
    docker buildx use cariddi-builder
    
    if "%PUSH%"=="true" (
        docker buildx build %BUILD_ARGS% --platform linux/amd64,linux/arm64 --push .
    ) else (
        docker buildx build %BUILD_ARGS% --platform linux/amd64,linux/arm64 --load .
    )
) else (
    docker build %BUILD_ARGS% .
    
    if "%PUSH%"=="true" (
        echo Pushing image to registry...
        docker push %IMAGE_NAME%:%TAG%
        docker push %IMAGE_NAME%:latest
    )
)

if errorlevel 1 (
    echo Build failed!
    exit /b 1
) else (
    echo Build completed successfully!
    echo Docker image built: %IMAGE_NAME%:%TAG%
    echo.
    echo Usage examples:
    echo   docker run --rm %IMAGE_NAME%:%TAG% --help
    echo   docker run --rm -v %CD%:/data %IMAGE_NAME%:%TAG% -intensive ^< /data/urls.txt
    echo   docker-compose up cariddi
)

endlocal
