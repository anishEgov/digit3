openapi: 3.0.3
info:
  title: Workflow service APIs definitions
  version: 3.0.0
  description: API for creating, updating, and searching workflow processes and their states/actions
servers:
  - url: /workflow/v3
security:
  - BearerAuth: []
tags:
  - name: Process
    description: Operations related to workflow processes
  - name: State
    description: Operations for managing states within a process
  - name: Action
    description: Operations for managing transitions between states

components:
  securitySchemes:
    BearerAuth:
      $ref: 'https://raw.githubusercontent.com/egovernments/DIGIT-3.0/master/common-3.0.0.yaml#/components/securitySchemes/BearerAuth'

  parameters:
    RequestIdHeader:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/parameters/RequestIdHeader'
    CorrelationIdHeader:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/parameters/CorrelationIdHeader'
    UserAgentHeader:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/parameters/UserAgentHeader'
    TenantIdHeader:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/parameters/TenantIdHeader'
    ForwardedForHeader:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/parameters/ForwardedForHeader'
    ClientId:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/parameters/ClientId'
    ClientSecret:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/parameters/ClientSecret'
    TimeStampHeader:
      name: X-Timestamp
      in: header
      description: Request timestamp (epoch millis)
      schema:
        type: integer
        format: int64
    ProcessIdPath:
      name: processId
      in: path
      required: true
      schema:
        type: string
      description: Unique identifier of the process
    IdPath:
      name: id
      in: path
      required: true
      schema:
        type: string
      description: Unique identifier of the resource
    StateIdPath:
      name: stateid
      in: path
      required: true
      schema:
        type: string
      description: Unique identifier of the state

  headers:
    X-Response-Time:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Response-Time'
    X-Response-Timestamp:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Response-Timestamp'
    X-Request-ID:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Request-ID'
    X-Correlation-ID:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Correlation-ID'
    X-Tenant-ID:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Tenant-ID'
    X-Rate-Limit:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Rate-Limit'
    X-Rate-Limit-Remaining:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Rate-Limit-Remaining'
    X-Rate-Limit-Reset:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/headers/X-Rate-Limit-Reset'

  schemas:
    Error:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/schemas/Error'
    AuditDetail:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/schemas/AuditDetail'
    Document:
      $ref: 'https://raw.githubusercontent.com/digitnxt/digit3/refs/heads/master/docs/Specifications/Common/Common-3.0.0.yaml#/components/schemas/Document'

    Process:
      type: object
      required:
        - id
        - name
        - code
      properties:
        id:
          type: string
          description: System-generated unique identifier
          readOnly: true
          minLength: 2
          maxLength: 64
        name:
          type: string
          description: Human-readable name
          minLength: 2
          maxLength: 128
        code:
          type: string
          description: Unique process code
          minLength: 2
          maxLength: 128
        description:
          type: string
          description: Detailed description
          minLength: 4
          maxLength: 512
        version:
          type: string
          description: Version string
          minLength: 1
          maxLength: 32
        sla:
          type: integer
          format: int64
          description: SLA duration (minutes)
        auditDetail:
          $ref: '#/components/schemas/AuditDetail'

    State:
      type: object
      required:
        - code
        - name
        - processId
      properties:
        code:
          type: string
          description: State code
          minLength: 2
          maxLength: 64
        name:
          type: string
          description: State name
          minLength: 2
          maxLength: 128
        description:
          type: string
          description: State description
          maxLength: 512
        processId:
          type: string
          description: Parent process ID
        sla:
          type: integer
          format: int64
          description: SLA duration (minutes)
        isInitial:
          type: boolean
          description: Initial state flag
        isParallel:
          type: boolean
          description: Parallel branch flag
        isJoin:
          type: boolean
          description: Join state flag
        branchStates:
          type: array
          items:
            type: string
          description: Parallel branch state codes
        auditDetail:
          $ref: '#/components/schemas/AuditDetail'

    Action:
      type: object
      required:
        - id
        - name
        - currentState
        - nextState
      properties:
        id:
          type: string
          description: Action identifier
          readOnly: true
        name:
          type: string
          description: Action name
          minLength: 2
          maxLength: 64
        label:
          type: string
          description: Display label
          maxLength: 128
        currentState:
          type: string
          description: Origin state code
        nextState:
          type: string
          description: Destination state code
        roles:
          type: array
          items:
            type: string
          description: Allowed roles
        attributeValidation:
          $ref: '#/components/schemas/AttributeValidation'
        auditDetail:
          $ref: '#/components/schemas/AuditDetail'

    StateDetail:
      allOf:
        - $ref: '#/components/schemas/State'
        - type: object
          properties:
            actions:
              type: array
              items:
                $ref: '#/components/schemas/Action'

    ProcessDefinitionDetail:
      allOf:
        - $ref: '#/components/schemas/Process'
        - type: object
          properties:
            states:
              type: array
              items:
                $ref: '#/components/schemas/StateDetail'

    ProcessInstance:
      type: object
      properties:
        id:
          type: string
          description: Instance ID
          readOnly: true
          minLength: 2
          maxLength: 64
        processId:
          type: string
          description: Definition ID
        entityId:
          type: string
          description: Entity ID
        action:
          type: string
          description: Action code
          minLength: 2
          maxLength: 128
        status:
          type: string
          description: Status (readOnly)
          minLength: 2
          maxLength: 64
        comment:
          type: string
          description: Comment
          minLength: 2
          maxLength: 512
        documents:
          type: array
          items:
            $ref: '#/components/schemas/Document'
        assigner:
          type: string
        assignees:
          type: array
          items:
            type: string
        currentState:
          type: string
          description: Current state code
        stateSla:
          type: integer
          format: int64
          description: State SLA start timestamp (epoch ms)
        processSla:
          type: integer
          format: int64
          description: Process SLA start timestamp (epoch ms)
        attributes:
          type: object
          additionalProperties:
            type: array
            items:
              type: string
          description: Guard evaluation attributes
        auditDetails:
          $ref: '#/components/schemas/AuditDetail'

    AttributeValidation:
      type: object
      required:
        - id
      properties:
        id:
          type: string
          description: Guard ID
          readOnly: true
        attributes:
          type: object
          additionalProperties:
            type: array
            items:
              type: string
          description: Allowed attribute values
        assigneeCheck:
          type: boolean
          description: Enforce assignee match
        auditDetail:
          $ref: '#/components/schemas/AuditDetail'

