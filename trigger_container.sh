#!/usr/bin/env bash

# Actually a good reason to do this, don't want to enter weird namespaces.
set -euo pipefail

# Create a guix container with all the goodies we need
echo "Setting variables"

CONTAINER_FILE="./go_dev_container.scm"
PID_FILE=$(mktemp /tmp/guix-container-XXXXXX.pid)


trap 'rm -rf "$PID_FILE"' EXIT

echo "Creating container"

RUN_PATH=$(sudo guix system container "$CONTAINER_FILE" --network \
     --expose=$HOME/Projects/Github/bootdev-projects/http_servers=/root/src/ \
     --expose=$HOME/.ssh=/root/.ssh | tail -n 1)

echo "Running container"

sudo "$RUN_PATH" --pid-file="$PID_FILE" > /dev/null 2>&1 &

# Gotta give it a minute, one day I'll just turn this into a Guile script
sleep 5

# At this point one might ask why I'm doing this vice just using podman.
CONTAINER_PID=$(cat "$PID_FILE")

# Especially since I'm _running_ rootless-podman on Guix System
echo "Container running at $CONTAINER_PID"

sudo guix container exec $CONTAINER_PID /run/current-system/profile/bin/bash -c "export GOPATH=/run/current-system/profile/:/root/go GOMODCACHE=/run/curren-system/profile:/root/go/pkg/mod GOCACHE=/run/current-system/profile:/root/.cache/go-build && /run/current-system/profile/bin/go install github.com/pressly/goose/v3/cmd/goose@latest && /run/current-system/profile/bin/go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"

sudo guix container exec $CONTAINER_PID /run/current-system/profile/bin/bash -c "cd /root/src"

sudo nsenter -a -t "$CONTAINER_PID"

# Because I like the pain
