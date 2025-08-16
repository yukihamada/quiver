#!/bin/bash
set -euo pipefail

# QUIVer Provider macOS Notarization Script
# This script handles code signing and notarization for macOS distribution

# Configuration - these should be set as environment variables
DEVELOPER_ID="${APPLE_DEVELOPER_ID:-}"
TEAM_ID="${APPLE_TEAM_ID:-}"
APP_BUNDLE_ID="${APP_BUNDLE_ID:-com.quiver.provider}"
APPLE_ID="${APPLE_ID:-}"
APP_PASSWORD="${APP_PASSWORD:-}"  # App-specific password from appleid.apple.com

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[STATUS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Xcode command line tools are installed
    if ! command -v xcrun &> /dev/null; then
        print_error "Xcode command line tools not found. Please install them first."
        exit 1
    fi
    
    # Check environment variables
    if [[ -z "$DEVELOPER_ID" ]]; then
        print_error "APPLE_DEVELOPER_ID environment variable not set"
        exit 1
    fi
    
    if [[ -z "$TEAM_ID" ]]; then
        print_error "APPLE_TEAM_ID environment variable not set"
        exit 1
    fi
    
    if [[ -z "$APPLE_ID" ]]; then
        print_error "APPLE_ID environment variable not set"
        exit 1
    fi
    
    if [[ -z "$APP_PASSWORD" ]]; then
        print_error "APP_PASSWORD environment variable not set"
        exit 1
    fi
    
    print_status "Prerequisites check passed"
}

# Sign a binary or package
sign_binary() {
    local file="$1"
    print_status "Signing $file..."
    
    codesign --force --options runtime \
        --sign "$DEVELOPER_ID" \
        --timestamp \
        --identifier "$APP_BUNDLE_ID" \
        "$file"
    
    # Verify signature
    codesign --verify --deep --strict "$file"
    print_status "Successfully signed $file"
}

# Sign a PKG installer
sign_pkg() {
    local pkg_file="$1"
    local signed_pkg="${pkg_file%.pkg}-signed.pkg"
    
    print_status "Signing package $pkg_file..."
    
    productsign --sign "$DEVELOPER_ID" \
        --timestamp \
        "$pkg_file" "$signed_pkg"
    
    # Verify signature
    pkgutil --check-signature "$signed_pkg"
    print_status "Successfully signed package as $signed_pkg"
    
    echo "$signed_pkg"
}

# Create a notarization request
notarize_file() {
    local file="$1"
    local request_uuid=""
    
    print_status "Submitting $file for notarization..."
    
    # Submit for notarization
    request_uuid=$(xcrun altool --notarize-app \
        --primary-bundle-id "$APP_BUNDLE_ID" \
        --username "$APPLE_ID" \
        --password "$APP_PASSWORD" \
        --team-id "$TEAM_ID" \
        --file "$file" 2>&1 | grep RequestUUID | awk '{print $3}')
    
    if [[ -z "$request_uuid" ]]; then
        print_error "Failed to submit for notarization"
        exit 1
    fi
    
    print_status "Notarization request submitted with UUID: $request_uuid"
    
    # Wait for notarization to complete
    wait_for_notarization "$request_uuid"
    
    # Staple the notarization ticket
    print_status "Stapling notarization ticket..."
    xcrun stapler staple "$file"
    
    print_status "Successfully notarized and stapled $file"
}

# Wait for notarization to complete
wait_for_notarization() {
    local request_uuid="$1"
    local status="in progress"
    local attempt=0
    local max_attempts=60  # 30 minutes timeout (30 seconds * 60)
    
    print_status "Waiting for notarization to complete..."
    
    while [[ "$status" == "in progress" ]] && [[ $attempt -lt $max_attempts ]]; do
        sleep 30
        attempt=$((attempt + 1))
        
        status=$(xcrun altool --notarization-info "$request_uuid" \
            --username "$APPLE_ID" \
            --password "$APP_PASSWORD" 2>&1 | grep "Status:" | awk '{print $2, $3}')
        
        print_status "Notarization status: $status (attempt $attempt/$max_attempts)"
    done
    
    if [[ "$status" != "success" ]]; then
        print_error "Notarization failed with status: $status"
        
        # Get detailed log
        xcrun altool --notarization-info "$request_uuid" \
            --username "$APPLE_ID" \
            --password "$APP_PASSWORD"
        
        exit 1
    fi
}

# Main function to sign and notarize
sign_and_notarize() {
    local input_file="$1"
    local file_type="${input_file##*.}"
    
    check_prerequisites
    
    case "$file_type" in
        "pkg")
            # Sign PKG
            signed_file=$(sign_pkg "$input_file")
            # Notarize PKG
            notarize_file "$signed_file"
            print_status "Output: $signed_file"
            ;;
        "dmg")
            # Sign DMG contents first if needed
            sign_binary "$input_file"
            # Notarize DMG
            notarize_file "$input_file"
            print_status "Output: $input_file"
            ;;
        "app"|"zip")
            # For .app bundles or zipped apps
            sign_binary "$input_file"
            notarize_file "$input_file"
            print_status "Output: $input_file"
            ;;
        *)
            # For raw binaries
            sign_binary "$input_file"
            # Create a zip for notarization
            zip_file="${input_file}.zip"
            zip -j "$zip_file" "$input_file"
            notarize_file "$zip_file"
            rm "$zip_file"
            print_status "Output: $input_file"
            ;;
    esac
}

# Script usage
usage() {
    echo "Usage: $0 <file_to_sign_and_notarize>"
    echo ""
    echo "Environment variables required:"
    echo "  APPLE_DEVELOPER_ID - Developer ID certificate name"
    echo "  APPLE_TEAM_ID - Apple Developer Team ID"
    echo "  APPLE_ID - Apple ID email"
    echo "  APP_PASSWORD - App-specific password"
    echo "  APP_BUNDLE_ID - Bundle identifier (default: com.quiver.provider)"
    echo ""
    echo "Supported file types: .pkg, .dmg, .app, .zip, or raw binaries"
}

# Main execution
if [[ $# -ne 1 ]]; then
    usage
    exit 1
fi

if [[ ! -f "$1" ]]; then
    print_error "File not found: $1"
    exit 1
fi

sign_and_notarize "$1"