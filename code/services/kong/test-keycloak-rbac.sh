#!/bin/bash

echo "=== Testing Keycloak RBAC Integration ==="

KEYCLOAK_URL="http://localhost:8080"
REALM="test-realm"
CLIENT_ID="kong-gateway"
CLIENT_SECRET="kong-secret"

# Test 1: Get user token
echo "1. Getting user token..."
USER_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=testuser" \
  -d "password=user123" \
  -d "grant_type=password" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET" | jq -r '.access_token')

if [ "$USER_TOKEN" = "null" ] || [ -z "$USER_TOKEN" ]; then
  echo "Failed to get user token"
  exit 1
fi

echo "Got user token: ${USER_TOKEN:0:50}..."

# Test 2: Decode JWT to see payload
echo "2. Decoding JWT payload..."
JWT_PAYLOAD=$(echo $USER_TOKEN | cut -d'.' -f2)
# Add padding if needed
case $((${#JWT_PAYLOAD} % 4)) in
  2) JWT_PAYLOAD="${JWT_PAYLOAD}==" ;;
  3) JWT_PAYLOAD="${JWT_PAYLOAD}=" ;;
esac

echo $JWT_PAYLOAD | base64 -d | jq .

# Test 3: Test Keycloak Authorization Services (UMA)
echo "3. Testing Keycloak Authorization Services..."
AUTH_RESPONSE=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:uma-ticket" \
  -d "audience=$CLIENT_ID" \
  -d "permission=/api/test#GET")

echo "Authorization response:"
echo $AUTH_RESPONSE | jq .

# Test 4: Test with different resource
echo "4. Testing different resource..."
AUTH_RESPONSE2=$(curl -s -X POST "$KEYCLOAK_URL/realms/$REALM/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -H "Authorization: Bearer $USER_TOKEN" \
  -d "grant_type=urn:ietf:params:oauth:grant-type:uma-ticket" \
  -d "audience=$CLIENT_ID" \
  -d "permission=/admin/test#GET")

echo "Admin resource authorization response:"
echo $AUTH_RESPONSE2 | jq .

echo "=== Test completed ===" 