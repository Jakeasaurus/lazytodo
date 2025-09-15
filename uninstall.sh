#!/bin/bash

# lazytodo uninstaller script
# Removes lazytodo from system locations

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
echo -e "${PURPLE}│   ${RED}🗑️  LAZYTODO UNINSTALLER ${PURPLE}│${NC}"
echo -e "${PURPLE}│   ${RED}Removing Electric Dreams ${PURPLE}│${NC}"  
echo -e "${PURPLE}└─────────────────────────────────────────┘${NC}"
echo ""

# Common installation locations
COMMON_LOCATIONS=(
    "/usr/local/bin/lazytodo"
    "$HOME/.local/bin/lazytodo"
    "/usr/bin/lazytodo"
    "/opt/local/bin/lazytodo"
)

echo -e "${CYAN}🔍 Searching for lazytodo installations...${NC}"

FOUND_LOCATIONS=()
for location in "${COMMON_LOCATIONS[@]}"; do
    if [[ -f "$location" ]]; then
        FOUND_LOCATIONS+=("$location")
        echo -e "${YELLOW}📍 Found: ${location}${NC}"
    fi
done

# Also check for lazytodo in PATH
if command -v lazytodo >/dev/null 2>&1; then
    WHICH_LOCATION=$(which lazytodo)
    if [[ ! " ${FOUND_LOCATIONS[@]} " =~ " ${WHICH_LOCATION} " ]]; then
        FOUND_LOCATIONS+=("$WHICH_LOCATION")
        echo -e "${YELLOW}📍 Found in PATH: ${WHICH_LOCATION}${NC}"
    fi
fi

if [[ ${#FOUND_LOCATIONS[@]} -eq 0 ]]; then
    echo -e "${GREEN}ℹ️  No lazytodo installations found${NC}"
    echo -e "${BLUE}💡 lazytodo may already be uninstalled${NC}"
    exit 0
fi

echo ""
echo -e "${RED}⚠️  The following installations will be removed:${NC}"
for location in "${FOUND_LOCATIONS[@]}"; do
    echo -e "  ${RED}✗ ${location}${NC}"
done

echo ""
read -p "$(echo -e ${YELLOW}❓ Continue with uninstallation? [y/N]: ${NC})" -n 1 -r
echo

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}ℹ️  Uninstallation cancelled${NC}"
    exit 0
fi

echo ""
echo -e "${CYAN}🗑️  Removing lazytodo installations...${NC}"

for location in "${FOUND_LOCATIONS[@]}"; do
    if [[ -f "$location" ]]; then
        dir=$(dirname "$location")
        if [[ -w "$dir" ]]; then
            rm -f "$location"
            echo -e "${GREEN}✅ Removed: ${location}${NC}"
        else
            if sudo rm -f "$location" 2>/dev/null; then
                echo -e "${GREEN}✅ Removed (sudo): ${location}${NC}"
            else
                echo -e "${RED}❌ Failed to remove: ${location}${NC}"
            fi
        fi
    else
        echo -e "${YELLOW}⚠️  Not found (may have been removed): ${location}${NC}"
    fi
done

# Check if any installations remain
echo ""
echo -e "${CYAN}🔍 Verifying removal...${NC}"
if command -v lazytodo >/dev/null 2>&1; then
    REMAINING=$(which lazytodo)
    echo -e "${YELLOW}⚠️  lazytodo still found at: ${REMAINING}${NC}"
    echo -e "${BLUE}💡 You may need to manually remove it or check your PATH${NC}"
else
    echo -e "${GREEN}✅ lazytodo successfully removed from PATH${NC}"
fi

# Cleanup note
echo ""
echo -e "${BLUE}📝 Note: Your todo.txt files and configuration remain untouched${NC}"
echo -e "${BLUE}   These are located at:${NC}"
if [[ -f "$HOME/todo.txt" ]]; then
    echo -e "${BLUE}   - ~/todo.txt${NC}"
fi
if [[ -d "$HOME/.todo" ]]; then
    echo -e "${BLUE}   - ~/.todo/${NC}"
fi

# Success message
echo ""
echo -e "${PURPLE}┌─────────────────────────────────────────┐${NC}"
echo -e "${PURPLE}│   ${GREEN}✅ UNINSTALL COMPLETE ${PURPLE}│${NC}"
echo -e "${PURPLE}└─────────────────────────────────────────┘${NC}"
echo ""
echo -e "${CYAN}Thanks for trying lazytodo! 🌆${NC}"
echo -e "${BLUE}💡 Your todo.txt files remain safe for other tools${NC}"
echo ""
echo -e "${PURPLE}Until we meet again in the neon grid... ⚡${NC}"