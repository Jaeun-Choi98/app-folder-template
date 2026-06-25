#!/bin/bash
set -e

IMAGE_NAME="appfolder-cpp"
TAG="latest"

docker build -t "${IMAGE_NAME}:${TAG}" .
echo "Built ${IMAGE_NAME}:${TAG}"
