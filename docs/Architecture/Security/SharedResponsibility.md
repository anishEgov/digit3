## 🧭 Shared Responsibility Model for Security

Security in a citizen-centric public service platform spans multiple layers — from the infrastructure and platform to services, APIs, data, and user credentials. Ensuring end-to-end protection requires a **shared responsibility model** between the following teams:

| Layer                          | Primary Responsible Team     | Supporting Teams                      |
|-------------------------------|------------------------------|----------------------------------------|
| **Platform infrastructure**   | Platform/DevOps Team         | Cloud/Network Admins                  |
| **Authentication & IAM**      | Platform Security Team       | Application Developers                |
| **Secrets Management**        | Platform/DevOps Team         | Developers, Database Admins           |
| **API & Service Security**    | Application Developers       | Platform Team                         |
| **Credential Management (VCs)**| Platform Security Team       | App Developers, Wallet Team           |
| **Data Security & Encryption**| Database Admins              | Platform Team, Compliance             |
| **Network Security**          | Network/Cloud Team           | DevOps                                |
| **Monitoring & Audit Logs**   | DevOps / SRE Team            | Platform Team                         |
| **Compliance & Risk**         | Security & Legal Team        | All Engineering Functions             |

---

### Platform Team Responsibilities

- Set up and manage **OAuth 2.0 / OIDC provider** (e.g., Keycloak)
- Configure **default security policies** for tokens, scopes, roles
- Implement **secrets management system** (e.g., Vault) and grant scoped access
- Provision and manage **service mesh** (e.g., Istio) for secure internal communication
- Enforce **TLS**, **mTLS**, and firewall rules
- Centralize **logging and monitoring** (ELK, Loki, Prometheus)
- Manage **certificate infrastructure** (TLS and VC signing)
- Set up **VC registry** and **revocation infrastructure**
- Build APIs/tools for **self-service onboarding** of departments

---

### Application Developer Responsibilities

- Validate tokens (JWTs) and scopes for every API call
- Use short-lived tokens; avoid hardcoded credentials
- Integrate with centralized auth (Keycloak/OIDC)
- Secure inter-service calls using OAuth or mesh identities
- Store and present VCs using standard protocols (e.g., OID4VC)
- Apply OWASP best practices (input validation, XSS prevention)
- Log meaningful actions with traceability

---

### DevOps / SRE Responsibilities

- Secure infrastructure-as-code (e.g., Terraform, Helm)
- Restrict and audit access to secrets and deployments
- Automate key/token rotation and provisioning
- Monitor runtime environments for anomalies
- Enforce TLS termination, rate limiting, and resource quotas
- Patch and harden OS and container base images

---

### Database Administrator Responsibilities

- Encrypt sensitive data at rest (with column-level encryption if needed)
- Apply least-privilege database roles and credentials
- Enable access and query logging
- Use Vault or similar for secret injection into apps
- Ensure secure and tested backup processes

---

### Compliance / Governance Responsibilities

- Maintain data classification and privacy policies
- Define citizen data sharing and consent standards
- Monitor external access and revocation compliance
- Track alignment with regulations (e.g., DPDP, GDPR)
- Lead breach notification and incident response protocols