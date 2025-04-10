# DIGIT Platform - Backend API Design Document

## Common Conventions
- **Auth**: All endpoints secured via JWT (issued by Identity Service).
- **Headers**:
  - `Authorization: Bearer <token>`
  - `X-Tenant-ID: <tenant-id>`
- **Format**: JSON over HTTPS
- **Audit Fields** (included in all read responses):
  - `created_by`, `created_on`, `modified_by`, `modified_on`
- **Error Schema**:
```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": "string"
  }
}
```

---

## 1. Account Service

### POST /v3/accounts
Create a new account with basic configuration
```json
{
  "name": "string",
  "domain": "string",
  "oidc_config": { "issuer": "string", "client_id": "string" },
  "administrator": "user_id"
}
```

### GET /v3/accounts/{id}
Retrieve account details by ID
```json
{
  "id": "string",
  "name": "string",
  "domain": "string",
  "status": "active|closed",
  "administrator": "user_id",
  "oidc_config": { "issuer": "string", "client_id": "string" },
  "created_by": "string",
  "created_on": "datetime",
  "modified_by": "string",
  "modified_on": "datetime"
}
```

### POST /v3/accounts/{id}/status
Update account status
```json
{ "status": "active|closed" }
```

### POST /v3/accounts/{id}/administrator
Update account administrator
```json
{ "administrator": "user_id" }
```

### POST /v3/accounts/{id}/users
Add a new user to the account
```json
{
  "name": "string",
  "email": "string",
  "phone": "string",
  "unique_id": "string",
  "roles": ["string"]
}
```

### GET /v3/accounts/{id}/users
List all users in the account
```json
[
  {
    "user_id": "string",
    "name": "string",
    "email": "string",
    "phone": "string",
    "unique_id": "string",
    "roles": ["string"],
    "created_by": "string",
    "created_on": "datetime",
    "modified_by": "string",
    "modified_on": "datetime"
  }
]
```

### POST /v3/accounts/{id}/roles
Create a new role in the account
```json
{ "name": "string", "permissions": ["string"] }
```

### GET /v3/accounts/{id}/roles
List all roles in the account
```json
[
  { "role_id": "string", "name": "string", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
]
```

---

## 2. Identity Service (OIDC)

### POST /v3/auth/token
Authenticate user and get JWT token
```json
{ "username": "string", "password": "string" }
```

### GET /v3/auth/userinfo
Get user information from JWT token
```json
{ "sub": "string", "name": "string", "email": "string" }
```

### POST /v3/auth/logout
Invalidate current JWT token
```json
{ "success": true }
```

### GET /v3/.well-known/openid-configuration
Get OIDC configuration details
```json
{
  "issuer": "string",
  "authorization_endpoint": "string",
  "token_endpoint": "string",
  "userinfo_endpoint": "string",
  "jwks_uri": "string",
  "response_types_supported": ["string"],
  "subject_types_supported": ["string"],
  "id_token_signing_alg_values_supported": ["string"]
}
```

---

## 3. Catalogue Service

### POST /v3/catalogue/services
Register a new service in the catalogue
```json
{ "name": "string", "description": "string", "endpoint": "string", "configuration": {} }
```

### GET /v3/catalogue/services
List all registered services
```json
[
  {
    "id": "string",
    "name": "string",
    "description": "string",
    "endpoint": "string",
    "configuration": {},
    "created_by": "string",
    "created_on": "datetime",
    "modified_by": "string",
    "modified_on": "datetime"
  }
]
```

