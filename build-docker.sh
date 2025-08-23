#!/bin/bash

# Build script for cariddi Docker image
# This script builds the Docker image with proper metadata

set -e

# Default values
IMAGE_NAME="edoardottt/cariddi"
TAG="latest"
PUSH=false
PLATFORMS="linux/amd64,linux/arm64"

# Get build information
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")

usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -t, --tag TAG        Image tag (default: latest)"
    echo "  -p, --push           Push image to registry after build"
    echo "  -m, --multi-arch     Build multi-architecture image"
    echo "  -v, --version VER    Set version (default: git tag or 'dev')"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                   # Build local image"
    echo "  $0 -t v1.0.0 -p      # Build and push with tag v1.0.0"
    echo "  $0 -m -p             # Build multi-arch and push"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -p|--push)
            PUSH=true
            shift
            ;;
        -m|--multi-arch)
            MULTI_ARCH=true
            shift
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

echo "Building cariddi Docker image..."
echo "Image: ${IMAGE_NAME}:${TAG}"
echo "Version: ${VERSION}"
echo "Build Date: ${BUILD_DATE}"
echo "VCS Ref: ${VCS_REF}"

# Build arguments
BUILD_ARGS=(
    --build-arg "VERSION=${VERSION}"
    --build-arg "BUILD_DATE=${BUILD_DATE}"
    --build-arg "VCS_REF=${VCS_REF}"
    --tag "${IMAGE_NAME}:${TAG}"
    --tag "${IMAGE_NAME}:latest"
)

if [ "$MULTI_ARCH" = true ]; then
    echo -e "${BLUE}Building multi-architecture image...${NC}"
    
    # Create and use buildx builder if it doesn't exist
    if ! docker buildx ls | grep -q cariddi-builder; then
        docker buildx create --name cariddi-builder --driver docker-container --bootstrap
    fi
    docker buildx use cariddi-builder
    
    BUILD_COMMAND="docker buildx build"
    BUILD_ARGS+=(--platform "${PLATFORMS}")
    
    if [ "$PUSH" = true ]; then
        BUILD_ARGS+=(--push)
    else
        BUILD_ARGS+=(--load)
    fi
else
    BUILD_COMMAND="docker build"
fi

# Execute build
echo -e "${BLUE}Executing: ${BUILD_COMMAND} ${BUILD_ARGS[*]} .${NC}"
${BUILD_COMMAND} "${BUILD_ARGS[@]}" .

if [ $? -eq 0 ]; then
    echo "Build completed successfully!"
    
    if [ "$PUSH" = true ] && [ "$MULTI_ARCH" != true ]; then
        echo "Pushing image to registry..."
        docker push "${IMAGE_NAME}:${TAG}"
        docker push "${IMAGE_NAME}:latest"
        echo "Push completed successfully!"
    fi
    
    echo "Docker image built: ${IMAGE_NAME}:${TAG}"
    echo -e "${YELLOW}Usage examples:${NC}"
    echo -e "  docker run --rm ${IMAGE_NAME}:${TAG} --help"
    echo -e "  docker run --rm -v \$(pwd):/data ${IMAGE_NAME}:${TAG} -intensive < /data/urls.txt"
    echo -e "  docker-compose up cariddi"
else
    echo "Build failed!"
    exit 1
fi
