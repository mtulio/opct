#!/bin/bash

# OPCT Plugin Test Script
# Validates the plugin structure and basic functionality

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PLUGIN_DIR="$SCRIPT_DIR"

print_info "OPCT Plugin Structure Validation"
print_info "================================="

# Check plugin directory structure
print_info "Checking plugin directory structure..."

required_dirs=(
    "commands"
    "skills/init/scripts"
    ".claude-plugin"
)

for dir in "${required_dirs[@]}"; do
    if [ -d "$PLUGIN_DIR/$dir" ]; then
        print_success "✓ Directory exists: $dir"
    else
        print_error "✗ Missing directory: $dir"
        exit 1
    fi
done

# Check required files
print_info "Checking required files..."

required_files=(
    "README.md"
    "commands/init.md"
    "skills/init/SKILL.md"
    "skills/init/scripts/init_session.sh"
    ".claude-plugin/plugin.json"
)

for file in "${required_files[@]}"; do
    if [ -f "$PLUGIN_DIR/$file" ]; then
        print_success "✓ File exists: $file"
    else
        print_error "✗ Missing file: $file"
        exit 1
    fi
done

# Check script permissions
print_info "Checking script permissions..."

if [ -x "$PLUGIN_DIR/skills/init/scripts/init_session.sh" ]; then
    print_success "✓ Script is executable: init_session.sh"
else
    print_warning "! Script is not executable: init_session.sh"
    print_info "Fixing permissions..."
    chmod +x "$PLUGIN_DIR/skills/init/scripts/init_session.sh"
    print_success "✓ Fixed permissions"
fi

# Validate JSON configuration
print_info "Validating plugin configuration..."

if command -v jq >/dev/null 2>&1; then
    if jq empty "$PLUGIN_DIR/.claude-plugin/plugin.json" 2>/dev/null; then
        print_success "✓ Plugin JSON is valid"
    else
        print_error "✗ Plugin JSON is invalid"
        exit 1
    fi
else
    print_warning "! jq not available, skipping JSON validation"
fi

# Check script syntax
print_info "Checking script syntax..."

if bash -n "$PLUGIN_DIR/skills/init/scripts/init_session.sh" 2>/dev/null; then
    print_success "✓ Script syntax is valid"
else
    print_error "✗ Script syntax is invalid"
    exit 1
fi

# Test script help
print_info "Testing script help output..."

if "$PLUGIN_DIR/skills/init/scripts/init_session.sh" 2>&1 | grep -q "Usage:"; then
    print_success "✓ Script help output works"
else
    print_warning "! Script help output may have issues"
fi

print_info "Plugin structure validation completed!"
print_success "OPCT plugin is ready for use!"

echo ""
echo "Next steps:"
echo "1. Test with a real OPCT tarball: /opct:init <tarball-file>"
echo "2. Add additional commands as needed"
echo "3. Integrate with other plugins for enhanced workflows"
