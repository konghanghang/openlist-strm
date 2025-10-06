#!/bin/sh
set -e

# Function to fix permissions for a directory
fix_permissions() {
    local dir=$1
    if [ -d "$dir" ]; then
        # Check if we can write to the directory
        if [ ! -w "$dir" ]; then
            echo "Warning: No write permission for $dir, trying to fix..."
            # Try to change ownership if running as root
            if [ "$(id -u)" = "0" ]; then
                chown -R app:app "$dir" 2>/dev/null || true
            fi
        fi
    else
        # Create directory if it doesn't exist
        mkdir -p "$dir" 2>/dev/null || true
        if [ "$(id -u)" = "0" ]; then
            chown -R app:app "$dir" 2>/dev/null || true
        fi
    fi
}

# Fix permissions for data directories
fix_permissions "/app/data"
fix_permissions "/app/logs"

# If running as root, switch to app user
if [ "$(id -u)" = "0" ]; then
    echo "Starting as app user..."
    exec su-exec app "$@"
else
    echo "Starting as current user ($(id -u))..."
    exec "$@"
fi
