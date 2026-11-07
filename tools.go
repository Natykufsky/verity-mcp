package mcp

import (
	"context"
	"encoding/json"
)

// ToolDefinition models an MCP tool callable by an LLM or agent.
type ToolDefinition struct {
	Name        string                                                                 `json:"name"`
	Description string                                                                 `json:"description"`
	InputSchema map[string]any                                                         `json:"inputSchema"`
	Permission  string                                                                 `json:"permission,omitempty"`
	Handler     func(ctx context.Context, args json.RawMessage) ([]ContentItem, error) `json:"-"`
}

// ToolItem is the serialized schema representation for tools/list.
type ToolItem struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ToolsListResult is returned by tools/list.
type ToolsListResult struct {
	Tools []ToolItem `json:"tools"`
}

// ToolCallParams defines parameters received in tools/call.
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ToolCallResult is returned by tools/call.
type ToolCallResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError"`
}
