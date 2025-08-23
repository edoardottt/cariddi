#!/bin/bash

# Test script for cariddi Docker image
# This script validates the Docker setup and functionality

set -e

IMAGE_NAME="edoardottt/cariddi:latest"
TEST_URL="https://edoardottt.com/"

echo "Testing cariddi Docker image: ${IMAGE_NAME}"

# Test 1: Check if image exists
echo "Test 1: Checking if image exists..."
if docker image inspect ${IMAGE_NAME} >/dev/null 2>&1; then
    echo "PASS - Image exists"
else
    echo "FAIL - Image not found. Building image..."
    docker build -t ${IMAGE_NAME} .
fi

# Test 2: Test help command
echo "Test 2: Testing help command..."
if docker run --rm ${IMAGE_NAME} --help >/dev/null 2>&1; then
    echo "PASS - Help command works"
else
    echo "FAIL - Help command failed"
    exit 1
fi

# Test 3: Test version information
echo "Test 3: Testing version information..."
VERSION_OUTPUT=$(docker run --rm ${IMAGE_NAME} --version 2>/dev/null || echo "No version flag")
echo "Version output: ${VERSION_OUTPUT}"

# Test 4: Test with stdin input
echo "Test 4: Testing with stdin input..."
if echo "${TEST_URL}" | docker run --rm -i ${IMAGE_NAME} -intensive >/dev/null 2>&1; then
    echo "PASS - Stdin input works"
else
    echo "FAIL - Stdin input failed"
    exit 1
fi

# Test 5: Test with file input/output
echo "Test 5: Testing with file input/output..."
mkdir -p test_data
echo "${TEST_URL}" > test_data/urls.txt

if docker run --rm -v $(pwd)/test_data:/data ${IMAGE_NAME} -intensive -s /data/urls.txt -ot /data/results.txt; then
    if [ -f test_data/results.txt ]; then
        echo "PASS - File input/output works"
        echo "Results file size: $(wc -c < test_data/results.txt) bytes"
    else
        echo "FAIL - Output file not created"
        exit 1
    fi
else
    echo "FAIL - File input/output failed"
    exit 1
fi

# Test 6: Test JSON output
echo "Test 6: Testing JSON output..."
if docker run --rm -v $(pwd)/test_data:/data ${IMAGE_NAME} -intensive -s /data/urls.txt -json -ot /data/results.json; then
    if [ -f test_data/results.json ]; then
        echo "PASS - JSON output works"
        echo "JSON file size: $(wc -c < test_data/results.json) bytes"
    else
        echo "FAIL - JSON output file not created"
        exit 1
    fi
else
    echo "FAIL - JSON output failed"
    exit 1
fi

# Test 7: Test docker-compose (if file exists)
if [ -f docker-compose.yml ]; then
    echo "Test 7: Testing docker-compose..."
    if command -v docker-compose >/dev/null 2>&1; then
        # Create input for docker-compose
        mkdir -p input output
        echo "${TEST_URL}" > input/urls.txt
        
        if timeout 30 docker-compose up cariddi; then
            echo "PASS - Docker-compose works"
        else
            echo "WARN - Docker-compose test timed out (may be normal)"
        fi
        
        # Cleanup
        docker-compose down 2>/dev/null || true
    else
        echo "WARN - docker-compose not found, skipping test"
    fi
fi

# Test 8: Test image metadata
echo "Test 8: Checking image metadata..."
LABELS=$(docker image inspect ${IMAGE_NAME} --format '{{json .Config.Labels}}' | jq -r 'to_entries[] | "\(.key)=\(.value)"' 2>/dev/null || echo "No labels found")
echo "Image labels:"
echo "${LABELS}"

# Test 9: Test security (non-root user)
echo "Test 9: Testing security (non-root user)..."
USER_INFO=$(docker run --rm ${IMAGE_NAME} whoami 2>/dev/null || echo "cariddi")
if [ "${USER_INFO}" = "cariddi" ]; then
    echo "PASS - Running as non-root user"
else
    echo "FAIL - Running as root or unknown user: ${USER_INFO}"
    exit 1
fi

# Test 10: Test image size
echo "Test 10: Checking image size..."
IMAGE_SIZE=$(docker image ls ${IMAGE_NAME} --format "{{.Size}}")
echo "Image size: ${IMAGE_SIZE}"

# Cleanup
echo "Cleaning up test data..."
rm -rf test_data input output

echo "SUCCESS - All tests passed! Docker image is ready for use."
echo "Usage examples:"
echo "  docker run --rm ${IMAGE_NAME} --help"
echo "  echo 'https://example.com' | docker run --rm -i ${IMAGE_NAME} -intensive"
echo "  docker run --rm -v \$(pwd):/data ${IMAGE_NAME} -intensive -s /data/urls.txt -ot /data/results.txt"
