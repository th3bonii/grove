#!/bin/bash
set -e

echo "========================================"
echo "GROVE Installation Script"
echo "========================================"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.23 from: https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+')
if (( $(echo "$GO_VERSION < 1.23" | bc -l) )); then
    echo -e "${YELLOW}Warning: Go version $GO_VERSION detected. Recommended: 1.23+${NC}"
fi

# Get script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "Project: $PROJECT_DIR"
echo ""

# Build binaries
echo "Building GROVE..."
cd "$PROJECT_DIR"
mkdir -p bin

go build -ldflags "-X main.version=1.0.0" -o bin/grove-spec ./cmd/grove-spec
go build -ldflags "-X main.version=1.0.0" -o bin/grove-loop ./cmd/grove-loop
go build -ldflags "-X main.version=1.0.0" -o bin/grove-opti ./cmd/grove-opti

echo -e "${GREEN}✓ Binaries built${NC}"

# Install skills
echo ""
echo "Installing GROVE skills to OpenCode..."

SKILLS_DIR="$HOME/.config/opencode/skills"
mkdir -p "$SKILLS_DIR"

if [ -d "$PROJECT_DIR/skills/grove-spec" ]; then
    cp -r "$PROJECT_DIR/skills/grove-spec" "$SKILLS_DIR/"
    echo -e "${GREEN}✓ grove-spec skill installed${NC}"
fi

if [ -d "$PROJECT_DIR/skills/grove-loop" ]; then
    cp -r "$PROJECT_DIR/skills/grove-loop" "$SKILLS_DIR/"
    echo -e "${GREEN}✓ grove-loop skill installed${NC}"
fi

if [ -d "$PROJECT_DIR/skills/grove-opti" ]; then
    cp -r "$PROJECT_DIR/skills/grove-opti" "$SKILLS_DIR/"
    echo -e "${GREEN}✓ grove-opti skill installed${NC}"
fi

# Install binaries
echo ""
echo "Installing binaries..."

BIN_DIR="/usr/local/bin"
if [ -d "$HOME/bin" ]; then
    BIN_DIR="$HOME/bin"
fi

cp "$PROJECT_DIR/bin/grove-spec" "$BIN_DIR/grove-spec"
cp "$PROJECT_DIR/bin/grove-loop" "$BIN_DIR/grove-loop"
cp "$PROJECT_DIR/bin/grove-opti" "$BIN_DIR/grove-opti"

chmod +x "$BIN_DIR/grove-spec"
chmod +x "$BIN_DIR/grove-loop"
chmod +x "$BIN_DIR/grove-opti"

echo -e "${GREEN}✓ Binaries installed to $BIN_DIR${NC}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}GROVE installed successfully!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "Quick start:"
echo "  grove-spec --input ./my-ideas    # Generate specs"
echo "  grove-loop                       # Build from specs"
echo "  grove-opti \"add login button\"    # Optimize prompts"
echo ""
