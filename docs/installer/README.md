# QUIVer One-Click Installer

This directory contains the one-click installer solution for QUIVer Provider on macOS.

## Overview

The installer provides a seamless installation experience that:
- Bypasses macOS "damaged app" errors through proper code signing and notarization
- Provides both web-based and direct download installation options
- Automatically configures and starts QUIVer Provider
- Sets up LaunchAgent for automatic startup

## Components

### 1. Notarization Script (`macos/notarize.sh`)
Handles Apple code signing and notarization for distribution:
- Signs binaries and packages with Developer ID
- Submits to Apple for notarization
- Staples notarization ticket for offline verification

### 2. PKG Builder Script (`macos/build-pkg.sh`)
Creates a signed macOS installer package:
- Bundles provider and gateway binaries
- Creates LaunchAgent for auto-start
- Includes default configuration files
- Generates installer UI with welcome/conclusion screens

### 3. Web Installer (`web/install.html`)
Browser-based installer with automatic detection:
- Detects if QUIVer is already installed
- One-click download and installation
- Real-time installation progress tracking
- Automatic launch after installation

### 4. Website Landing Page (`../website/index.html`)
Marketing website with install call-to-action:
- Platform detection (macOS, Windows, Linux)
- Modal for choosing installation method
- Feature highlights and how-it-works section
- Responsive design for all devices

### 5. GitHub Actions Workflow (`.github/workflows/release-macos.yml`)
Automated build and release pipeline:
- Builds and signs installer on release
- Uploads to GitHub releases
- Deploys to CDN for web installer
- Handles certificate management securely

## Setup Instructions

### Prerequisites

1. **Apple Developer Account** with Developer ID certificate
2. **Environment Variables**:
   ```bash
   export APPLE_DEVELOPER_ID="Developer ID Installer: Your Name (TEAMID)"
   export APPLE_TEAM_ID="YOURTEAMID"
   export APPLE_ID="your@email.com"
   export APP_PASSWORD="xxxx-xxxx-xxxx-xxxx"  # App-specific password
   ```

### Building Locally

1. Build the binaries first:
   ```bash
   cd ../../provider && go build -o build/provider ./cmd/provider
   cd ../gateway && go build -o build/gateway ./cmd/gateway
   ```

2. Create the installer:
   ```bash
   cd docs/installer/macos
   ./build-pkg.sh
   ```

3. The signed installer will be in `dist/QUIVerProvider-signed.pkg`

### GitHub Actions Setup

Add these secrets to your GitHub repository:
- `APPLE_CERT_BASE64`: Base64 encoded .p12 certificate
- `APPLE_CERT_PASSWORD`: Certificate password
- `APPLE_DEVELOPER_ID`: Developer ID string
- `APPLE_TEAM_ID`: Apple Developer Team ID
- `APPLE_ID`: Apple ID email
- `APP_PASSWORD`: App-specific password
- `AWS_ACCESS_KEY_ID`: For CDN upload (optional)
- `AWS_SECRET_ACCESS_KEY`: For CDN upload (optional)
- `CLOUDFRONT_DISTRIBUTION_ID`: For cache invalidation (optional)

### Testing the Web Installer

1. Serve the website locally:
   ```bash
   cd docs/website
   python3 -m http.server 8000
   ```

2. Open http://localhost:8000 in Safari

3. Click "Install for macOS" to test the flow

## Security Considerations

- All installers are signed with Apple Developer ID
- Packages are notarized by Apple for Gatekeeper approval
- Web installer uses HTTPS for all downloads
- No sensitive data is logged or transmitted
- LaunchAgent runs with user privileges only

## Troubleshooting

### "Damaged App" Error
If users still see this error:
1. Ensure the installer is properly signed and notarized
2. Check that stapling was successful
3. Verify the download wasn't corrupted

### Notarization Failures
Common causes:
- Missing entitlements
- Unsigned binaries in package
- Network issues during submission
- Invalid certificate

### Installation Issues
- Check Console.app for detailed error logs
- Verify LaunchAgent permissions
- Ensure Ollama is installed and running
- Check firewall settings for P2P connectivity

## Future Enhancements

- [ ] Windows installer with MSI package
- [ ] Linux packages (deb, rpm, AppImage)
- [ ] Auto-update mechanism
- [ ] Uninstaller tool
- [ ] Integration with system preferences