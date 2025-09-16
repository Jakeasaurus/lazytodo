#!/bin/bash

# lazytodo installer script
# Installs lazytodo to a proper system location

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Synthwave banner
echo -e "${PURPLE}┌─────────────────────────────────────────┐${NC}"
echo -e "${PURPLE}│   ${CYAN}⚡ LAZYTODO INSTALLER ${PURPLE}│${NC}"
echo -e "${PURPLE}│   ${CYAN}Neon Todo Management ${PURPLE}│${NC}"  
echo -e "${PURPLE}└─────────────────────────────────────────┘${NC}"
echo ""

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

# Default installation directory
if [[ "$OS" == "Darwin" ]]; then
    DEFAULT_INSTALL_DIR="/usr/local/bin"
elif [[ "$OS" == "Linux" ]]; then
    if [[ -d "$HOME/.local/bin" ]]; then
        DEFAULT_INSTALL_DIR="$HOME/.local/bin"
    else
        DEFAULT_INSTALL_DIR="/usr/local/bin"
    fi
else
    echo -e "${RED}❌ Unsupported OS: $OS${NC}"
    exit 1
fi

# Allow user to override install directory
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_INSTALL_DIR}"

echo -e "${CYAN}🌆 Installing lazytodo...${NC}"
echo -e "${BLUE}📍 Target directory: ${YELLOW}$INSTALL_DIR${NC}"
echo ""

# Check if we need sudo
NEED_SUDO=false
if [[ ! -w "$INSTALL_DIR" ]]; then
    NEED_SUDO=true
    echo -e "${YELLOW}⚠️  Need sudo access to write to $INSTALL_DIR${NC}"
fi

# Check if lazytodo binary exists
if [[ ! -f "./lazytodo" ]]; then
    echo -e "${RED}❌ lazytodo binary not found in current directory${NC}"
    echo -e "${BLUE}💡 Run 'go build -o lazytodo' first${NC}"
    exit 1
fi

# Create install directory if it doesn't exist
if [[ "$NEED_SUDO" == "true" ]]; then
    sudo mkdir -p "$INSTALL_DIR"
else
    mkdir -p "$INSTALL_DIR"
fi

# Copy binary
echo -e "${CYAN}📦 Installing binary...${NC}"
if [[ "$NEED_SUDO" == "true" ]]; then
    sudo cp "./lazytodo" "$INSTALL_DIR/lazytodo"
    sudo chmod +x "$INSTALL_DIR/lazytodo"
else
    cp "./lazytodo" "$INSTALL_DIR/lazytodo"
    chmod +x "$INSTALL_DIR/lazytodo"
fi

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo -e "${YELLOW}⚠️  $INSTALL_DIR is not in your PATH${NC}"
    echo -e "${BLUE}💡 Add this to your shell profile:${NC}"
    echo -e "${CYAN}export PATH=\"\$PATH:$INSTALL_DIR\"${NC}"
    echo ""
    
    # Offer to add to PATH automatically
    if [[ "$OS" == "Darwin" ]] && [[ "$SHELL" == *"zsh" ]]; then
        read -p "$(echo -e ${BLUE}🔧 Add to PATH automatically? [y/N]: ${NC})" -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> ~/.zshrc
            echo -e "${GREEN}✅ Added to ~/.zshrc${NC}"
            echo -e "${BLUE}💡 Run 'source ~/.zshrc' or restart your terminal${NC}"
        fi
    fi
else
    echo -e "${GREEN}✅ $INSTALL_DIR is in your PATH${NC}"
fi

# Test installation
echo ""
echo -e "${CYAN}🧪 Testing installation...${NC}"
if command -v lazytodo >/dev/null 2>&1; then
    echo -e "${GREEN}✅ lazytodo is accessible from PATH${NC}"
    echo -e "${BLUE}📊 Version: ${NC}$(lazytodo --version)"
else
    echo -e "${YELLOW}⚠️  lazytodo not found in PATH (may need to restart terminal)${NC}"
fi

# Check for todo.txt-cli
echo ""
echo -e "${CYAN}🔍 Checking prerequisites...${NC}"
if command -v todo.sh >/dev/null 2>&1 || command -v todo >/dev/null 2>&1; then
    echo -e "${GREEN}✅ todo.txt-cli is installed${NC}"
else
    echo -e "${YELLOW}⚠️  todo.txt-cli not found${NC}"
    echo -e "${BLUE}💡 Install with: ${CYAN}brew install todo-txt${NC} (macOS) or package manager"
fi

# Check for todo files
echo ""
echo -e "${CYAN}📁 Checking todo.txt setup...${NC}"
if [[ -f "$HOME/todo.txt" ]]; then
    echo -e "${GREEN}✅ Found todo.txt at $HOME/todo.txt${NC}"
elif [[ -f "$HOME/.todo/config" ]]; then
    echo -e "${GREEN}✅ Found todo.txt config at $HOME/.todo/config${NC}"
else
    echo -e "${YELLOW}ℹ️  No todo.txt found - will be created on first use${NC}"
fi

# Success message
echo ""
echo -e "${PURPLE}┌─────────────────────────────────────────┐${NC}"
echo -e "${PURPLE}│   ${GREEN}⚡ INSTALLATION COMPLETE! ${PURPLE}│${NC}"
echo -e "${PURPLE}└─────────────────────────────────────────┘${NC}"
echo ""
echo -e "${CYAN}🚀 Run ${YELLOW}lazytodo${CYAN} to start the neon todo experience!${NC}"
echo -e "${BLUE}💡 Run ${YELLOW}lazytodo --help${BLUE} for usage information${NC}"
echo ""
echo -e "${PURPLE}Welcome to the electric future of productivity! 🌆⚡${NC}"