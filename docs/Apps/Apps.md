# DIGIT Frontend Applications

## Overview

DIGIT provides a suite of frontend applications built in Flutter, designed to serve different user roles and needs:

- **DIGIT Console**: Account(or Tenant) administration user interface.
- **DIGIT Studio**: A low code no code Service Design and Management portal for service providers.
- **DIGIT Citizen**: Unified Interface for Citizens to discover and engage with services.
- **DIGIT Workbench**: Unified Interface for Employees\Vendors to track and fulfill service requests
- **DIGIT Dashboard**: Unified Interface for Administrators to monitor and plan for Services

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

### 4. DIGIT Workbench

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

### 5. DIGIT Dashboard

**Purpose**: Monitoring and analytics dashboard for administrators.

**High-Level Requirements**:
- View request volume and turnaround time across services
- Monitor service performance and bottlenecks
- User activity audit logs
- System health indicators (per backend service)
- Manage workflows and routing rules
- Export and schedule reports
- Set escalation rules or auto-reminders
