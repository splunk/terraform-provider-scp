#!/bin/bash

DEVELOP_BRANCH='develop'
STAGING_BRANCH='staging'

USER_EMAIL='ssg-acs-eng@splunk.com'
USER_NAME='Admin Config Service'

# Helper function does branch checkout and pulls the latest changes
checkout_branch() {
    if [ -z "$1" ]; then
        echo "No branch provided"
        exit 1
    fi

    echo "Checking out branch: $1"

    git fetch origin "$1"
    git checkout "$1"
    git pull
}

# Configure git
echo "Configuring Git"
git config --global user.email "$USER_EMAIL"
git config --global user.name "$USER_NAME"

git remote set-url origin https://${GITLAB_TOKEN_USERNAME}:${GITLAB_TOKEN}@cd.splunkdev.com/${CI_PROJECT_PATH}.git

# Check available branches
echo "Available branch"
git branch -vaa

# Get the latest changes from develop branch
checkout_branch "$DEVELOP_BRANCH"

# Get the latest changes from staging branch
checkout_branch "$STAGING_BRANCH"

# Merge changes from develop to staging branch
echo "Merging changes from $STAGING_BRANCH to $DEVELOP_BRANCH ..."
git merge $DEVELOP_BRANCH

# Push changes to remote staging branch
echo "Pushing merges to remote branch..."
git push