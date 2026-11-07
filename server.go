package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

// PermissionChecker verifies whether the given context caller holds a required permission.
type PermissionChecker func(ctx context.Context, permission string) bool

// ServerConfig configures an MCP Server instance.
type ServerConfig struct {
	Name            string
	Version         string
	Instructions    string
	CheckPermission PermissionChecker
}

// Server implements the Model Context Protocol (MCP) JSON-RPC 2.0 engine.
type Server struct {
	mu           sync.RWMutex
	name         string
	version      string
	instructions string
	checkPerm    PermissionChecker

	tools     map[string]ToolDefinition
	resources map[string]ResourceDefinition
	prompts   map[string]PromptDefinition
}

// NewServer creates a new modular MCP Server.
func NewServer(cfg ServerConfig) *Server {
	name := strings.TrimSpace(cfg.Name)
	if name == "" {
		name = "accountaid-mcp"
	}
	ver := strings.TrimSpace(cfg.Version)
	if ver == "" {
		ver = "1.0.0"
	}

	return &Server{
		name:         name,
		version:      ver,
		instructions: cfg.Instructions,
		checkPerm:    cfg.CheckPermission,
		tools:        make(map[string]ToolDefinition),
		resources:    make(map[string]ResourceDefinition),
		prompts:      make(map[string]PromptDefinition),
	}
}

// RegisterTool adds a tool to the server.
func (s *Server) RegisterTool(tool ToolDefinition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
}

// RegisterResource adds a resource to the server.
func (s *Server) RegisterResource(res ResourceDefinition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[res.URI] = res
}

// RegisterPrompt adds an interactive prompt template to the server.
func (s *Server) RegisterPrompt(prompt PromptDefinition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.prompts[prompt.Name] = prompt
}

// HandleMessage processes a raw JSON-RPC 2.0 message (single or batch).
func (s *Server) HandleMessage(ctx context.Context, data []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return json.Marshal(Response{
			ID:    nil,
			Error: &Error{Code: CodeInvalidRequest, Message: "Empty request body"},
		})
	}

	// Batch request
	if trimmed[0] == '[' {
		var rawList []Request
		if err := json.Unmarshal(trimmed, &rawList); err != nil {
			return json.Marshal(Response{
				ID:    nil,
				Error: &Error{Code: CodeParseError, Message: "Parse error: invalid JSON array"},
			})
		}
		if len(rawList) == 0 {
			return json.Marshal(Response{
				ID:    nil,
				Error: &Error{Code: CodeInvalidRequest, Message: "Invalid Request: empty batch"},
			})
		}

		var responses []Response
		for _, req := range rawList {
			resp := s.dispatch(ctx, req)
			if resp != nil {
				responses = append(responses, *resp)
			}
		}
		if len(responses) == 0 {
			return nil, nil
		}
		return json.Marshal(responses)
	}

	// Single request
	var req Request
	if err := json.Unmarshal(trimmed, &req); err != nil {
		return json.Marshal(Response{
			ID:    nil,
			Error: &Error{Code: CodeParseError, Message: "Parse error: invalid JSON format"},
		})
	}

	resp := s.dispatch(ctx, req)
	if resp == nil {
		return nil, nil
	}
	return json.Marshal(resp)
}

