#!/bin/bash

# Kong Gateway Test Script
# This script contains all the essential CURL commands for testing Kong Gateway with Keycloak
# Prerequisites: Port forwards must be active for Kong and Keycloak

# Configuration
KONG_ADMIN_URL="http://localhost:8005"
KONG_PROXY_URL="http://localhost:8006"
KEYCLOAK_URL="http://localhost:8007/keycloak-test"
HOST_HEADER="digit-lts.digit.org"

echo "🚀 Kong Gateway Test Script"
echo "=========================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if services are running
check_services() {
    print_step "Checking if services are accessible..."
    
    # Check Kong Admin
    if curl -s "$KONG_ADMIN_URL/status" > /dev/null; then
        print_success "Kong Admin API is accessible"
    else
        print_error "Kong Admin API is not accessible. Check port forward: kubectl port-forward -n egov svc/kong-db-admin 8005:8001"
        exit 1
    fi
    
    # Check Kong Proxy
    if curl -s "$KONG_PROXY_URL" > /dev/null; then
        print_success "Kong Proxy is accessible"
    else
        print_error "Kong Proxy is not accessible. Check port forward: kubectl port-forward -n egov svc/kong-db-proxy 8006:80"
        exit 1
    fi
    
    # Check Keycloak
    if curl -s "$KEYCLOAK_URL/realms/master" > /dev/null; then
        print_success "Keycloak is accessible"
    else
        print_error "Keycloak is not accessible. Check port forward: kubectl port-forward -n keycloak svc/keycloak 8007:8080"
        exit 1
    fi
}

# Function to get Keycloak admin token
get_admin_token() {
    print_step "Getting Keycloak admin token..."
    
    ACCESS_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/master/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=digit&password=digit@321&grant_type=password&client_id=admin-cli" | jq -r '.access_token')
    
    if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
        print_success "Got admin token: ${ACCESS_TOKEN:0:50}..."
        export ACCESS_TOKEN
    else
        print_error "Failed to get admin token"
        exit 1
    fi
}

# Function to get user JWT token
get_user_token() {
    print_step "Getting user JWT token..."
    
    JWT_TOKEN=$(curl -s -X POST "$KEYCLOAK_URL/realms/DEMOUSERCAMUNDA/protocol/openid-connect/token" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=demousercamunda@gmail.com&password=password123&grant_type=password&client_id=kong-test-client" | jq -r '.access_token')
    
    if [ "$JWT_TOKEN" != "null" ] && [ -n "$JWT_TOKEN" ]; then
        print_success "Got user JWT token: ${JWT_TOKEN:0:50}..."
        export JWT_TOKEN
    else
        print_error "Failed to get user JWT token"
        exit 1
    fi
}

# Function to check Kong status
check_kong_status() {
    print_step "Checking Kong status..."
    
    curl -s "$KONG_ADMIN_URL/status" | jq '.'
}

# Function to list Kong plugins
list_kong_plugins() {
    print_step "Listing Kong plugins..."
    
    curl -s "$KONG_ADMIN_URL/plugins" | jq '.data[] | {name: .name, enabled: .enabled}'
}

# Function to list Kong services
list_kong_services() {
    print_step "Listing Kong services..."
    
    curl -s "$KONG_ADMIN_URL/services" | jq '.data[] | {name: .name, host: .host, port: .port}'
}

# Function to list Kong routes
list_kong_routes() {
    print_step "Listing Kong routes..."
    
    curl -s "$KONG_ADMIN_URL/routes" | jq '.data[] | {name: .name, paths: .paths, hosts: .hosts}'
}

# Function to test authentication (should fail)
test_no_auth() {
    print_step "Testing without authentication (should fail with 401)..."
    
    response=$(curl -s -w "HTTP_CODE:%{http_code}" \
        -H "Host: $HOST_HEADER" \
        "$KONG_PROXY_URL/localization/localization/v1/search?tenantId=pb&module=rainmaker-common&locale=en_IN")
    
    http_code=$(echo "$response" | sed -n 's/.*HTTP_CODE://p')
    
    if [ "$http_code" = "401" ]; then
        print_success "Authentication test passed - got 401 as expected"
    else
        print_error "Authentication test failed - expected 401, got $http_code"
    fi
}

# Function to test complete flow
test_complete_flow() {
    print_step "Testing complete flow with JWT authentication..."
    
    response=$(curl -s -w "HTTP_CODE:%{http_code}" \
        -X POST \
        -H "Host: $HOST_HEADER" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        "$KONG_PROXY_URL/localization/messages/v1/_upsert" \
        -d '{
            "RequestInfo": {
                "apiId": "emp",
                "ver": "1.0",
                "action": "create",
                "did": "1",
                "key": "abcdkey",
                "msgId": "20170310130900",
                "requesterId": "rajesh",
                "authToken": "0cfe07e1-94b5-4f50-a7a0-c7c186feb9d5",
                "userInfo": {
                    "id": 128
                }
            },
            "tenantId": "DEFAULT",
            "messages": [
                {
                    "code": "TestMessage",
                    "message": "Test Message from Script",
                    "module": "digit-ui",
                    "locale": "en_IN"
                }
            ]
        }')
    
    http_code=$(echo "$response" | sed -n 's/.*HTTP_CODE://p')
    response_body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    
    if [ "$http_code" = "200" ]; then
        print_success "Complete flow test passed - got 200"
        echo "Response: $response_body" | jq '.'
    else
        print_error "Complete flow test failed - got $http_code"
        echo "Response: $response_body"
    fi
}

# Function to test rate limiting
test_rate_limiting() {
    print_step "Testing rate limiting (making 3 requests to check headers)..."
    
    for i in {1..3}; do
        echo "Request $i:"
        curl -s -I \
            -X POST \
            -H "Host: $HOST_HEADER" \
            -H "Authorization: Bearer $JWT_TOKEN" \
            -H "Content-Type: application/json" \
            "$KONG_PROXY_URL/localization/messages/v1/_upsert" \
            -d '{"RequestInfo":{"apiId":"rate-test"},"tenantId":"DEFAULT","messages":[{"code":"RateTest","message":"Rate test","module":"test","locale":"en_IN"}]}' \
            | grep -E "(HTTP|X-RateLimit|RateLimit)"
        echo ""
    done
}

# Function to add rate limiting plugin
add_rate_limiting() {
    print_step "Adding rate limiting plugin..."
    
    curl -X POST "$KONG_ADMIN_URL/plugins" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "rate-limiting",
            "config": {
                "minute": 100,
                "hour": 1000,
                "policy": "local"
            }
        }' | jq '.'
}

