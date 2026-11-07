# Verity MCP Platform (SaaS) — Product & Technical Specification

> **Positioning:** The platform is a **managed, multi-tenant runtime, gateway, and marketplace for MCP servers** — an “MCP PaaS / Enterprise AI Gateway” that lets teams deploy, govern, monetize, and observe MCP servers without managing infrastructure. It sits between AI clients (Claude Desktop, Cursor, Copilot Studio, LangChain, autonomous agents) and backends (APIs, databases, SaaS tools) that MCP servers expose.

---

## 1. 🏗️ Core Platform & Runtime Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                             CLIENT LAYER                                    │
│   Claude Desktop · Cursor · Copilot Studio · Custom AI Agents · Web Apps   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ HTTP / SSE / Stdio
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                   VERITY MCP GATEWAY & AUTH PIPELINE                        │
│  - SSE Transport Termination: /v1/sse/{workspace_id}                        │
│  - Dual-Channel Auth Pipeline (Shared JWT, user pool, billing ledger):       │
│      • /mcp/... (AI Agents & Marketplace)                                   │
│      • /api/... (Web / Mobile Frontends)                                    │
│  - Tenancy & Isolation (Powered by github.com/Natykufsky/verity-iam):       │
│      • Multi-tenant isolation via JWT claims (`aud` + `product_id`)         │
│      • Multi-Round-Trip State Binding: HMAC(user_id, product_id, jti)       │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                     CORE SAAS ENGINE & RUNTIME POOL                         │
│  ┌─────────────────────────┐  ┌─────────────────────────┐                   │
│  │   Dynamic Tool Registry │  │ Execution Sandbox       │                   │
│  │   - OpenAPI -> MCP      │  │ - Pre-warmed Worker Pool│                   │
│  │   - DB Query Connectors │  │ - Multi-Server Routing  │                   │
│  │   - Low-Code Tool Maker │  │ - Tool Name Prefixing   │                   │
│  └─────────────────────────┘  └─────────────────────────┘                   │
│  ┌─────────────────────────┐  ┌─────────────────────────┐                   │
│  │   UI Streaming Engine   │  │ Prompt Hub & Resources  │                   │
│  │   - ui:// SPA Resources │  │ - Tenant Context Inject │                   │
│  │   - Host Iframe Streams │  │ - RAG Vector Gateways   │                   │
│  └─────────────────────────┘  └─────────────────────────┘                   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
        ┌──────────────────────────────┼──────────────────────────────┐
        ▼                              ▼                              ▼
┌───────────────┐              ┌───────────────┐              ┌───────────────┐
│  PostgreSQL   │              │     Redis     │              │ Observability │
│ (Workspaces,  │              │ (Rate limits, │              │ (OpenTelemetry│
│  Tools, ACL,  │              │  SSE streams, │              │  Prometheus,  │
│  Ledger, Auth)│              │  credit state)│              │  Audit logs)  │
└───────────────┘              └───────────────┘              └───────────────┘
```

---

## 2. 🔌 Tooling, Connectors & Registry

| Feature | Description | Implementation Strategy |
|---|---|---|
| **Pre-built Connector Library** | Out-of-the-box integrations (Stripe, GitHub, Slack, Postgres, Jira, Google Suite, AccountAid). | Connector templates with pre-defined JSON schemas. |
| **Custom Connector / Tool Builder** | Low-code or code-first builder to wrap proprietary APIs or internal databases. | OpenAPI 3.0 / Swagger importer mapping endpoints to MCP tools. |
| **MCP Server Registry** | Searchable public/private catalog of MCP servers with versioning and one-click install. | Metadata store catalog with semantic search and permission badges. |
| **Tool Schema Management** | Real-time validation and dynamic capability discovery (`initialize` handshake). | JSON Schema Draft 7 validation with runtime query negotiation. |
| **Tool Prefixing & Routing** | Aggregates multiple remote MCP servers behind one endpoint with collision-free naming (e.g. `accountaid_create_invoice`, `crm_find_lead`). | Gateway namespace router dispatching sub-requests. |
| **UI Distribution (`ui://`)** | Stream pre-built single-page apps or auto-generated HTML forms directly to AI host iframes. | MCP resource provider serving `ui://` protocols and inline HTML widgets. |

---

## 3. 🔐 Security, Auth & Governance

