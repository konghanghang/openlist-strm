#!/bin/sh
set -e

echo "-----------------------------------"
echo "  OpenList-STRM Docker Container  "
echo "-----------------------------------"
echo "Running as UID:$(id -u) GID:$(id -g)"
echo "-----------------------------------"

# Create necessary directories if they don't exist
mkdir -p /app/data /app/logs /app/configs

# Execute the main command
exec "$@"