# Function to add opentelemetry plugin
add_opentelemetry() {
    print_step "Adding OpenTelemetry plugin..."
    
    curl -X POST "$KONG_ADMIN_URL/plugins" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "opentelemetry",
            "config": {
                "endpoint": "http://jaeger-collector.observability.svc.cluster.local:14268/api/traces",
                "headers": {
                    "X-Source": "kong-gateway"
                },
                "resource_attributes": {
                    "service.name": "kong-gateway",
                    "service.version": "3.9.1"
                }
            }
        }' | jq '.'
}

# Function to decode JWT token
decode_jwt() {
    print_step "Decoding JWT token payload..."
    
    if [ -n "$JWT_TOKEN" ]; then
        echo "$JWT_TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq '.'
    else
        print_error "No JWT token available"
    fi
}

# Function to show help
show_help() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  check-services    - Check if Kong and Keycloak are accessible"
    echo "  get-tokens       - Get admin and user tokens"
    echo "  kong-status      - Check Kong status"
    echo "  list-plugins     - List all Kong plugins"
    echo "  list-services    - List all Kong services"
    echo "  list-routes      - List all Kong routes"
    echo "  test-no-auth     - Test without authentication (should fail)"
    echo "  test-complete    - Test complete flow with authentication"
    echo "  test-rate-limit  - Test rate limiting"
    echo "  add-rate-limit   - Add rate limiting plugin"
    echo "  add-opentelemetry - Add OpenTelemetry plugin"
    echo "  decode-jwt       - Decode JWT token payload"
    echo "  full-test        - Run complete test suite"
    echo "  help             - Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 full-test     - Run all tests"
    echo "  $0 test-complete - Test authenticated flow"
    echo "  $0 decode-jwt    - Decode current JWT token"
}

# Function to run full test suite
run_full_test() {
    print_step "Running full test suite..."
    echo ""
    
    check_services
    get_admin_token
    get_user_token
    check_kong_status
    list_kong_plugins
    list_kong_services
    list_kong_routes
    test_no_auth
    test_complete_flow
    test_rate_limiting
    decode_jwt
    
    print_success "Full test suite completed!"
}

# Main execution
case "${1:-help}" in
    "check-services")
        check_services
        ;;
    "get-tokens")
        get_admin_token
        get_user_token
        ;;
    "kong-status")
        check_kong_status
        ;;
    "list-plugins")
        list_kong_plugins
        ;;
    "list-services")
        list_kong_services
        ;;
    "list-routes")
        list_kong_routes
        ;;
    "test-no-auth")
        test_no_auth
        ;;
    "test-complete")
        get_admin_token
        get_user_token
        test_complete_flow
        ;;
    "test-rate-limit")
        get_admin_token
        get_user_token
        test_rate_limiting
        ;;
    "add-rate-limit")
        add_rate_limiting
        ;;
    "add-opentelemetry")
        add_opentelemetry
        ;;
    "decode-jwt")
        decode_jwt
        ;;
    "full-test")
        run_full_test
        ;;
    "help"|*)
        show_help
        ;;
esac 