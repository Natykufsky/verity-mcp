# Verity MCP (Model Context Protocol Go SDK)

`verity-mcp` is a lightweight, zero-dependency Go implementation of Anthropic's **Model Context Protocol (MCP)** specification (JSON-RPC 2.0).

It allows any Go application to expose **Tools**, **Resources**, and **Prompts** to AI models and agents (Claude Desktop, Cursor, IDE extensions) over **HTTP** and **Stdio** transports.

---

## Features

- **Zero External Dependencies:** Go stdlib only (`net/http`, `encoding/json`, `context`, `sync`).
- **Full Protocol Specification:**
  - **Tools (`tools/list`, `tools/call`):** JSON schema inputs, permission gating, execution handlers.
  - **Resources (`resources/list`, `resources/read`):** Read-only system state, schemas, documents.
  - **Prompts (`prompts/list`, `prompts/get`):** Interactive AI workflow prompt templates.
- **Transports:**
  - JSON-RPC 2.0 over HTTP (`net/http`)
  - Interactive Stdio (`ServeStdio`) for Claude Desktop / Cursor
  - SSE endpoint stub (endpoint event only)
- **RBAC-Ready:** Pluggable `PermissionChecker` for tenant/role-level access control on every tool, resource, or prompt.

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
	"encoding/json"
	"fmt"
	"net/http"

	mcp "github.com/Natykufsky/verity-mcp"
)

func main() {
	server := mcp.NewServer(mcp.ServerConfig{
		Name:    "verity-analytics",
		Version: "1.0.0",
		CheckPermission: func(ctx context.Context, perm string) bool {
			return true // replace with real RBAC
		},
	})

	server.RegisterTool(mcp.ToolDefinition{
		Name:        "get_metrics",
		Description: "Fetches live system KPIs and analytics",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"period": map[string]any{
					"type": "string",
					"enum": []string{"day", "week", "month"},
				},
			},
		},
		Handler: func(ctx context.Context, args json.RawMessage) ([]mcp.ContentItem, error) {
			return []mcp.ContentItem{{
				Type: "text",
				Text: `{"active_users":15420,"uptime":"99.99%"}`,
			}}, nil
		},
	})

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

	if err := server.ServeStdio(context.Background(), nil, nil); err != nil {
		panic(err)
	}
}
```

---

## License

MIT License.

