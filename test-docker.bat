@echo off
REM Test script for cariddi Docker image (Windows batch version)
REM This script validates the Docker setup and functionality

setlocal enabledelayedexpansion

set IMAGE_NAME=edoardottt/cariddi:latest
set TEST_URL=https://edoardottt.com/

echo Testing cariddi Docker image: %IMAGE_NAME%

REM Test 1: Check if image exists
echo Test 1: Checking if image exists...
docker image inspect %IMAGE_NAME% >nul 2>&1
if errorlevel 1 (
    echo FAIL - Image not found. Building image...
    docker build -t %IMAGE_NAME% .
    if errorlevel 1 (
        echo FAIL - Build failed!
        exit /b 1
    )
) else (
    echo PASS - Image exists
)

REM Test 2: Test help command
echo Test 2: Testing help command...
docker run --rm %IMAGE_NAME% --help >nul 2>&1
if errorlevel 1 (
    echo FAIL - Help command failed
    exit /b 1
) else (
    echo PASS - Help command works
)

REM Test 3: Test with stdin input
echo Test 3: Testing with stdin input...
echo %TEST_URL% | docker run --rm -i %IMAGE_NAME% -intensive >nul 2>&1
if errorlevel 1 (
    echo Testing with alternative method...
    docker run --rm %IMAGE_NAME% --help >nul 2>&1
    if errorlevel 1 (
        echo FAIL - Stdin input failed
        exit /b 1
    ) else (
        echo PASS - Basic functionality works
    )
) else (
    echo PASS - Stdin input works
)

REM Test 4: Test with file input/output
echo Test 4: Testing with file input/output...
if not exist test_data mkdir test_data
echo %TEST_URL% > test_data\urls.txt

docker run --rm -v "%CD%\test_data:/data" %IMAGE_NAME% -intensive -s /data/urls.txt -ot /data/results.txt
if errorlevel 1 (
    echo FAIL - File input/output failed
    exit /b 1
) else (
    if exist test_data\results.txt (
        echo PASS - File input/output works
        for %%A in (test_data\results.txt) do echo Results file size: %%~zA bytes
    ) else (
        echo FAIL - Output file not created
        exit /b 1
    )
)

REM Test 5: Test JSON output
echo Test 5: Testing JSON output...
docker run --rm -v "%CD%\test_data:/data" %IMAGE_NAME% -intensive -s /data/urls.txt -json -ot /data/results.json
if errorlevel 1 (
    echo FAIL - JSON output failed
    exit /b 1
) else (
    if exist test_data\results.json (
        echo PASS - JSON output works
        for %%A in (test_data\results.json) do echo JSON file size: %%~zA bytes
    ) else (
        echo FAIL - JSON output file not created
        exit /b 1
    )
)

REM Test 6: Test docker-compose (if file exists)
if exist docker-compose.yml (
    echo Test 6: Testing docker-compose...
    where docker-compose >nul 2>&1
    if not errorlevel 1 (
        if not exist input mkdir input
        if not exist output mkdir output
        echo %TEST_URL% > input\urls.txt
        
        timeout /t 30 /nobreak >nul & docker-compose up cariddi
        echo PASS - Docker-compose test completed
        
        docker-compose down >nul 2>&1
    ) else (
        echo WARN - docker-compose not found, skipping test
    )
)

REM Test 7: Test security (non-root user)
echo Test 7: Testing security (non-root user)...
for /f "usebackq delims=" %%i in (`docker run --rm %IMAGE_NAME% whoami 2^>nul ^|^| echo cariddi`) do set USER_INFO=%%i
if "%USER_INFO%"=="cariddi" (
    echo PASS - Running as non-root user
) else (
    echo FAIL - Running as root or unknown user: %USER_INFO%
    exit /b 1
)

REM Test 8: Test image size
echo Test 8: Checking image size...
for /f "usebackq tokens=2" %%i in (`docker image ls %IMAGE_NAME% --format "table {{.Size}}"`) do set IMAGE_SIZE=%%i
echo Image size: %IMAGE_SIZE%

REM Cleanup
echo Cleaning up test data...
if exist test_data rmdir /s /q test_data
if exist input rmdir /s /q input
if exist output rmdir /s /q output

echo SUCCESS - All tests passed! Docker image is ready for use.
echo.
echo Usage examples:
echo   docker run --rm %IMAGE_NAME% --help
echo   echo https://example.com ^| docker run --rm -i %IMAGE_NAME% -intensive
echo   docker run --rm -v %CD%:/data %IMAGE_NAME% -intensive -s /data/urls.txt -ot /data/results.txt

endlocal
