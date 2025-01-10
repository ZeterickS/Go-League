#!/bin/bash

# Get the directory of the current script
SCRIPT_DIR=$(dirname "$(readlink -f "$0")")

# Load environment variables from .env file
source ".env"

# Authenticate with GitHub Container Registry
echo $GITHUB_TOKEN | docker login ghcr.io -u $GITHUB_USERNAME --password-stdin

# Determine the Docker image tag based on the DEVELOPMENT variable
if [ "$DEVELOPMENT" = "True" ]; then
    export IMAGE_TAG="test"
else
    export IMAGE_TAG="main"
fi

# Check if there is a new version of the Docker image
echo "Checking for new version of ghcr.io/zetericks/go-league:$IMAGE_TAG"
docker pull ghcr.io/zetericks/go-league:$IMAGE_TAG
LOCAL_IMAGE_ID=$(docker images -q ghcr.io/zetericks/go-league:$IMAGE_TAG)
REMOTE_IMAGE_ID=$(docker inspect --format='{{.Id}}' ghcr.io/zetericks/go-league:$IMAGE_TAG)

if [ "$LOCAL_IMAGE_ID" != "$REMOTE_IMAGE_ID" ]; then
    echo "New version detected, running docker compose up"
    docker compose --env-file .env up -d --build
else
    echo "No new version detected, skipping docker compose up"
fi