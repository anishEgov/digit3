#!/usr/bin/env python3
import requests
import json
import sys
from urllib.parse import urljoin

# Kong Admin API configuration
KONG_ADMIN_URL = "http://localhost:8097"
KEYCLOAK_BASE_URL = "http://keycloak:8080/keycloak-test"

# Service configurations - mapping service names to their backend URLs based on docker-compose
SERVICES = {
    "workflow": "http://workflow-service:8080",     # workflow-service container, internal port 8080
    "localization": "http://localisation:8088",    # localisation container, internal port 8088
    "notification": "http://notification:8080",    # notification container, internal port 8080
    "template-config": "http://template-config:8081", # template-config container, internal port 8081
    "boundary": "http://boundary-service:8081",     # boundary-service container, internal port 8081
    "tenant-management": "http://tenant-management:8081", # tenant-management container, internal port 8081
    "urlshortener": "http://urlshortener:8088",     # urlshortener container, internal port 8088
    "idgen": "http://idgen:8088",                   # idgen container, internal port 8088
    
    # For services not in docker-compose, using placeholder URLs
    "account": "http://localhost:8080",            # placeholder - add actual service
    "template": "http://localhost:8081",           # placeholder - add actual service  
    "generate": "http://localhost:8082",           # placeholder - add actual service
    "mdms-v2": "http://localhost:8083",            # placeholder - add actual service
    "filestore": "http://localhost:8084",          # placeholder - add actual service
    "shapefile": "http://boundary-service:8081"    # using boundary service container
}

# Route configurations - each route maps to a service
ROUTES = [
    # Account service routes
    {"name": "account-route", "paths": ["/account"], "service": "account"},
    {"name": "account-config-route", "paths": ["/account/config"], "service": "account"},
    
    # Template service routes
    {"name": "template-route", "paths": ["/template"], "service": "template"},
    {"name": "generate-route", "paths": ["/generate"], "service": "generate"},
    
    # Template config service routes
    {"name": "template-config-v1-route", "paths": ["/template-config/v1/config"], "service": "template-config"},
    {"name": "template-config-render-route", "paths": ["/template-config/v1/render"], "service": "template-config"},
    
    # Notification service routes
    {"name": "notification-template-route", "paths": ["/notification/template"], "service": "notification"},
    {"name": "notification-template-preview-route", "paths": ["/notification/template/preview"], "service": "notification"},
    {"name": "notification-email-route", "paths": ["/notification/email/send"], "service": "notification"},
    {"name": "notification-sms-route", "paths": ["/notification/sms/send"], "service": "notification"},
    
    # MDMS v2 service routes
    {"name": "mdms-v2-route", "paths": ["/mdms-v2/v2"], "service": "mdms-v2"},
    {"name": "mdms-v1-route", "paths": ["/mdms-v2/v1/mdms"], "service": "mdms-v2"},
    {"name": "mdms-schema-route", "paths": ["/mdms-v2/v1/schema"], "service": "mdms-v2"},
    

    
    # Workflow service routes
    {"name": "workflow-process-route", "paths": ["/workflow/v3/process"], "service": "workflow"},
    {"name": "workflow-process-definition-route", "paths": ["/workflow/v3/process/definition"], "service": "workflow"},
    {"name": "workflow-process-state-route", "paths": ["/workflow/v3/process/state"], "service": "workflow"},
    {"name": "workflow-state-route", "paths": ["/workflow/v3/state"], "service": "workflow"},
    {"name": "workflow-action-route", "paths": ["/workflow/v3/action"], "service": "workflow"},
    {"name": "workflow-transition-route", "paths": ["/workflow/v3/transition"], "service": "workflow"},
    
    # Filestore service routes
    {"name": "filestore-upload-route", "paths": ["/filestore/v1/files/upload"], "service": "filestore"},
    {"name": "filestore-download-route", "paths": ["/filestore/v1/files/download-urls"], "service": "filestore"},
    {"name": "filestore-upload-url-route", "paths": ["/filestore/v1/files/upload-url"], "service": "filestore"},
    {"name": "filestore-confirm-route", "paths": ["/filestore/v1/files/confirm-upload"], "service": "filestore"},
    {"name": "filestore-tag-route", "paths": ["/filestore/v1/files/tag"], "service": "filestore"},
    {"name": "filestore-metadata-route", "paths": ["/filestore/v1/files/metadata"], "service": "filestore"},
    {"name": "filestore-categories-route", "paths": ["/filestore/v1/files/document-categories"], "service": "filestore"},
    
    # Localization service routes
    {"name": "localization-messages-route", "paths": ["/localization/messages"], "service": "localization"},
    {"name": "localization-missing-route", "paths": ["/localization/messages/_missing"], "service": "localization"},
    {"name": "localization-upsert-route", "paths": ["/localization/messages/_upsert"], "service": "localization"},
    {"name": "localization-cache-route", "paths": ["/localization/cache/_bust"], "service": "localization"},
    
    # Boundary service routes
    {"name": "boundary-route", "paths": ["/boundary"], "service": "boundary"},
    {"name": "boundary-hierarchy-route", "paths": ["/boundary-hierarchy-definition"], "service": "boundary"},
    {"name": "boundary-relationships-route", "paths": ["/boundary-relationships"], "service": "boundary"},
    {"name": "shapefile-boundary-route", "paths": ["/shapefile/boundary/create"], "service": "boundary"},
    
    # URL Shortener service routes
    {"name": "shortener-route", "paths": ["/shortener"], "service": "urlshortener"},
    
    # ID Generation service routes  
    {"name": "idgen-route", "paths": ["/idgen"], "service": "idgen"},
    
    # Tenant Management service routes
    {"name": "tenant-management-route", "paths": ["/tenant"], "service": "tenant-management"}
]

