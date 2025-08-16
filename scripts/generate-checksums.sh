#!/bin/bash

set -e

echo "Generating checksums for QUIVer binaries..."

# Create checksums directory
mkdir -p checksums

# Function to generate checksums for a binary
generate_checksum() {
    local binary_path=$1
    local binary_name=$(basename "$binary_path")
    
    if [ -f "$binary_path" ]; then
        echo "Generating checksum for $binary_name..."
        
        # Generate SHA256 checksum
        if [[ "$OSTYPE" == "darwin"* ]]; then
            shasum -a 256 "$binary_path" > "checksums/${binary_name}.sha256"
        else
            sha256sum "$binary_path" > "checksums/${binary_name}.sha256"
        fi
        
        # Generate SHA512 checksum
        if [[ "$OSTYPE" == "darwin"* ]]; then
            shasum -a 512 "$binary_path" > "checksums/${binary_name}.sha512"
        else
            sha512sum "$binary_path" > "checksums/${binary_name}.sha512"
        fi
        
        echo "✓ Checksums generated for $binary_name"
    else
        echo "✗ Binary not found: $binary_path"
    fi
}

# Build all binaries first
echo "Building all binaries..."
make build-all

# Generate checksums for all binaries
generate_checksum "provider/provider"
generate_checksum "gateway/gateway"
generate_checksum "aggregator/aggregator"
generate_checksum "bootstrap/bootstrap"

# Generate checksums for macOS builds if they exist
if [ -d "build/darwin" ]; then
    for binary in build/darwin/*; do
        generate_checksum "$binary"
    done
fi

# Generate checksums for Linux builds if they exist
if [ -d "build/linux" ]; then
    for binary in build/linux/*; do
        generate_checksum "$binary"
    done
fi

# Create a manifest file
cat > checksums/MANIFEST.json << EOF
{
  "version": "$(git describe --tags --always 2>/dev/null || echo "dev")",
  "build_date": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "commit": "$(git rev-parse HEAD 2>/dev/null || echo "unknown")",
  "checksums": {
EOF

# Add checksums to manifest
first=true
for checksum_file in checksums/*.sha256; do
    if [ -f "$checksum_file" ]; then
        binary_name=$(basename "$checksum_file" .sha256)
        sha256=$(cut -d' ' -f1 < "$checksum_file")
        sha512=$(cut -d' ' -f1 < "checksums/${binary_name}.sha512")
        
        if [ "$first" = true ]; then
            first=false
        else
            echo "," >> checksums/MANIFEST.json
        fi
        
        cat >> checksums/MANIFEST.json << EOF
    "$binary_name": {
      "sha256": "$sha256",
      "sha512": "$sha512"
    }
EOF
    fi
done

cat >> checksums/MANIFEST.json << EOF

  }
}
EOF

echo ""
echo "Checksum generation complete!"
echo "Checksums saved to checksums/"
echo ""
echo "To verify a binary, run:"
echo "  shasum -a 256 -c checksums/[binary_name].sha256"