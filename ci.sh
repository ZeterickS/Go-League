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
    export TEST="true"
else
    export IMAGE_TAG="main"
fi

LOCAL_IMAGE_ID=$(docker images -q ghcr.io/zetericks/go-league:$IMAGE_TAG)

# Check if there is a new version of the Docker image
echo "Checking for new version of ghcr.io/zetericks/go-league:$IMAGE_TAG"
docker pull ghcr.io/zetericks/go-league:$IMAGE_TAG

REMOTE_IMAGE_ID=$(docker inspect --format='{{.Id}}' ghcr.io/zetericks/go-league:$IMAGE_TAG)
REMOTE_IMAGE_ID_STRIPPED=${REMOTE_IMAGE_ID#sha256:}
echo "Local image ID: $LOCAL_IMAGE_ID"
echo "Remote image ID: $REMOTE_IMAGE_ID"
echo "Remote image ID stripped: $REMOTE_IMAGE_ID_STRIPPED"

if [ "$LOCAL_IMAGE_ID" != "${REMOTE_IMAGE_ID_STRIPPED:0:12}" ]; then
    echo "New version detected, running docker compose up"
    docker compose --env-file .env up -d --build
else
    echo "No new version detected, skipping docker compose up"
fi

# Check if the Docker Compose service is running
SERVICE_STATUS=$(docker inspect -f '{{.State.Status}}' GoL-Tracker${TEST:+-test})

if [ "$SERVICE_STATUS" != "running" ]; then
    echo "Service is not running, restarting docker compose"
    docker compose --env-file .env up -d --build
else
    echo "Service is running"
fi