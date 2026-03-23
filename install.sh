#!/bin/bash
set -e

echo ""
echo "  ╔═══════════════════════════════════════════════════════╗"
echo "  ║           GROVE - Instalador v1.3.0                 ║"
echo "  ╚═══════════════════════════════════════════════════════╝"
echo ""

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

PLATFORM="${OS}-${ARCH}"
echo "Platform: $PLATFORM"

# Create temp directory
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "[1/4] Descargando GROVE..."
DOWNLOAD_URL="https://github.com/th3bonii/grove/releases/latest/download/grove-${PLATFORM}.zip"

if command -v curl &> /dev/null; then
    curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/grove.zip"
elif command -v wget &> /dev/null; then
    wget -q "$DOWNLOAD_URL" -O "$TEMP_DIR/grove.zip"
else
    echo "ERROR: curl o wget requerido"
    exit 1
fi

echo "[2/4] Extrayendo..."
unzip -q "$TEMP_DIR/grove.zip" -d "$TEMP_DIR/grove"

echo "[3/4] Instalando..."
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

cp "$TEMP_DIR/grove/grove-spec" "$INSTALL_DIR/grove-spec"
cp "$TEMP_DIR/grove/grove-loop" "$INSTALL_DIR/grove-loop"
cp "$TEMP_DIR/grove/grove-opti" "$INSTALL_DIR/grove-opti"
chmod +x "$INSTALL_DIR/grove-"*

# Add to PATH if not already
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "" >> "$HOME/.bashrc"
    echo "# GROVE" >> "$HOME/.bashrc"
    echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.bashrc"
    
    if [ -f "$HOME/.zshrc" ]; then
        echo "" >> "$HOME/.zshrc"
        echo "# GROVE" >> "$HOME/.zshrc"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$HOME/.zshrc"
    fi
fi

echo "[4/4] Instalando skills..."
SKILLS_DIR="$HOME/.config/opencode/skills"
mkdir -p "$SKILLS_DIR"

if [ -d "$TEMP_DIR/grove/skills" ]; then
    cp -r "$TEMP_DIR/grove/skills/"* "$SKILLS_DIR/"
fi

echo ""
echo "  ╔═══════════════════════════════════════════════════════╗"
echo "  ║           ✅ GROVE INSTALADO                         ║"
echo "  ╚═══════════════════════════════════════════════════════╝"
echo ""
echo "  Ubicación: $INSTALL_DIR"
echo ""
echo "  Cierra esta terminal y abre una nueva."
echo ""
echo "  Luego prueba:"
echo "    grove-spec --help"
echo ""
