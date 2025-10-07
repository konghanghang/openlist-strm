#!/bin/sh
set -e

# Default PUID and PGID
PUID=${PUID:-1000}
PGID=${PGID:-1000}

echo "-----------------------------------"
echo "  OpenList-STRM Docker Container  "
echo "-----------------------------------"
echo "User UID: $PUID"
echo "User GID: $PGID"
echo "-----------------------------------"

# Function to fix permissions for a directory
fix_permissions() {
    local dir=$1
    if [ -d "$dir" ]; then
        echo "Fixing permissions for $dir..."
        chown -R app:app "$dir" 2>/dev/null || echo "Warning: Cannot change ownership of $dir"
    else
        # Create directory if it doesn't exist
        echo "Creating directory $dir..."
        mkdir -p "$dir" 2>/dev/null || echo "Warning: Cannot create $dir"
        chown -R app:app "$dir" 2>/dev/null || true
    fi
}

# Only run as root to setup user and permissions
if [ "$(id -u)" = "0" ]; then
    # Modify app user's UID and GID to match PUID/PGID
    echo "Adjusting user app to UID:$PUID and GID:$PGID..."

    # Change GID if different
    if [ "$(id -g app)" != "$PGID" ]; then
        groupmod -o -g "$PGID" app
    fi

    # Change UID if different
    if [ "$(id -u app)" != "$PUID" ]; then
        usermod -o -u "$PUID" app
    fi

    # Fix permissions for data directories
    fix_permissions "/app/data"
    fix_permissions "/app/logs"
    fix_permissions "/app/configs"

    # Fix permissions for STRM mount point if it exists
    if [ -d "/mnt/strm" ]; then
        fix_permissions "/mnt/strm"
    fi

    echo "Starting application as app user (UID:$PUID GID:$PGID)..."
    exec su-exec app "$@"
else
    echo "Starting as current user ($(id -u):$(id -g))..."
    exec "$@"
fi