paths:
  /process:
    post:
      summary: Create a new process
      tags: [Process]
      parameters:
        - $ref: '#/components/parameters/RequestIdHeader'
        - $ref: '#/components/parameters/CorrelationIdHeader'
        - $ref: '#/components/parameters/UserAgentHeader'
        - $ref: '#/components/parameters/TenantIdHeader'
        - $ref: '#/components/parameters/ForwardedForHeader'
        - $ref: '#/components/parameters/ClientId'
        - $ref: '#/components/parameters/ClientSecret'
        - $ref: '#/components/parameters/TimeStampHeader'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Process'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Process'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    get:
      summary: List processes
      tags: [Process]
      parameters:
        - $ref: '#/components/parameters/RequestIdHeader'
        - $ref: '#/components/parameters/CorrelationIdHeader'
        - $ref: '#/components/parameters/UserAgentHeader'
        - $ref: '#/components/parameters/TenantIdHeader'
        - $ref: '#/components/parameters/ForwardedForHeader'
        - $ref: '#/components/parameters/ClientId'
        - $ref: '#/components/parameters/ClientSecret'
        - $ref: '#/components/parameters/TimeStampHeader'
        - in: query
          name: id
          schema:
            type: array
            items:
              type: string
            minItems: 1
        - in: query
          name: name
          schema:
            type: array
            items:
              type: string
            minItems: 1
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Process'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

  /process/{id}:
    parameters:
      - $ref: '#/components/parameters/RequestIdHeader'
      - $ref: '#/components/parameters/CorrelationIdHeader'
      - $ref: '#/components/parameters/UserAgentHeader'
      - $ref: '#/components/parameters/TenantIdHeader'
      - $ref: '#/components/parameters/ForwardedForHeader'
      - $ref: '#/components/parameters/ClientId'
      - $ref: '#/components/parameters/ClientSecret'
      - $ref: '#/components/parameters/TimeStampHeader'
      - $ref: '#/components/parameters/IdPath'

    put:
      summary: Update a process
      tags: [Process]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Process'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Process'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    get:
      summary: Get a process by ID
      tags: [Process]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Process'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    delete:
      summary: Delete a process
      tags: [Process]
      responses:
        '204':
          description: No Content
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

  /process/definition:
    get:
      summary: List process definitions
      tags: [Process]
      parameters:
        - $ref: '#/components/parameters/RequestIdHeader'
        - $ref: '#/components/parameters/CorrelationIdHeader'
        - $ref: '#/components/parameters/UserAgentHeader'
        - $ref: '#/components/parameters/TenantIdHeader'
        - $ref: '#/components/parameters/ForwardedForHeader'
        - $ref: '#/components/parameters/ClientId'
        - $ref: '#/components/parameters/ClientSecret'
        - $ref: '#/components/parameters/TimeStampHeader'
        - in: query
          name: id
          schema:
            type: array
            items:
              type: string
            minItems: 1
        - in: query
          name: name
          schema:
            type: array
            items:
              type: string
            minItems: 1
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/ProcessDefinitionDetail'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

  /process/{processId}/state:
    parameters:
      - $ref: '#/components/parameters/RequestIdHeader'
      - $ref: '#/components/parameters/CorrelationIdHeader'
      - $ref: '#/components/parameters/UserAgentHeader'
      - $ref: '#/components/parameters/TenantIdHeader'
      - $ref: '#/components/parameters/ForwardedForHeader'
      - $ref: '#/components/parameters/ClientId'
      - $ref: '#/components/parameters/ClientSecret'
      - $ref: '#/components/parameters/TimeStampHeader'
      - $ref: '#/components/parameters/ProcessIdPath'

    post:
      summary: Create a new state
      tags: [State]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/State'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/State'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    get:
      summary: List states for a process
      tags: [State]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/State'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

  /state/{id}:
    parameters:
      - $ref: '#/components/parameters/RequestIdHeader'
      - $ref: '#/components/parameters/CorrelationIdHeader'
      - $ref: '#/components/parameters/UserAgentHeader'
      - $ref: '#/components/parameters/TenantIdHeader'
      - $ref: '#/components/parameters/ForwardedForHeader'
      - $ref: '#/components/parameters/ClientId'
      - $ref: '#/components/parameters/ClientSecret'
      - $ref: '#/components/parameters/TimeStampHeader'
      - $ref: '#/components/parameters/IdPath'

    put:
      summary: Update a state
      tags: [State]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/State'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/State'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    get:
      summary: Get a state by ID
      tags: [State]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/State'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    delete:
      summary: Delete a state
      tags: [State]
      responses:
        '204':
          description: No Content
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

  /state/{stateid}/action:
    parameters:
      - $ref: '#/components/parameters/RequestIdHeader'
      - $ref: '#/components/parameters/CorrelationIdHeader'
      - $ref: '#/components/parameters/UserAgentHeader'
      - $ref: '#/components/parameters/TenantIdHeader'
      - $ref: '#/components/parameters/ForwardedForHeader'
      - $ref: '#/components/parameters/ClientId'
      - $ref: '#/components/parameters/ClientSecret'
      - $ref: '#/components/parameters/TimeStampHeader'
      - $ref: '#/components/parameters/StateIdPath'

    post:
      summary: Create a new action
      tags: [Action]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Action'
      responses:
        '201':
          description: Created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Action'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    get:
      summary: List actions for a state
      tags: [Action]
      parameters:
        - in: query
          name: currentState
          schema:
            type: string
        - in: query
          name: nextState
          schema:
            type: string
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Action'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

  /action/{id}:
    parameters:
      - $ref: '#/components/parameters/RequestIdHeader'
      - $ref: '#/components/parameters/CorrelationIdHeader'
      - $ref: '#/components/parameters/UserAgentHeader'
      - $ref: '#/components/parameters/TenantIdHeader'
      - $ref: '#/components/parameters/ForwardedForHeader'
      - $ref: '#/components/parameters/ClientId'
      - $ref: '#/components/parameters/ClientSecret'
      - $ref: '#/components/parameters/TimeStampHeader'
      - $ref: '#/components/parameters/IdPath'

    put:
      summary: Update an action
      tags: [Action]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Action'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Action'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    get:
      summary: Get an action by ID
      tags: [Action]
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Action'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

    delete:
      summary: Delete an action
      tags: [Action]
      responses:
        '204':
          description: No Content
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'   
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'  
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Bad Request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error'

  /transition:
    post:
      summary: Transition a process instance to the next state
      description: |
        Triggers a state transition on an existing process instance, verifying guard conditions
        (role/attribute/assignee checks) before performing the transition.
      tags: [Process]
      parameters:
        - $ref: '#/components/parameters/RequestIdHeader'
        - $ref: '#/components/parameters/CorrelationIdHeader'
        - $ref: '#/components/parameters/UserAgentHeader'
        - $ref: '#/components/parameters/TenantIdHeader'
        - $ref: '#/components/parameters/ForwardedForHeader'
        - $ref: '#/components/parameters/ClientId'
        - $ref: '#/components/parameters/ClientSecret'
        - $ref: '#/components/parameters/TimeStampHeader'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ProcessInstance'
      responses:
        '202':
          description: Accepted
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ProcessInstance'
          headers:
            X-Response-Time:
              $ref: '#/components/headers/X-Response-Time'
            X-Response-Timestamp:
              $ref: '#/components/headers/X-Response-Timestamp'
            X-Request-ID:
              $ref: '#/components/headers/X-Request-ID'
            X-Correlation-ID:
              $ref: '#/components/headers/X-Correlation-ID'
            X-Tenant-ID:
              $ref: '#/components/headers/X-Tenant-ID'
            X-Rate-Limit:
              $ref: '#/components/headers/X-Rate-Limit'
            X-Rate-Limit-Remaining:
              $ref: '#/components/headers/X-Rate-Limit-Remaining'
            X-Rate-Limit-Reset:
              $ref: '#/components/headers/X-Rate-Limit-Reset'
        '400':
          description: Guard validation failed or invalid request
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Error' 