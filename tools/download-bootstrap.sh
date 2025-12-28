#!/bin/bash

# Download Bootstrap CSS to static folder

BOOTSTRAP_VERSION="5.3.8"
BOOTSTRAP_CSS_URL="https://cdn.jsdelivr.net/npm/bootstrap@${BOOTSTRAP_VERSION}/dist/css/bootstrap.min.css"
STATIC_DIR="../static"

# Check that script is executed from tools folder
if [ ! -d "$STATIC_DIR" ]; then
    echo "✗ Error: Script must be executed from the tools folder"
    echo "  $STATIC_DIR directory not found"
    exit 1
fi

echo "Downloading Bootstrap ${BOOTSTRAP_VERSION} CSS..."

# Download Bootstrap CSS

if curl -L -o "${STATIC_DIR}/bootstrap.min.css" "$BOOTSTRAP_CSS_URL"; then
    echo "✓ Bootstrap CSS downloaded successfully to ${STATIC_DIR}/bootstrap.min.css"
else
    echo "✗ Failed to download Bootstrap CSS"
    exit 1
fi
