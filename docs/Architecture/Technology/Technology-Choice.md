# Technology Choice

This document consolidates technology patterns and choices from various [architectural principles](../../Practices%20and%20Patterns/Principles.md). Each section provides a list of technology patterns and choices that can be used to implement the principles.

## Table of Contents

- [Security & Privacy](#security--privacy)
- [Single Source of Truth](#single-source-of-truth)
- [Reliable and Cost-Effective](#reliable-and-cost-effective)
- [Scalable and Performant](#scalable-and-performant)
- [User-Centered and Inclusive](#user-centered-and-inclusive)
- [Observable and Transparent](#observable-and-transparent)
- [Interoperable](#interoperable)
- [Open Source](#open-source)
- [Modular and Evolvable](#modular-and-evolvable)
- [Configuration-First and Extensible](#configuration-first-and-extensible)
- [Intelligent](#intelligent)

## Security & Privacy

*Source: [Secure and Privacy-Protective](../Practices%20and%20Patterns/Secure-and-Privacy-Protective.md)*

### Technology Patterns
- **Centralized IAM**  
  OAuth 2.0 + OIDC via Keycloak for SSO and token issuance
- **Service-specific scopes**  
  Department-defined fine-grained permissions
- **Decentralized VC issuance**  
  Issue & verify credentials using OID4VC and DID
- **Centralized vault**  
  Secure storage of client secrets, signing keys, tokens
- **Centralized logging**  
  Track logins, API calls, VC issuance and sharing
- **VC status registry**  
  Implement VC revocation using statusList2021 or registry API
- **Gateway + Mesh**  
  API Gateway (external) + Service Mesh (internal)

### Technology Choices
- **Authentication**: OAuth 2.0, OIDC (Keycloak)
- **Credential Issuance**: OID4VC, DID, VC
- **Secrets Management**: HashiCorp Vault
- **Audit Logging**: ELK Stack, Loki
- **WAF & API Protection**: Cloudflare, Kong Gateway, AWS WAF, Nginx
- **Service Mesh**: Istio, Linkerd
- **Revocation Registry**: statusList2021

## Single Source of Truth

*Source: [Single Source of Truth](../Practices%20and%20Patterns/Single-Source-of-Truth.md)*

### Technology Patterns
- **Master Data Management (MDM)**  
  Maintain consistent and accurate master data across the organization.
- **Event-Driven Architecture (EDA)**  
  Components communicate through events, promoting loose coupling.
- **Event Sourcing**  
  Store state changes as a sequence of events.
- **Command Query Responsibility Segregation (CQRS)**  
  Separate read and write operations.
- **Data Lineage Tracking**  
  Trace data flow from origin to consumption.
- **Data Governance Frameworks**  
  Define data ownership, quality standards, and access controls.

### Technology Choices
- **Databases**: PostgreSQL, MongoDB
- **Event Streaming**: Apache Kafka, RabbitMQ
- **MDM Tools**: Pimcore, AtroCore
- **Schema Management**: Apache Avro, JSON Schema, Protocol Buffers
- **Data Lineage**: Apache Atlas, OpenLineage, DataHub, OpenMetadata
- **Integration**: Apache NiFi

## Reliable and Cost-Effective

*Source: [Reliable and Cost Effective](../Practices%20and%20Patterns/Reliable-and-Cost-Effective.md)*

### Technology Patterns
- **Fault Tolerance**  
  Continue operating properly during component failures.
- **Circuit Breakers**  
  Stop request flow when services are failing.
- **Observability and Monitoring**  
  Gain insights into system performance.
- **Autoscaling**  
  Adjust resources based on demand.
- **Load Shedding**  
  Gracefully degrade service during overload.
- **Health Checks**  
  Verify service availability.

### Technology Choices
- **Resilience Libraries**: Resilience4j, Failsafe
- **Monitoring**: Prometheus, Grafana, Zabbix
- **Tracing**: OpenTelemetry, Jaeger, Zipkin
- **Autoscaling**: Kubernetes, KEDA
- **Load Shedding**: Envoy, NGINX
- **Health Checks**: Consul, Nagios

## Scalable and Performant

*Source: [Scalable and Performant](../Practices%20and%20Patterns/Scalable-and-Performant.md)*

### Technology Patterns
- **Asynchronous Processing**  
  Handle operations asynchronously for improved throughput.
- **Message Queues**  
  Decouple services and manage communication efficiently.
- **Load Balancing**  
  Distribute network traffic across multiple servers.
- **Caching**  
  Implement at various levels to reduce latency.
- **Circuit Breaker Pattern**  
  Prevent cascading failures.
- **Bulkhead Pattern**  
  Isolate different parts of the system.
- **Auto-Scaling**  
  Automatically adjust computing resources.

### Technology Choices
- **Messaging**: Apache Kafka, RabbitMQ
- **Caching**: Redis, Infinispan
- **Load Balancer**: NGINX, HAProxy, Traefik, Apache Traffic Server
- **Auto-Scaling**: Kubernetes, Docker Swarm
- **Performance Testing**: Apache JMeter, Gatling

## User-Centered and Inclusive

*Source: [User-Centered and Inclusive](../Practices%20and%20Patterns/User-Centered-and-Inclusive.md)*

### Technology Patterns
- **Multi-Modal Interfaces**  
  Support various interaction modes including voice, text, and visual.
- **Accessibility Standards**  
  Implement WCAG 2.1 guidelines for inclusive design.
- **Localization and Internationalization**  
  Support multiple languages and regional formats.
- **Progressive Enhancement**  
  Ensure basic functionality works without advanced features.

### Technology Choices
- **UI Frameworks**: React, Angular, Vue.js
- **Accessibility Tools**: axe-core, WAVE
- **Localization**: i18next, react-i18next
- **Voice Interfaces**: Web Speech API, Speech Recognition API

## Observable and Transparent

*Source: [Observable and Transparent](../Practices%20and%20Patterns/Observable-and-Transparent.md)*

### Technology Patterns
- **Business Metrics Instrumentation**  
  Capture and analyze business-related data points.
- **Citizen Feedback Loops**  
  Collect and respond to citizen input.
- **Outcome-Based Reporting**  
  Focus on service outcomes and impacts.

### Technology Choices
- **Monitoring**: Grafana, Metabase
- **Data Collection**: Apache Superset, Redash
- **Feedback Tools**: LimeSurvey, Formspree

## Interoperable

*Source: [Interoperable](../Practices%20and%20Patterns/Interoperable.md)*

### Technology Patterns
- **RESTful APIs**  
  Create stateless, scalable web services.
- **GraphQL**  
  Enable flexible data retrieval.
- **OpenAPI / Swagger Definitions**  
  Describe RESTful APIs in machine-readable format.
- **gRPC with Protocol Buffers**  
  High-performance, language-agnostic RPCs.
- **Schema Registry**  
  Centralized repository for data schemas.
- **API Gateway**  
  Handle request routing and protocol translation.

### Technology Choices
- **API Design**: OpenAPI Specification, Swagger UI, Swagger Editor, Redoc
- **Data Exchange**: JSON, Protocol Buffers, Apache Avro
- **RPC Frameworks**: gRPC
- **API Gateways**: Kong, Tyk, KrakenD
- **Schema Registries**: Confluent Schema Registry, Apicurio Registry

## Open Source

*Source: [Open Source](../Practices%20and%20Patterns/Open-Source.md)*

### Technology Patterns
- **Distributed Collaboration and Governance**  
  Meritocratic or liberal contribution models.
- **Public CI/CD Workflows**  
  Transparent integration and deployment.
- **InnerSource Practices**  
  Open-source methodologies within organizations.
- **Modular Architecture**  
  Independent component development.

### Technology Choices
- **Version Control**: GitHub, GitLab, Gitea
- **Documentation**: MkDocs, Docusaurus, Read the Docs
- **CI/CD**: GitHub Actions, GitLab CI/CD, Jenkins, CircleCI, Travis CI, GoCD

## Modular and Evolvable

*Source: [Modular and Evolvable](../Practices%20and%20Patterns/Modular-and-Evolvable.md)*

### Technology Patterns
- **Microservices architecture**  
  Decompose applications into loosely coupled services that can be developed, deployed, and scaled independently.
- **API Gateway for routing, aggregation, and versioning**  
  Centralize API management to handle request routing, protocol translation, and version control.
- **Service Mesh for inter-service communication and observability**  
  Implement a dedicated infrastructure layer to manage service-to-service communication.
- **Event-driven architecture**  
  Design systems that respond to events, promoting decoupling and scalability.
- **Plugin architecture**  
  Allow the addition of new functionalities through plugins without modifying the core system.
- **Circuit breaker pattern**  
  Prevent cascading failures in distributed systems.
- **Strangler pattern**  
  Incrementally refactor a monolithic system by replacing specific pieces with new services.

### Technology Choices
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **Service Mesh**: Istio, Linkerd
- **API Gateway**: Kong, Ambassador
- **CI/CD Tools**: Jenkins, GitLab CI/CD, CircleCI
- **Monitoring and Logging**: Prometheus, Grafana, ELK Stack
- **Testing Frameworks**: JUnit, TestNG, Selenium
- **Package Management**: npm, pip, Maven

## Configuration-First and Extensible

*Source: [Configuration-First and Extensible](../Practices%20and%20Patterns/Configuration-First-and-Extensible.md)*

### Technology Patterns
- **Low-Code/No-Code Platforms**  
  Enable business users to configure services without coding.
- **Declarative Configuration**  
  Define system behavior through configuration files.
- **Plugin Architecture**  
  Extend functionality through modular plugins.
- **Template Systems**  
  Provide reusable templates for common configurations.

### Technology Choices
- **Configuration Management**: Ansible, Terraform
- **Low-Code Platforms**: Node-RED, Appsmith
- **Template Engines**: Jinja2, Handlebars
- **Plugin Frameworks**: OSGi, Spring Plugin Framework

## Intelligent

*Source: [Intelligent](../Practices%20and%20Patterns/Intelligent.md)*

### Technology Patterns
- **Conversational AI and Voice Assistants**  
  Guide users through services in local languages.
- **Multi-Modal User Interfaces**  
  Support various interaction modes.
- **Real-Time Monitoring and Alerting Systems**  
  Detect anomalies and trigger alerts.
- **Data Fusion and Analytics Platforms**  
  Merge data from multiple sources.

### Technology Choices
- **Language/Voice**: Bhashini, Mycroft, Vosk, SpeechBrain
- **Multi-Modal**: OpenOmni, ADVISER
- **Issue Detection**: OpenDroneMap, QGIS, TensorFlow, PyTorch
- **Data Integration**: Apache NiFi, Apache Kafka, Metabase 
