#!/bin/bash

echo "🔐 Testing QUIVer Gateway Authentication"
echo "======================================"

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Test without authentication
echo -e "\n${YELLOW}1. Testing without authentication...${NC}"
response=$(curl -s -X POST "$GATEWAY_URL/generate" \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "model": "llama3.2:3b"}' \
  -w "\nHTTP_STATUS:%{http_code}")

http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS:/d')

if [ "$http_status" = "200" ]; then
  echo -e "${GREEN}✓ Request succeeded without auth (auth disabled)${NC}"
elif [ "$http_status" = "401" ]; then
  echo -e "${GREEN}✓ Request correctly rejected without auth${NC}"
else
  echo -e "${RED}✗ Unexpected status: $http_status${NC}"
fi
echo "Response: $body"

# Generate test JWT token
echo -e "\n${YELLOW}2. Generating test JWT token...${NC}"
JWT_SECRET="quiver-secret-key-change-in-production"
header='{"alg":"HS256","typ":"JWT"}'
payload='{"user_id":"test_user","plan":"pro","exp":'$(($(date +%s) + 3600))'}'

# Base64 encode
header_b64=$(echo -n "$header" | base64 | tr -d '=' | tr '/+' '_-')
payload_b64=$(echo -n "$payload" | base64 | tr -d '=' | tr '/+' '_-')

# Create signature
signature=$(echo -n "${header_b64}.${payload_b64}" | openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | base64 | tr -d '=' | tr '/+' '_-')

jwt_token="${header_b64}.${payload_b64}.${signature}"
echo "JWT Token: ${jwt_token:0:50}..."

# Test with JWT authentication
echo -e "\n${YELLOW}3. Testing with JWT authentication...${NC}"
response=$(curl -s -X POST "$GATEWAY_URL/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $jwt_token" \
  -d '{"prompt": "Hello", "model": "llama3.2:3b", "max_tokens": 100}' \
  -w "\nHTTP_STATUS:%{http_code}")

http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS:/d')

if [ "$http_status" = "200" ]; then
  echo -e "${GREEN}✓ Request succeeded with valid JWT${NC}"
else
  echo -e "${RED}✗ Unexpected status: $http_status${NC}"
fi
echo "Response: $body"

# Test with API key
echo -e "\n${YELLOW}4. Testing with API key...${NC}"
# Simulate API key generation
api_key="qvr_test_user_$(date +%s)"
signature=$(echo -n "$api_key" | openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | base64)
full_api_key="${api_key}_${signature}"

response=$(curl -s -X POST "$GATEWAY_URL/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: $full_api_key" \
  -d '{"prompt": "Hello", "model": "llama3.2:3b", "max_tokens": 100}' \
  -w "\nHTTP_STATUS:%{http_code}")

http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_STATUS:/d')

if [ "$http_status" = "200" ]; then
  echo -e "${GREEN}✓ Request succeeded with valid API key${NC}"
else
  echo -e "${RED}✗ Unexpected status: $http_status${NC}"
fi
echo "Response: $body"

# Test rate limiting
echo -e "\n${YELLOW}5. Testing rate limiting...${NC}"
echo "Sending 10 rapid requests..."

success_count=0
rate_limited_count=0

for i in {1..10}; do
  response=$(curl -s -X POST "$GATEWAY_URL/generate" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $jwt_token" \
    -d '{"prompt": "Test", "model": "llama3.2:3b", "max_tokens": 10}' \
    -w "\nHTTP_STATUS:%{http_code}")
  
  http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)
  
  if [ "$http_status" = "200" ]; then
    ((success_count++))
    echo -n "."
  elif [ "$http_status" = "429" ]; then
    ((rate_limited_count++))
    echo -n "!"
  else
    echo -n "?"
  fi
done

echo ""
echo -e "Successful requests: ${GREEN}$success_count${NC}"
echo -e "Rate limited requests: ${YELLOW}$rate_limited_count${NC}"

if [ "$rate_limited_count" -gt 0 ]; then
  echo -e "${GREEN}✓ Rate limiting is working${NC}"
else
  echo -e "${YELLOW}⚠️  Rate limiting may not be configured${NC}"
fi

# Test with invalid token
echo -e "\n${YELLOW}6. Testing with invalid token...${NC}"
response=$(curl -s -X POST "$GATEWAY_URL/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid_token" \
  -d '{"prompt": "Hello", "model": "llama3.2:3b"}' \
  -w "\nHTTP_STATUS:%{http_code}")

http_status=$(echo "$response" | grep -o "HTTP_STATUS:[0-9]*" | cut -d: -f2)

if [ "$http_status" = "401" ]; then
  echo -e "${GREEN}✓ Invalid token correctly rejected${NC}"
else
  echo -e "${RED}✗ Expected 401, got: $http_status${NC}"
fi

echo -e "\n${GREEN}Authentication testing complete!${NC}"