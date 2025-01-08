#!/bin/bash
cd /opt/Go-League-Test || exit 1
# Pull the latest changes
git pull

# Check if there are any changes
if [ "$(git rev-parse HEAD)" != "$(git rev-parse @{u})" ]; then
    echo "Changes detected, running docker compose up"
    docker compose -f docker-compose-test.yml up -d --build
else
    echo "No changes detected, skipping docker compose up"
fi

# Check if there are any changes in the working directory
if [ -n "$(git status --porcelain)" ]; then
    echo "Changes detected in the working directory, running docker compose up"
    docker compose -f docker-compose-test.yml up -d --build
else
    echo "No changes detected in the working directory, skipping docker compose up"
fi