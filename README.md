# Verity MCP (Model Context Protocol Go SDK)

`verity-mcp` is a lightweight, zero-dependency Go implementation of Anthropic's **Model Context Protocol (MCP)** specification (JSON-RPC 2.0).

It allows any Go application or backend service to expose **Tools**, **Resources**, and **Prompts** to AI models, agent swarms, and desktop assistants (e.g., Claude Desktop, Cursor, and IDE extensions) across both **HTTP Server-Sent Events (SSE)** and **Stdio** transports.

---

## Features

- **Zero External Dependencies:** Built entirely with Go standard library (`net/http`, `encoding/json`, `context`, `sync`).
- **Full Protocol Specification:**
  - **Tools (`tools/list`, `tools/call`):** Dynamic schema parameters, custom permissions, and execution handlers.
  - **Resources (`resources/list`, `resources/read`):** Read-only system state, schemas, and live documents.
  - **Prompts (`prompts/list`, `prompts/get`):** Interactive AI workflow prompt templates.
- **Transports:**
  - Standard JSON-RPC 2.0 HTTP Handler (`net/http`)
  - Server-Sent Events (SSE) streaming (`Accept: text/event-stream`)
  - Interactive Stdio CLI (`ServeStdio`) for local sub-processes.
- **Enterprise RBAC Ready:** Pluggable `PermissionChecker` func to enforce tenant-level or role-level access control on every tool, prompt, or resource.

---

## Installation

```bash
go get github.com/Natykufsky/verity-mcp
```

---

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Natykufsky/verity-mcp"
)

func main() {
	// 1. Initialize MCP Server
	server := mcp.NewServer(mcp.ServerConfig{
		Name:    "verity-analytics",
		Version: "1.0.0",
		CheckPermission: func(ctx context.Context, perm string) bool {
			// Optional custom authorization check
			return true
		},
	})

	// 2. Register a Tool
	server.RegisterTool(mcp.ToolDefinition{
		Name:        "get_metrics",
		Description: "Fetches live system KPIs and analytics",
		ParametersSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"period": map[string]any{
					"type": "string",
					"enum": []string{"day", "week", "month"},
				},
			},
		},
		Handler: func(ctx context.Context, argsJSON string) (any, error) {
			return map[string]any{
				"active_users": 15420,
				"uptime":       "99.99%",
			}, nil
		},
	})

	// 3. Mount HTTP Handler (supports standard JSON-RPC POST and SSE GET)
	http.Handle("/mcp", server.HTTPHandler())

	fmt.Println("MCP Server running on :8080/mcp")
	_ = http.ListenAndServe(":8080", nil)
}
```

---

## Stdio CLI Usage (Claude Desktop / Cursor)

```go
func main() {
	server := mcp.NewServer(mcp.ServerConfig{Name: "my-cli"})
	// Register tools...

	// Runs interactive JSON-RPC over stdin/stdout
	if err := server.ServeStdio(context.Background()); err != nil {
		panic(err)
	}
}
```

---

## License

MIT License.
