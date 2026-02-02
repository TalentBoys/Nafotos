#!/bin/bash

# AwesomeSharing v2.0 Local Run Script

# Set colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║        Starting AwesomeSharing v2.0                   ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
echo ""

# Load .env.local if it exists (in parent directory)
if [ -f "../.env.local" ]; then
    export $(grep -v '^#' ../.env.local | xargs)
fi

# Set environment variables with defaults
export CONFIG_DIR="./config"
export UPLOAD_DIR="./upload"
export PORT="${BACKEND_PORT:-8080}"
export ALLOWED_ORIGIN="http://localhost:${FRONTEND_PORT:-3000}"

# Create directories if they don't exist
mkdir -p $CONFIG_DIR
mkdir -p $UPLOAD_DIR

echo -e "${YELLOW}Configuration:${NC}"
echo "  CONFIG_DIR: $CONFIG_DIR"
echo "  UPLOAD_DIR: $UPLOAD_DIR"
echo "  PORT: $PORT"
echo ""

# Check if database exists
if [ -f "$CONFIG_DIR/awesome-sharing.db" ]; then
    echo -e "${GREEN}✓ Database found${NC}"
else
    echo -e "${YELLOW}⚠ Database will be created on first run${NC}"
fi

echo ""
echo -e "${GREEN}Starting server...${NC}"
echo ""

# Run the server
go run cmd/server/main.go
