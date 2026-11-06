package mcp

import (
	"encoding/json"
)

// MCP Protocol and Error Standards (Specification: 2024-11-05).
const (
	ProtocolVersion = "2024-11-05"

	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
	CodeForbidden      = -32003
)

// Request represents an incoming JSON-RPC 2.0 message.
type Request struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

// Error represents a structured JSON-RPC 2.0 error.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string
	ID      any
	Result  any
	Error   *Error
}

// MarshalJSON guarantees strict JSON-RPC 2.0 compliance:
// - On success: "result" field is present (even when empty) and "error" is omitted.
// - On error: "error" field is present and "result" is omitted.
func (r Response) MarshalJSON() ([]byte, error) {
	if r.Error != nil {
		return json.Marshal(struct {
			JSONRPC string `json:"jsonrpc"`
			ID      any    `json:"id"`
			Error   *Error `json:"error"`
		}{
			JSONRPC: "2.0",
			ID:      r.ID,
			Error:   r.Error,
		})
	}
	return json.Marshal(struct {
		JSONRPC string `json:"jsonrpc"`
		ID      any    `json:"id"`
		Result  any    `json:"result"`
	}{
		JSONRPC: "2.0",
		ID:      r.ID,
		Result:  r.Result,
	})
}

// InitializeResult represents the outcome of an "initialize" handshake.
type InitializeResult struct {
	ProtocolVersion string          `json:"protocolVersion"`
	Capabilities    Capabilities    `json:"capabilities"`
	ServerInfo      ServerInfo      `json:"serverInfo"`
	Instructions    string          `json:"instructions,omitempty"`
}

// Capabilities declare what the server supports.
type Capabilities struct {
	Tools     CapabilityTools     `json:"tools"`
	Resources CapabilityResources `json:"resources"`
	Prompts   CapabilityPrompts   `json:"prompts"`
}

type CapabilityTools struct {
	ListChanged bool `json:"listChanged"`
}

type CapabilityResources struct {
	Subscribe   bool `json:"subscribe"`
	ListChanged bool `json:"listChanged"`
}

type CapabilityPrompts struct {
	ListChanged bool `json:"listChanged"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ContentItem represents a formatted response payload item (text, image, resource).
type ContentItem struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}
