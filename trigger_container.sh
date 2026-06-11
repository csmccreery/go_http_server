#!/usr/bin/env bash

# Actually a good reason to do this, don't want to enter weird namespaces.
set -euo pipefail

# Create a guix container with all the goodies we need
echo "Setting variables"

USER_NAME=$1
CONTAINER_FILE=$2
GIT_PATH=$3
PID_FILE=$(mktemp /tmp/guix-container-XXXXXX.pid)


trap 'rm -rf "$PID_FILE"' EXIT

echo "Creating container"

RUN_PATH=$(sudo guix system container "$CONTAINER_FILE" --network \
     --expose=$HOME/.gitconfig=/home/$USER_NAME/.gitconfig \
     --expose=$HOME/.gitconfig-github=/home/$USER_NAME/.gitconfig-github \
     --expose=$HOME/.gitconfig-codeberg=/home/$USER_NAME/.gitconfig-codeberg \
     --expose=$HOME/Projects/Github/bootdev-projects/http_servers=/home/$USER_NAME/src/ \
     --expose=$HOME/.ssh=/home/$USER_NAME/.ssh | tail -n 1)

echo "Running container"

sudo "$RUN_PATH" --pid-file="$PID_FILE" > /dev/null 2>&1 &

# Gotta give it a minute, one day I'll just turn this into a Guile script
sleep 5

# At this point one might ask why I'm doing this vice just using podman.
CONTAINER_PID=$(cat "$PID_FILE")

# Need to put something here to change the owner of the "dev" user and make
# the necessary directories

# Especially since I'm _running_ rootless-podman on Guix System
echo "Container running at $CONTAINER_PID"

sudo nsenter -a -t "$CONTAINER_PID"

# Because I like the pain
