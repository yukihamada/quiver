#!/bin/bash

# Cloudflare DNS Setup Script for quiver.network
# This script requires:
# - CLOUDFLARE_API_TOKEN environment variable
# - CLOUDFLARE_ZONE_ID environment variable

set -e

# Check required environment variables
if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
    echo "Error: CLOUDFLARE_API_TOKEN environment variable is required"
    echo "Please set it with: export CLOUDFLARE_API_TOKEN='your-api-token'"
    exit 1
fi

if [ -z "$CLOUDFLARE_ZONE_ID" ]; then
    echo "Error: CLOUDFLARE_ZONE_ID environment variable is required"
    echo "Please set it with: export CLOUDFLARE_ZONE_ID='your-zone-id'"
    exit 1
fi

# API endpoint
API_ENDPOINT="https://api.cloudflare.com/client/v4/zones/$CLOUDFLARE_ZONE_ID/dns_records"

# Function to create DNS record
create_dns_record() {
    local name=$1
    local content=$2
    local type=${3:-"A"}
    local proxied=${4:-"false"}
    
    echo "Creating $type record: $name.$DOMAIN -> $content"
    
    response=$(curl -s -X POST "$API_ENDPOINT" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        --data "{
            \"type\": \"$type\",
            \"name\": \"$name\",
            \"content\": \"$content\",
            \"ttl\": 1,
            \"proxied\": $proxied
        }")
    
    if echo "$response" | grep -q '"success":true'; then
        echo "✓ Successfully created $name"
    else
        echo "✗ Failed to create $name"
        echo "$response"
    fi
}

# Domain
DOMAIN="quiver.network"

echo "Setting up DNS records for $DOMAIN..."
echo ""

# Load balancer endpoints (proxied through Cloudflare)
create_dns_record "api" "34.117.4.40" "A" "true"
create_dns_record "quiver-global-lb" "34.117.4.40" "A" "true"
create_dns_record "quiver-asia" "34.117.4.40" "A" "true"
create_dns_record "quiver-us" "34.117.4.40" "A" "true"
create_dns_record "quiver-eu" "34.117.4.40" "A" "true"

# Direct endpoints (not proxied for WebSocket/WebRTC)
create_dns_record "signal" "34.146.216.182" "A" "false"
create_dns_record "bootstrap1" "34.146.197.230" "A" "false"
create_dns_record "bootstrap2" "34.146.92.189" "A" "false"
create_dns_record "bootstrap3" "104.198.123.81" "A" "false"

# Stats server
create_dns_record "stats" "34.146.97.202" "A" "false"

echo ""
echo "DNS setup complete!"
echo ""
echo "Next steps:"
echo "1. Wait 1-2 minutes for DNS propagation"
echo "2. Google Cloud SSL certificates will be automatically provisioned (10-15 minutes)"
echo "3. Access the playground at: https://yukihamada.github.io/quiver/playground-stream.html"