func (s *Server) dispatch(ctx context.Context, req Request) *Response {
	var id any
	if req.ID != nil {
		_ = json.Unmarshal(*req.ID, &id)
	}

	if req.JSONRPC != "2.0" {
		if id == nil {
			return nil
		}
		return &Response{
			ID:    id,
			Error: &Error{Code: CodeInvalidRequest, Message: "Invalid Request: jsonrpc version must be 2.0"},
		}
	}

	isNotification := req.ID == nil

	switch req.Method {
	case "initialize":
		res := InitializeResult{
			ProtocolVersion: ProtocolVersion,
			Capabilities: Capabilities{
				Tools:     CapabilityTools{ListChanged: false},
				Resources: CapabilityResources{Subscribe: false, ListChanged: false},
				Prompts:   CapabilityPrompts{ListChanged: false},
			},
			ServerInfo: ServerInfo{
				Name:    s.name,
				Version: s.version,
			},
			Instructions: s.instructions,
		}
		if isNotification {
			return nil
		}
		return &Response{ID: id, Result: res}

	case "ping":
		if isNotification {
			return nil
		}
		return &Response{ID: id, Result: map[string]any{}}

	case "notifications/initialized":
		return nil

	case "tools/list":
		if isNotification {
			return nil
		}
		s.mu.RLock()
		defer s.mu.RUnlock()

		items := make([]ToolItem, 0, len(s.tools))
		for _, tool := range s.tools {
			if tool.Permission != "" && s.checkPerm != nil && !s.checkPerm(ctx, tool.Permission) {
				continue
			}
			items = append(items, ToolItem{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.InputSchema,
			})
		}
		return &Response{ID: id, Result: ToolsListResult{Tools: items}}

	case "tools/call":
		if isNotification {
			return nil
		}
		var params ToolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil || params.Name == "" {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInvalidParams, Message: "Invalid params for tools/call: 'name' is required"},
			}
		}

		s.mu.RLock()
		tool, exists := s.tools[params.Name]
		s.mu.RUnlock()

		if !exists {
			return &Response{
				ID: id,
				Result: ToolCallResult{
					Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Tool %q not found", params.Name)}},
					IsError: true,
				},
			}
		}

		if tool.Permission != "" && s.checkPerm != nil && !s.checkPerm(ctx, tool.Permission) {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeForbidden, Message: fmt.Sprintf("Forbidden: missing %q permission", tool.Permission)},
			}
		}

		// FIX: guard against nil handler.
		if tool.Handler == nil {
			return &Response{
				ID: id,
				Result: ToolCallResult{
					Content: []ContentItem{{Type: "text", Text: fmt.Sprintf("Tool %q has no handler registered", params.Name)}},
					IsError: true,
				},
			}
		}

		content, err := tool.Handler(ctx, params.Arguments)
		if err != nil {
			return &Response{
				ID: id,
				Result: ToolCallResult{
					Content: []ContentItem{{Type: "text", Text: err.Error()}},
					IsError: true,
				},
			}
		}
		return &Response{ID: id, Result: ToolCallResult{Content: content, IsError: false}}

	case "resources/list":
		if isNotification {
			return nil
		}
		s.mu.RLock()
		defer s.mu.RUnlock()

		items := make([]ResourceItem, 0, len(s.resources))
		for _, res := range s.resources {
			if res.Permission != "" && s.checkPerm != nil && !s.checkPerm(ctx, res.Permission) {
				continue
			}
			items = append(items, ResourceItem{
				URI:         res.URI,
				Name:        res.Name,
				Description: res.Description,
				MimeType:    res.MimeType,
			})
		}
		return &Response{ID: id, Result: ResourcesListResult{Resources: items}}

	case "resources/read":
		if isNotification {
			return nil
		}
		var params ResourceReadParams
		if err := json.Unmarshal(req.Params, &params); err != nil || params.URI == "" {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInvalidParams, Message: "Invalid params: 'uri' is required"},
			}
		}

		s.mu.RLock()
		res, exists := s.resources[params.URI]
		s.mu.RUnlock()

		if !exists {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInvalidParams, Message: fmt.Sprintf("Resource not found: %s", params.URI)},
			}
		}

		if res.Permission != "" && s.checkPerm != nil && !s.checkPerm(ctx, res.Permission) {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeForbidden, Message: fmt.Sprintf("Forbidden: missing %q permission", res.Permission)},
			}
		}

		// FIX: guard against nil reader.
		if res.Reader == nil {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInternalError, Message: fmt.Sprintf("Resource %q has no reader registered", params.URI)},
			}
		}

		contents, err := res.Reader(ctx, params.URI)
		if err != nil {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInternalError, Message: err.Error()},
			}
		}
		return &Response{ID: id, Result: ResourceReadResult{Contents: contents}}

	case "prompts/list":
		if isNotification {
			return nil
		}
		s.mu.RLock()
		defer s.mu.RUnlock()

		items := make([]PromptItem, 0, len(s.prompts))
		for _, p := range s.prompts {
			if p.Permission != "" && s.checkPerm != nil && !s.checkPerm(ctx, p.Permission) {
				continue
			}
			items = append(items, PromptItem{
				Name:        p.Name,
				Description: p.Description,
				Arguments:   p.Arguments,
			})
		}
		return &Response{ID: id, Result: PromptsListResult{Prompts: items}}

	case "prompts/get":
		if isNotification {
			return nil
		}
		var params PromptGetParams
		if err := json.Unmarshal(req.Params, &params); err != nil || params.Name == "" {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInvalidParams, Message: "Invalid params: 'name' is required"},
			}
		}

		s.mu.RLock()
		p, exists := s.prompts[params.Name]
		s.mu.RUnlock()

		if !exists {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInvalidParams, Message: fmt.Sprintf("Prompt %q not found", params.Name)},
			}
		}

		if p.Permission != "" && s.checkPerm != nil && !s.checkPerm(ctx, p.Permission) {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeForbidden, Message: fmt.Sprintf("Forbidden: missing %q permission", p.Permission)},
			}
		}

		// FIX: guard against nil handler.
		if p.Handler == nil {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInternalError, Message: fmt.Sprintf("Prompt %q has no handler registered", params.Name)},
			}
		}

		messages, err := p.Handler(ctx, params.Arguments)
		if err != nil {
			return &Response{
				ID:    id,
				Error: &Error{Code: CodeInternalError, Message: err.Error()},
			}
		}
		return &Response{ID: id, Result: PromptGetResult{Description: p.Description, Messages: messages}}

	default:
		if isNotification {
			return nil
		}
		return &Response{
			ID:    id,
			Error: &Error{Code: CodeMethodNotFound, Message: fmt.Sprintf("Method %q not supported", req.Method)},
		}
	}
}

// HTTPHandler returns a standard net/http handler supporting POST JSON-RPC.
func (s *Server) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
			s.handleSSE(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 2<<20)
		data, err := io.ReadAll(r.Body)
		if err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				_ = json.NewEncoder(w).Encode(Response{
					ID:    nil,
					Error: &Error{Code: CodeInvalidRequest, Message: "Payload too large (max 2MB)"},
				})
				return
			}
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		out, err := s.HandleMessage(r.Context(), data)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if len(out) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
	}
}

// handleSSE is a stub SSE endpoint (endpoint event only, no streaming).
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	endpointURI := r.URL.Path
	_, _ = fmt.Fprintf(w, "event: endpoint\ndata: %s\n\n", endpointURI)
	flusher.Flush()

	<-r.Context().Done()
}
