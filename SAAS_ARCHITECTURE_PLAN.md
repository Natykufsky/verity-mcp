# Verity MCP Platform (SaaS) Architecture & Execution Plan

## 1. Executive Summary

**Verity MCP Platform** transforms the `verity-mcp` protocol core into a multi-tenant, hosted Model Context Protocol (MCP) Cloud SaaS. 

It enables businesses, AI development teams, and external enterprise clients (such as MogRex AccountAid, CatholicPay, and HarvestPad) to deploy, secure, monitor, monetize, and orchestrate MCP tool registries, agent integrations, and context resources without managing self-hosted infrastructure.

---

## 2. Platform Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                             CLIENT LAYER                                    │
│  AI Agents (Claude Desktop, Cursor, DeepSeek, LangChain, Custom AI Apps)    │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ Bearer API Key / JWT
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          EDGE & GATEWAY LAYER                               │
│  - SSE Transport Termination: https://mcp.verity.cloud/v1/sse/{workspace_id}│
│  - Auth & Tenancy Gateway (Powered by github.com/Natykufsky/verity-iam)      │
│  - Rate Limiting, Request Validation & Quota Metering                       │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │ Validated RPC
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                       VERITY MCP SAAS CORE ENGINE                           │
│  ┌─────────────────────────┐  ┌─────────────────────────┐                   │
│  │   Dynamic Tool Registry │  │ Execution Engine (Proxy)│                   │
│  │   - OpenAPI -> MCP      │  │ - Safe HTTP Tool Exec   │                   │
│  │   - SQL / DB Connectors │  │ - Webhook / Async Exec  │                   │
│  │   - Custom JS/Go Runner │  │ - Tenant Context Inject │                   │
│  └─────────────────────────┘  └─────────────────────────┘                   │
│  ┌─────────────────────────┐  ┌─────────────────────────┐                   │
│  │   Resource Gateway      │  │ Prompt Hub & Versioning │                   │
│  │   - S3/Storage Context  │  │ - Dynamic System Prompts│                   │
│  │   - Vector RAG Context  │  │ - Multi-tenant Templates│                   │
│  └─────────────────────────┘  └─────────────────────────┘                   │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
        ┌──────────────────────────────┼──────────────────────────────┐
        ▼                              ▼                              ▼
┌───────────────┐              ┌───────────────┐              ┌───────────────┐
│  PostgreSQL   │              │     Redis     │              │ Audit Log /   │
│ (Metadata &   │              │ (Token Cache, │              │ ClickHouse    │
│  Tenant CoA)  │              │  SSE Streams) │              │ (Usage Stats) │
└───────────────┘              └───────────────┘              └───────────────┘
```

---

## 3. Core Modules & Responsibilities

### 3.1. Gateway & Authentication (Powered by `verity-iam`)
- **API Key & Bearer Tokens**: Client-facing endpoints authenticate using `verity-iam` middleware (`RequireAuth()`, `RequireTenant()`).
- **Workspace Isolation**: Every MCP connection specifies or resolves a `workspace_id` / `tenant_id`. Tools registered in Workspace A cannot be enumerated or invoked by Workspace B.
- **Granular Permission Gating**: Enforces RBAC permissions per tool (e.g. `tools:invoices:read`, `tools:payments:execute`).

### 3.2. Dynamic Tool Registry
1. **HTTP / REST Tool Proxies**: Users register OpenAPI/Swagger endpoints; Verity MCP dynamically maps REST endpoints into MCP tools with input JSON schemas.
2. **Database Connectors**: Secure, read-only SQL execution tools against customer databases with query sanitization.
3. **Webhook & Event Triggers**: Asynchronous action dispatch with audit confirmation steps.

### 3.3. Safe Execution Sandbox
- Outbound egress policy enforcement.
- Per-tool execution timeouts and circuit breakers.
- Secrets vault integration (customer API keys never leave the server unencrypted).

### 3.4. Usage Metering & Billing
- Granular tracking of tool calls, prompt tokens, and SSE session duration.
- Tiered subscription quotas (Free, Developer, Team, Enterprise).

---

## 4. Phase-by-Phase Roadmap

### Phase 1: SaaS Foundation & HTTP/SSE Gateway (Weeks 1–2)
- [ ] Implement `cmd/mcp-server/main.go` multi-tenant server.
- [ ] Integrate `github.com/Natykufsky/verity-iam` for tenant workspaces and API key auth.
- [ ] Add dynamic workspace SSE routing: `/v1/sse?workspace_id={id}`.
- [ ] In-memory session tracking with Redis Pub/Sub broadcast for clustered SSE.

### Phase 2: Dynamic Tool & Resource Registry (Weeks 3–4)
- [ ] Database schema for dynamic tools:
  - `mcp_workspaces`, `mcp_tools`, `mcp_resources`, `mcp_api_keys`.
- [ ] OpenAPI 3.0 Importer: parse Swagger/OpenAPI specs into registered MCP tools.
- [ ] Secure HTTP proxy dispatcher for dynamic tool invocation.

### Phase 3: Developer Dashboard & Observability (Weeks 5–6)
- [ ] Admin & Developer Web UI:
  - Workspace management & API key generation.
  - Interactive Tool Playground (test tool calls in browser).
  - Real-time invocation logs, latencies, and error tracking.
- [ ] Export standard client configs (one-click copy for Claude Desktop `claude_desktop_config.json`, Cursor, and VS Code).

### Phase 4: Monetization & Enterprise Readiness (Weeks 7–8)
- [ ] Usage metering & rate-limiting per plan.
- [ ] Stripe / CatholicPay subscription billing integration.
- [ ] Immutable audit logging for tool executions (caller identity, inputs, execution timestamp, response status).

---

## 5. Technology Stack

- **Core Engine & Transports**: Go 1.22+, `github.com/Natykufsky/verity-mcp`
- **Identity & Multi-Tenancy**: `github.com/Natykufsky/verity-iam`
- **HTTP Routing**: `go-chi/chi/v5`
- **Database**: PostgreSQL 16 (`pgx` / `lib/pq` with versioned migrations)
- **Cache & Pub/Sub**: Redis 7
- **Deployment**: Docker, Docker Compose, Kubernetes / Helm

