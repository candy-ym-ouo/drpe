#!/bin/bash
set -e
IMAGE_NAME=${1:-drpe}
DOCKER_PLATFORM=${2:-linux/amd64}
docker build --platform "$DOCKER_PLATFORM" -f benzhi.Dockerfile -t "$IMAGE_NAME" .
echo "Docker image '$IMAGE_NAME:latest' built successfully!"
echo "Run: docker run -it $IMAGE_NAME:latest"
