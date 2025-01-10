#!/bin/bash

# This script checks for repository updates and triggers the ci.sh script.
# It should be run periodically using cron to ensure continuous integration.

# Define the repository paths
REPO_PATHS=("/opt/Go-League" "/opt/Go-League-Test")

# Function to fetch and run ci.sh if there are new changes
fetch_and_run() {
  local repo_path=$1
  cd $repo_path

  # Get the current branch name
  local branch=$(git symbolic-ref --short HEAD)
  echo "Checking for new changes in $repo_path on branch $branch..."

  # Attempt to pull changes
  if ! git pull origin "$branch"; then
    echo "git pull failed, performing git reset --hard"
    git reset --hard origin/"$branch"
  else
    echo "git pull succeeded"
  fi

  # Run the ci.sh script, that checks for new docker image
  /bin/bash ./ci.sh
}

# Iterate over each repository path and perform fetch and run
for repo_path in "${REPO_PATHS[@]}"; do
  fetch_and_run $repo_path
done