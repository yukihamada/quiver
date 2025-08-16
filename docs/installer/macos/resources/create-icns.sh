#!/bin/bash
# Script to create macOS .icns file from SVG

# Create temporary directory
TEMP_DIR=$(mktemp -d)
ICONSET_DIR="$TEMP_DIR/QUIVer.iconset"
mkdir -p "$ICONSET_DIR"

# Convert SVG to PNG at different sizes
for size in 16 32 64 128 256 512 1024; do
    half_size=$((size / 2))
    
    # Regular resolution
    rsvg-convert -w $size -h $size icon.svg -o "$ICONSET_DIR/icon_${size}x${size}.png"
    
    # Retina resolution (except for 1024)
    if [ $size -ne 1024 ]; then
        rsvg-convert -w $size -h $size icon.svg -o "$ICONSET_DIR/icon_${half_size}x${half_size}@2x.png"
    fi
done

# Create icns file
iconutil -c icns "$ICONSET_DIR" -o QUIVer.icns

# Clean up
rm -rf "$TEMP_DIR"

echo "Created QUIVer.icns successfully"