### GET /v3/catalogue/services/{id}
Get service details by ID
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "endpoint": "string",
  "configuration": {},
  "created_by": "string",
  "created_on": "datetime",
  "modified_by": "string",
  "modified_on": "datetime"
}
```

### PUT /v3/catalogue/services/{id}
Update service details
```json
{ "description": "string", "endpoint": "string", "configuration": {} }
```

---

## 4. Registration Service

### POST /v3/requests
Create a new registration request
```json
{ "service_id": "string", "data": {} }
```

### GET /v3/requests/{id}
Get registration request details
```json
{
  "id": "string",
  "status": "string",
  "data": {},
  "created_by": "string",
  "created_on": "datetime",
  "modified_by": "string",
  "modified_on": "datetime"
}
```

### PUT /v3/requests/{id}
Update registration request data
```json
{ "data": {} }
```

### GET /v3/requests?serviceId=&status=&userId=
Search registration requests with filters
```json
[
  {
    "id": "string",
    "status": "string",
    "created_by": "string",
    "created_on": "datetime",
    "modified_by": "string",
    "modified_on": "datetime"
  }
]
```

---

## 5. Registry Service

### POST /v3/registry/schemas
Register a new data schema
```json
{ "name": "string", "schema": {} }
```

### GET /v3/registry/schemas/{name}
Get schema details by name
```json
{ "name": "string", "schema": {}, "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### POST /v3/registry/data/{schema}
Create new data record for a schema
```json
{ "data": {} }
```

### GET /v3/registry/data/{schema}/{id}
Get data record by ID
```json
{ "id": "string", "data": {}, "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### PUT /v3/registry/data/{schema}/{id}
Update data record
```json
{ "data": {} }
```

### DELETE /v3/registry/data/{schema}/{id}
Delete data record
```json
{ "deleted": true }
```

### GET /v3/registry/data/{schema}?filters
Search data records with filters
```json
[
  { "id": "string", "data": {}, "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
]
```

---

## 6. Workflow Service

### POST /v3/workflows
Create a new workflow definition
```json
{ "name": "string", "states": [], "transitions": [] }
```

### GET /v3/workflows/{id}
Get workflow definition details
```json
{ "id": "string", "name": "string", "states": [], "transitions": [], "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### POST /v3/workflows/{id}/instances
Create a new workflow instance
```json
{ "entity_id": "string" }
```

### POST /v3/workflow-instances/{id}/transition
Transition workflow instance to next state
```json
{ "action": "string" }
```

### GET /v3/workflow-instances/{id}
Get workflow instance details
```json
{ "id": "string", "state": "string", "history": [], "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

---

## 7. Notification Service

### POST /v3/notifications
Send a notification using template
```json
{ "template_id": "string", "recipient": "string", "data": {} }
```

### GET /v3/notifications/{id}
Get notification status
```json
{ "id": "string", "status": "string", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### POST /v3/templates
Create a new notification template
```json
{ "name": "string", "content": "string", "type": "email|sms|inapp" }
```

### GET /v3/templates
List all notification templates
```json
[
  { "id": "string", "name": "string", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
]
```

---

## 8. File Service

### POST /v3/files
Upload a new file
```json
{ "id": "string" }
```

### GET /v3/files/{id}
Get file metadata
```json
{ "id": "string", "filename": "string", "size": "number", "created_by": "string", "created_on": "datetime", "modified_by": "string", "modified_on": "datetime" }
```

### GET /v3/files/{id}/url
Get file download URL
```json
{ "url": "string" }
```

### DELETE /v3/files/{id}
Delete a file
```json
{ "deleted": true }
```

---

## 9. Certificate Service

### POST /v3/certificates
Issue a new certificate
```json
{
  "subject_id": "string",
  "certificate_type": "string",
  "data": { "any": "json" }
}
```

### GET /v3/certificates/{id}
Get certificate details
```json
{
  "id": "string",
  "certificate_type": "string",
  "subject_id": "string",
  "account_id": "string",
  "status": "issued",
  "issued_at": "datetime",
  "revoked_at": "datetime|null",
  "data": { "any": "json" }
}
```

### POST /v3/certificates/{id}/verify
Verify certificate validity
```json
{
  "id": "string",
  "valid": true,
  "reason": "string|null"
}
```

### POST /v3/certificates/{id}/revoke
Revoke a certificate
```json
{
  "id": "string",
  "status": "revoked",
  "revoked_at": "datetime"
}
```
