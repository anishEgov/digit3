#!/bin/bash

# Keycloak setup script for testing RBAC integration
KEYCLOAK_URL="http://localhost:8080"
ADMIN_USER="admin"
ADMIN_PASS="admin123"
REALM_NAME="test-realm"
CLIENT_ID="kong-gateway"

echo "=== Setting up Keycloak for Kong RBAC Testing ==="

# Wait for Keycloak to be ready
echo "Waiting for Keycloak to be ready..."
until curl -s "$KEYCLOAK_URL/health/ready" > /dev/null; do
  echo "Waiting for Keycloak..."
  sleep 5
done
echo "Keycloak is ready!"

# Get admin token
echo "Getting admin access token..."
ADMIN_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=$ADMIN_USER" \
  -d "password=$ADMIN_PASS" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | jq -r '.access_token')

if [ "$ADMIN_TOKEN" = "null" ] || [ -z "$ADMIN_TOKEN" ]; then
  echo "Failed to get admin token"
  exit 1
fi
echo "Got admin token!"

# Create realm
echo "Creating realm: $REALM_NAME"
curl -s -X POST "$KEYCLOAK_URL/admin/realms" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "realm": "'$REALM_NAME'",
    "enabled": true,
    "displayName": "Test Realm for Kong RBAC"
  }'

# Create client with authorization enabled
echo "Creating client: $CLIENT_ID"
CLIENT_RESPONSE=$(curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "clientId": "'$CLIENT_ID'",
    "enabled": true,
    "clientAuthenticatorType": "client-secret",
    "secret": "kong-secret",
    "serviceAccountsEnabled": true,
    "authorizationServicesEnabled": true,
    "standardFlowEnabled": true,
    "directAccessGrantsEnabled": true,
    "publicClient": false,
    "protocol": "openid-connect"
  }')

# Get client UUID
CLIENT_UUID=$(curl -s -X GET "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients?clientId=$CLIENT_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq -r '.[0].id')

echo "Client UUID: $CLIENT_UUID"

# Create roles
echo "Creating roles..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "admin", "description": "Administrator role"}'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "user", "description": "Regular user role"}'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/roles" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "citizen", "description": "Citizen role"}'

# Create test users
echo "Creating test users..."
# Admin user
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testadmin",
    "enabled": true,
    "credentials": [{
      "type": "password",
      "value": "admin123",
      "temporary": false
    }],
    "realmRoles": ["admin"]
  }'

# Regular user
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "enabled": true,
    "credentials": [{
      "type": "password",
      "value": "user123",
      "temporary": false
    }],
    "realmRoles": ["user"]
  }'

# Citizen user
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testcitizen",
    "enabled": true,
    "credentials": [{
      "type": "password",
      "value": "citizen123",
      "temporary": false
    }],
    "realmRoles": ["citizen"]
  }'

# Create authorization resources
echo "Creating authorization resources..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$CLIENT_UUID/authz/resource-server/resource" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "localization-api",
    "type": "urn:kong:resources:api",
    "uris": ["/localization/*"],
    "scopes": [
      {"name": "GET"},
      {"name": "POST"},
      {"name": "PUT"},
      {"name": "DELETE"}
    ]
  }'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$CLIENT_UUID/authz/resource-server/resource" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "admin-api",
    "type": "urn:kong:resources:api",
    "uris": ["/admin/*"],
    "scopes": [
      {"name": "GET"},
      {"name": "POST"},
      {"name": "PUT"},
      {"name": "DELETE"}
    ]
  }'

# Create authorization policies
echo "Creating authorization policies..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$CLIENT_UUID/authz/resource-server/policy/role" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin Policy",
    "roles": [{"id": "admin", "required": true}],
    "logic": "POSITIVE"
  }'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$CLIENT_UUID/authz/resource-server/policy/role" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "User Policy", 
    "roles": [{"id": "user", "required": false}, {"id": "citizen", "required": false}],
    "logic": "POSITIVE"
  }'

# Create permissions
echo "Creating permissions..."
curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$CLIENT_UUID/authz/resource-server/permission/resource" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Localization Access Permission",
    "resources": ["localization-api"],
    "policies": ["User Policy"],
    "decisionStrategy": "UNANIMOUS"
  }'

curl -s -X POST "$KEYCLOAK_URL/admin/realms/$REALM_NAME/clients/$CLIENT_UUID/authz/resource-server/permission/resource" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin Access Permission",
    "resources": ["admin-api"],
    "policies": ["Admin Policy"],
    "decisionStrategy": "UNANIMOUS"
  }'

echo "=== Keycloak setup completed! ==="
echo "Realm: $REALM_NAME"
echo "Client: $CLIENT_ID"
echo "Users: testadmin/admin123, testuser/user123, testcitizen/citizen123"
echo "Admin URL: $KEYCLOAK_URL/admin"
echo "Test token endpoint: $KEYCLOAK_URL/realms/$REALM_NAME/protocol/openid-connect/token" 