| Feature | Description | Implementation Strategy |
|---|---|---|
| **OAuth 2.1 & Dynamic Client Registration (DCR)** | Full OAuth 2.1 authorization server with PKCE. MCP clients register dynamically without manual provisioning. | Built-in OAuth 2.1 engine integrated with `verity-iam`. |
| **Managed OAuth for Connectors** | Single user authorization dance per third-party service; credentials stored securely for recurring agent tool calls. | Encrypted token vault (AES-256-GCM) with automatic background token refresh. |
| **Enterprise SSO & Identity** | SAML 2.0 / OIDC integrations (Okta, Azure AD, Google Workspace). | Delegated identity provider layer mapped to tenant workspace IDs. |
| **RBAC & Per-Tool ACL** | Granular whitelist/blacklist: define which API keys or roles can execute *specific* tools. | `verity-iam.RequirePermission("tools:<tool_name>:execute")` at the gateway level. |
| **Tenant-Aware Tool Masking** | Prevents AI models from hallucinating unpermitted tools by filtering the `tools/list` response per tenant. | Dynamic filtering in `tools/list` handler based on caller JWT claims. |
| **Multi-Round-Trip State Binding** | Cryptographically bind interactive multi-step prompts to `(user_id, product_id, jti)` via HMAC. | Session state tokens signed with server secret preventing cross-tenant injection. |
| **Audit Trail & Data Masking** | Immutable structured audit log with PII redaction and export API. | Asynchronous logging pipeline writing redacted JSON events to PostgreSQL/ClickHouse. |

---

## 4. 💰 Monetization & Credit Ledger

| Feature | Description | Implementation Strategy |
|---|---|---|
| **Per-Tool Dynamic Pricing** | Configurable credit pricing per tool (e.g. `search_lead`: 1 credit, `generate_invoice`: 5 credits). | Tool pricing matrix table checked prior to execution. |
| **Credit Ledger & Auto-402** | Real-time credit deduction with automatic `402 Payment Required` HTTP response when credits are exhausted. | Atomic Redis `DECRBY` balance deduction with Postgres reconciliation. |
| **Spending Limits & Team Budgets** | Hard spending caps per API key, team budgets, daily/monthly call quotas, and auto-reset. | Sliding window rate and quota limiter in Redis. |
| **Payment Discovery (`/.well-known/mcp-payment`)** | Machine-readable pricing endpoint for autonomous agents to inspect cost before calling tools. | Standard HTTP endpoint exposing JSON pricing catalog. |
| **Tiered Plans & Overages** | Free / Developer / Pro / Enterprise subscription management. | Stripe / CatholicPay integration syncing customer plan limits. |

---

## 5. 📊 Observability & Operations

| Feature | Description | Implementation Strategy |
|---|---|---|
| **Real-Time Activity Monitor** | Live streaming request feed showing active tool calls, callers, latencies, and statuses. | WebSocket / SSE admin activity feed. |
| **Distributed Tracing & Metrics** | Trace-ID propagation across MCP Gateway $\rightarrow$ Backend APIs; standard `/metrics` endpoint. | OpenTelemetry spans + Prometheus exporter (`verity-iam/middleware`). |
| **Structured JSON Logs & Masking** | Structured logs with automatic masking of secrets and credit card numbers. | Go standard `log/slog` with custom masking handler. |
| **Health Checks & Auto-Recovery** | Background probes testing liveness and readiness of wrapped customer backends. | Periodic health checker with circuit breaker tripping. |

---

## 6. 🗺️ Implementation Roadmap

### Phase 1: MVP (Must-Have Foundation)
1. **Managed MCP Server Gateway**: Multi-tenant HTTP/SSE runtime powered by `verity-mcp` + `go-chi`.
2. **OAuth 2.1 & API Key Auth**: Integrated with `verity-iam` for tenant workspaces and user credentials.
3. **Tool Registry & Execution Proxy**: Dynamic tool registration with JSON schema validation.
4. **Per-Tool Credit Ledger**: Redis-backed credit deduction with `402 Payment Required` flow.
5. **Basic Activity Log**: Structured audit log of tool calls and errors.

### Phase 2: SaaS Differentiation & Observability
1. **Tenant-Aware Tool Masking**: Filter `tools/list` so agents only see allowed tools.
2. **RBAC & Per-Tool ACL**: Whitelist/blacklist tools per API key and role.
3. **OpenAPI 3.0 Auto-Importer**: Convert any OpenAPI spec into an instant MCP server.
4. **Spending Limits, Quotas & Webhook Alerts**: Configurable budgets and threshold notifications.
5. **Prometheus Metrics & Distributed Tracing**: Standard `/metrics` and trace propagation.

### Phase 3: Enterprise, Monetization & Rich UI
1. **UI Distribution (`ui://`)**: Stream interactive HTML widgets directly to host iframes.
2. **Multi-Server Aggregation**: Mount multiple remote MCP servers behind one unified endpoint with prefix routing.
3. **Agent-Native Payment Discovery**: `/.well-known/mcp-payment` endpoint.
4. **Enterprise SSO & SAML / OIDC**.
5. **Encrypted Managed OAuth Vault** for third-party connector tokens.