def make_request(method, url, data=None, headers=None):
    """Make HTTP request and handle errors"""
    try:
        if headers is None:
            headers = {"Content-Type": "application/json"}
        
        response = requests.request(method, url, json=data, headers=headers, timeout=10)
        
        if response.status_code in [200, 201, 409]:  # 409 = already exists
            return response.json() if response.text else {}
        else:
            print(f"Error {response.status_code}: {response.text}")
            return None
    except Exception as e:
        print(f"Request failed: {e}")
        return None

def create_services():
    """Create all Kong services"""
    print("Creating Kong services...")
    
    for service_name, service_url in SERVICES.items():
        print(f"Creating service: {service_name} -> {service_url}")
        
        service_data = {
            "name": service_name,
            "url": service_url
        }
        
        # Use PUT to create/update service with specific name
        url = urljoin(KONG_ADMIN_URL, f"/services/{service_name}")
        result = make_request("PUT", url, service_data)
        
        if result:
            print(f"✓ Service {service_name} created/updated")
        else:
            print(f"✗ Failed to create service {service_name}")

def create_routes():
    """Create all Kong routes"""
    print("\nCreating Kong routes...")
    
    for route_config in ROUTES:
        route_name = route_config["name"]
        print(f"Creating route: {route_name}")
        
        route_data = {
            "name": route_name,
            "paths": route_config["paths"],
            "strip_path": False,
            "service": {"name": route_config["service"]}
        }
        
        # Use PUT to create/update route with specific name
        url = urljoin(KONG_ADMIN_URL, f"/routes/{route_name}")
        result = make_request("PUT", url, route_data)
        
        if result:
            print(f"✓ Route {route_name} created/updated")
        else:
            print(f"✗ Failed to create route {route_name}")

def apply_plugins_to_routes():
    """Apply dynamic-jwt and keycloak-rbac plugins to all routes"""
    print("\nApplying plugins to all routes...")
    
    # Get all routes
    routes_url = urljoin(KONG_ADMIN_URL, "/routes")
    response = make_request("GET", routes_url)
    
    if not response or 'data' not in response:
        print("Failed to get routes")
        return
    
    for route in response['data']:
        route_id = route['id']
        route_name = route['name']
        
        print(f"Applying plugins to route: {route_name}")
        
        # Apply dynamic-jwt plugin
        jwt_plugin_data = {
            "name": "dynamic-jwt",
            "config": {
                "keycloak_base_url": KEYCLOAK_BASE_URL
            }
        }
        
        jwt_url = urljoin(KONG_ADMIN_URL, f"/routes/{route_id}/plugins")
        jwt_result = make_request("POST", jwt_url, jwt_plugin_data)
        
        if jwt_result:
            print(f"  ✓ Dynamic JWT plugin applied to {route_name}")
        else:
            print(f"  ✗ Failed to apply JWT plugin to {route_name}")
        
        # Apply keycloak-rbac plugin
        rbac_plugin_data = {
            "name": "keycloak-rbac",
            "config": {
                "keycloak_base_url": KEYCLOAK_BASE_URL,
                "client_id": "auth-server"
            }
        }
        
        rbac_url = urljoin(KONG_ADMIN_URL, f"/routes/{route_id}/plugins")
        rbac_result = make_request("POST", rbac_url, rbac_plugin_data)
        
        if rbac_result:
            print(f"  ✓ Keycloak RBAC plugin applied to {route_name}")
        else:
            print(f"  ✗ Failed to apply RBAC plugin to {route_name}")
        
        # Apply header-enrichment plugin
        header_plugin_data = {
            "name": "header-enrichment",
            "config": {}  # Using default configuration
        }
        
        header_url = urljoin(KONG_ADMIN_URL, f"/routes/{route_id}/plugins")
        header_result = make_request("POST", header_url, header_plugin_data)
        
        if header_result:
            print(f"  ✓ Header Enrichment plugin applied to {route_name}")
        else:
            print(f"  ✗ Failed to apply Header Enrichment plugin to {route_name}")

def main():
    """Main function to set up everything"""
    print("Kong Route Setup Script")
    print("=" * 50)
    
    # Test Kong connectivity
    try:
        response = requests.get(urljoin(KONG_ADMIN_URL, "/"), timeout=5)
        if response.status_code != 200:
            print(f"Kong Admin API not accessible at {KONG_ADMIN_URL}")
            sys.exit(1)
        print(f"✓ Kong Admin API accessible at {KONG_ADMIN_URL}")
    except Exception as e:
        print(f"Cannot connect to Kong: {e}")
        sys.exit(1)
    
    # Create services
    create_services()
    
    # Create routes
    create_routes()
    
    # Apply plugins
    apply_plugins_to_routes()
    
    print("\n" + "=" * 50)
    print("Kong setup completed!")
    print(f"All routes configured with:")
    print(f"  - dynamic-jwt plugin (Keycloak: {KEYCLOAK_BASE_URL})")
    print(f"  - keycloak-rbac plugin (client_id: auth-server)")
    print(f"  - header-enrichment plugin (default config)")

if __name__ == "__main__":
    main() 