# DIGIT Platform Requirements

## Overview

DIGIT is being built as modular, multi-tenant digital public infrastructure for public service deliver. It provides pluggable identity, account and service management capabilities through a set of backend services (built in Go) and frontend UIs (built in Flutter). It is managed using Docker Compose and consists of following:

- **Frontend Applicaitons**
    - **DIGIT Console**: Account(or Tenant) administration user interface.
    - **DIGIT Studio**: A low code no code Service Design and Management portal for service providers.
    - **DIGIT Citizen**: Unified Interface for Citizens to discover and engage with services.
    - **DIGIT Employee**: Unified Interface for Employees to track and fulfill service requests
    - **DIGIT Administrator**: Unified Interface for Administrators to monitor and plan for Services

- **Backend Services**
    - **Account**: Provides APIs for registration and management of accounts and users of the service provider.
    - **Identity**: Provides OIDC endpoints to authenticate users.
    - **Catalogue**: Enables service providers to register and manage services and enables discovery of services by service consumers.
    - **Registration**: Manages registration and service requests by the service consumers. These could be citizens or other service providers.
    - **Registry**: Manages registry schema and data about the registry.
    - **Workflow**: Manages workflow schema and workflow instances. 
    - **Notification**: Manages notification configuration and notification requests like email, sms, apps etc.
    - **File**: Manages files and provides secure short urls to files.


## Frontend Applications

### 1. DIGIT Console

**Purpose**: Admin interface for account (tenant) setup and management.

**High-Level Requirements**:
- Tenant onboarding and lifecycle management
- Manage OIDC configuration for the tenant (Google, Keycloak, etc.)
- Manage tenant-specific configuration (logos, themes, settings)
- User and role management for tenant administrators
- Service enablement (select which backend services are active for the tenant)
- Show usage statistics and logs per tenant

---

### 2. DIGIT Studio

**Purpose**: Low-code/no-code environment to design and manage services.

**High-Level Requirements**:
- Service designer (form, workflow, rules engine)
- Define service schemas (uses Registry Service)
- Link services to workflows (uses Workflow Service)
- Preview and test service flows
- Version control and publishing services
- Enable/disable services for specific user roles
- Role-based access control for designers

---

### 3. DIGIT Citizen

**Purpose**: Unified portal for citizens to access and request services.

**High-Level Requirements**:
- Discover available services (uses Catalogue Service)
- View and initiate service requests (uses Registration Service)
- Track request status (uses Workflow Service)
- Receive notifications (uses Notification Service)
- Store profile and linked service IDs (e.g., electricity, water)
- Authenticate with OIDC (platform or tenant-level)
- Multilingual and responsive design

---

### 4. DIGIT Employee

**Purpose**: Operational interface for employees managing service requests.

**High-Level Requirements**:
- View assigned service requests (uses Workflow Service)
- Act on service requests (approve, reject, add comments)
- Role-based dashboards (inspector, verifier, supervisor, etc.)
- Search and filter service requests
- View citizen-submitted documents and data (uses File and Registry Services)
- Internal chat or comments on requests
- Notification inbox (internal memos, alerts)

---

### 5. DIGIT Administrator

**Purpose**: Monitoring and analytics dashboard for administrators.

**High-Level Requirements**:
- View request volume and turnaround time across services
- Monitor service performance and bottlenecks
- User activity audit logs
- System health indicators (per backend service)
- Manage workflows and routing rules
- Export and schedule reports
- Set escalation rules or auto-reminders

---

## Backend Services (Go)

### 1. Account Service

**Purpose**: Tenant (account) and user management.

**High-Level Requirements**:
- Create and manage tenants
- Store tenant configurations (themes, logos, OIDC config)
- User registration, role assignment, and authentication link to OIDC
- Link users to tenants and roles
- Provide user and role APIs to frontends
- Secure REST APIs with JWT

---

### 2. Identity Service

**Purpose**: Platform-level OIDC authentication.

**High-Level Requirements**:
- OIDC-compliant identity provider OR integrate with external OIDC (e.g., Firebase, Keycloak)
- Provide access and refresh tokens
- Support multi-tenant login with tenant-specific provider config
- Support user info and introspection endpoints
- Logout and session management APIs

---

### 3. Catalogue Service

**Purpose**: Service discovery and metadata.

**High-Level Requirements**:
- Register and update service metadata (name, description, version, endpoints)
- Tag services with categories, keywords, roles
- Mark services as public, restricted, or hidden
- Search and filter services for consumers (citizens/employees)
- Track usage metrics per service

---

### 4. Registration Service

**Purpose**: Handles service request initiation.

**High-Level Requirements**:
- Accept service request submissions (linked to catalogue entry)
- Validate inputs based on schema (uses Registry Service)
- Generate unique service request IDs
- Initiate workflow (uses Workflow Service)
- Maintain state and status of requests
- Support viewing, updating, and cancelling requests

---

### 5. Registry Service

**Purpose**: Schema-based, tenant-aware data store.

**High-Level Requirements**:
- Define and store JSON schema definitions per tenant
- Create collections (e.g., property, water-connection, grievance)
- CRUD APIs for registry data (with schema enforcement)
- Role-based access control to registry entries
- Audit logs and history for each entry
- Ensure data isolation by X-Tenant-ID

**Sub-services**:
- Database Service: Manages schema definitions
- Data Service: Manages CRUD operations on data based on schema

---

### 6. Workflow Service

**Purpose**: State machine and rules engine for requests.

**High-Level Requirements**:
- Define workflows per service (states, transitions, roles, rules)
- Initiate workflow instances per request
- Track workflow history and current state
- Trigger role-based notifications or actions
- Support SLAs and escalation rules
- Link workflows to UI actions and buttons

---

### 7. Notification Service

**Purpose**: Multi-channel messaging engine.

**High-Level Requirements**:
- Configure notification templates (SMS, email, in-app)
- Send notification events triggered by services or workflows
- Queue and retry failed deliveries
- Channel integrations (SMTP, SMS gateway, push)
- Log delivery status per notification
- Multilingual support for messages

---

### 8. File Service

**Purpose**: Secure file storage and short-lived URL generation.

**High-Level Requirements**:
- Upload and store files per request (PDFs, images, etc.)
- Tag files with metadata (service, user, status)
- Generate expiring short URLs for access
- Validate file types and size limits
- Encrypt files at rest and in transit
- Link files to service request records
