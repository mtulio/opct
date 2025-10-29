#!/bin/bash

# OPCT Session Initializer Script
# Extracts OPCT result tarballs and organizes them into structured sessions
#
# Usage: init_session.sh <tarball_file> [session_dir]
#   tarball_file: Path to OPCT result tarball (.tar.gz)
#   session_dir: Optional custom session directory name

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Function to show usage
show_usage() {
    echo "Usage: $0 <tarball_file> [session_dir]"
    echo ""
    echo "Arguments:"
    echo "  tarball_file  Path to OPCT result tarball (.tar.gz)"
    echo "  session_dir   Optional custom session directory name"
    echo ""
    echo "Examples:"
    echo "  $0 opct_results.tar.gz"
    echo "  $0 opct_results.tar.gz my-analysis-session"
    echo "  $0 /path/to/opct_202510091157_67175c7a.tar.gz"
}

# Function to validate tarball file
validate_tarball() {
    local tarball_file="$1"

    if [ ! -f "$tarball_file" ]; then
        print_error "Tarball file '$tarball_file' not found"
        return 1
    fi

    if [ ! -r "$tarball_file" ]; then
        print_error "Cannot read tarball file '$tarball_file' (permission denied)"
        return 1
    fi

    # Check if it's a valid tar.gz file
    if ! tar -tzf "$tarball_file" >/dev/null 2>&1; then
        print_error "Invalid or corrupted tarball file '$tarball_file'"
        return 1
    fi

    print_success "Tarball file validated: $tarball_file"
    return 0
}

# Function to determine session directory name
get_session_dir() {
    local tarball_file="$1"
    local custom_session_dir="$2"

    if [ -n "$custom_session_dir" ]; then
        echo "$custom_session_dir"
    else
        # Extract base name without .tar.gz extension
        basename "$tarball_file" .tar.gz
    fi
}

# Function to check for existing session directory
check_existing_session() {
    local session_dir="$1"

    if [ -d "$session_dir" ] && [ "$(ls -A "$session_dir" 2>/dev/null)" ]; then
        print_warning "Session directory '$session_dir' already exists and is not empty"
        echo "This will overwrite existing data. Continue? (y/N)"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_info "Operation cancelled by user"
            exit 0
        fi
        print_info "Proceeding with overwrite..."
    fi
}

# Function to create session structure
create_session_structure() {
    local session_dir="$1"

    print_info "Creating session directory structure..."

    # Create main session directory
    mkdir -p "$session_dir"

    # Create subdirectories
    mkdir -p "$session_dir/results"
    mkdir -p "$session_dir/must-gather"

    print_success "Session structure created: $session_dir"
}

# Function to extract OPCT results
extract_opct_results() {
    local tarball_file="$1"
    local session_dir="$2"

    print_info "Extracting OPCT results from '$tarball_file'..."

    # Extract entire tarball to results directory
    if tar -xzf "$tarball_file" -C "$session_dir/results"; then
        print_success "OPCT results extracted to: $session_dir/results/"
    else
        print_error "Failed to extract OPCT results from '$tarball_file'"
        return 1
    fi

    # Verify extraction
    if [ ! "$(ls -A "$session_dir/results" 2>/dev/null)" ]; then
        print_error "Results directory is empty after extraction"
        return 1
    fi

    return 0
}

# Function to extract must-gather data
extract_must_gather() {
    local session_dir="$1"

    # Expected location of must-gather archive
    local must_gather_archive="$session_dir/results/plugins/99-openshift-artifacts-collector/results/global/artifacts_must-gather.tar.xz"

    print_info "Looking for must-gather archive..."

    if [ ! -f "$must_gather_archive" ]; then
        print_warning "Must-gather archive not found at expected location:"
        print_warning "  $must_gather_archive"
        print_warning "Skipping must-gather extraction"
        return 0
    fi

    print_info "Found must-gather archive, extracting..."

    # Extract must-gather to dedicated directory
    if tar -xf "$must_gather_archive" -C "$session_dir/must-gather"; then
        print_success "Must-gather data extracted to: $session_dir/must-gather/"
    else
        print_error "Failed to extract must-gather data"
        return 1
    fi

    # Verify must-gather extraction
    if [ ! "$(ls -A "$session_dir/must-gather" 2>/dev/null)" ]; then
        print_warning "Must-gather directory is empty after extraction"
    fi

    return 0
}

# Function to verify session creation
verify_session() {
    local session_dir="$1"

    print_info "Verifying session creation..."

    local has_results=false
    local has_must_gather=false

    # Check results directory
    if [ -d "$session_dir/results" ] && [ "$(ls -A "$session_dir/results" 2>/dev/null)" ]; then
        has_results=true
        print_success "✓ OPCT results available"
    else
        print_error "✗ OPCT results not found"
    fi

    # Check must-gather directory
    if [ -d "$session_dir/must-gather" ] && [ "$(ls -A "$session_dir/must-gather" 2>/dev/null)" ]; then
        has_must_gather=true
        print_success "✓ Must-gather data available"
    else
        print_warning "✗ Must-gather data not available"
    fi

    return 0
}

# Function to show session summary
show_session_summary() {
    local session_dir="$1"
    local tarball_file="$2"

    echo ""
    echo "=================================================================================="
    echo "OPCT SESSION INITIALIZED SUCCESSFULLY"
    echo "=================================================================================="
    echo ""
    echo "Session Directory: $(realpath "$session_dir")"
    echo "Source Tarball:    $tarball_file"
    echo ""
    echo "Extracted Contents:"
    echo "  Results:         $session_dir/results/"
    echo "  Must-Gather:     $session_dir/must-gather/"
    echo ""
    echo "Next Steps:"
    echo "  - Analyze OPCT results: /opct:analyze $session_dir/results/"
    echo "  - Analyze must-gather:  /must-gather:analyze $session_dir/must-gather/"
    echo "  - Generate report:      /opct:report $session_dir/"
    echo ""
    echo "=================================================================================="
}

# Main function
main() {
    local tarball_file="$1"
    local custom_session_dir="${2:-}"

    print_info "OPCT Session Initializer"
    print_info "========================"

    # Validate input
    if [ $# -lt 1 ] || [ $# -gt 2 ]; then
        print_error "Invalid number of arguments"
        show_usage
        exit 1
    fi

    # Validate tarball file
    if ! validate_tarball "$tarball_file"; then
        exit 1
    fi

    # Determine session directory
    local session_dir
    session_dir=$(get_session_dir "$tarball_file" "$custom_session_dir")
    print_info "Session directory: $session_dir"

    # Check for existing session
    check_existing_session "$session_dir"

    # Create session structure
    create_session_structure "$session_dir"

    # Extract OPCT results
    if ! extract_opct_results "$tarball_file" "$session_dir"; then
        exit 1
    fi

    # Extract must-gather data
    extract_must_gather "$session_dir"

    # Verify session creation
    verify_session "$session_dir"

    # Show summary
    show_session_summary "$session_dir" "$tarball_file"

    print_success "OPCT session initialization completed successfully!"
}

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
