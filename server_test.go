package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/Natykufsky/verity-mcp"
)

func TestMCPLibrary_Initialize(t *testing.T) {
	srv := mcp.NewServer(mcp.ServerConfig{
		Name:    "test-server",
		Version: "1.2.3",
	})

	req := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`
	out, err := srv.HandleMessage(context.Background(), []byte(req))
	if err != nil {
		t.Fatalf("HandleMessage failed: %v", err)
	}

	var resp struct {
		Result mcp.InitializeResult `json:"result"`
	}
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.Result.ServerInfo.Name != "test-server" || resp.Result.ServerInfo.Version != "1.2.3" {
		t.Errorf("unexpected server info: %+v", resp.Result.ServerInfo)
	}
}

func TestMCPLibrary_Tools(t *testing.T) {
	srv := mcp.NewServer(mcp.ServerConfig{
		CheckPermission: func(ctx context.Context, perm string) bool {
			return perm == "allowed.perm"
		},
	})

	srv.RegisterTool(mcp.ToolDefinition{
		Name:        "allowed_tool",
		Description: "Allowed tool",
		Permission:  "allowed.perm",
		Handler: func(ctx context.Context, args json.RawMessage) ([]mcp.ContentItem, error) {
			return []mcp.ContentItem{{Type: "text", Text: "success"}}, nil
		},
	})

	srv.RegisterTool(mcp.ToolDefinition{
		Name:        "forbidden_tool",
		Description: "Forbidden tool",
		Permission:  "forbidden.perm",
	})

	ctx := context.Background()

	// 1. List tools
	req := `{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`
	out, err := srv.HandleMessage(ctx, []byte(req))
	if err != nil {
		t.Fatalf("tools/list error: %v", err)
	}

	var listResp struct {
		Result mcp.ToolsListResult `json:"result"`
	}
	_ = json.Unmarshal(out, &listResp)
	if len(listResp.Result.Tools) != 1 || listResp.Result.Tools[0].Name != "allowed_tool" {
		t.Errorf("tools/list filter failed, got: %+v", listResp.Result.Tools)
	}

	// 2. Call tool
	callReq := `{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"allowed_tool","arguments":{}}}`
	callOut, err := srv.HandleMessage(ctx, []byte(callReq))
	if err != nil {
		t.Fatalf("tools/call error: %v", err)
	}

	var callResp struct {
		Result mcp.ToolCallResult `json:"result"`
	}
	_ = json.Unmarshal(callOut, &callResp)
	if len(callResp.Result.Content) == 0 || callResp.Result.Content[0].Text != "success" {
		t.Errorf("tools/call failed, got: %+v", callResp.Result)
	}
}

func TestMCPLibrary_Prompts(t *testing.T) {
	srv := mcp.NewServer(mcp.ServerConfig{})

	srv.RegisterPrompt(mcp.PromptDefinition{
		Name:        "audit_review",
		Description: "Performs a financial audit review",
		Arguments: []mcp.PromptArgument{
			{Name: "period", Description: "Accounting period", Required: true},
		},
		Handler: func(ctx context.Context, args map[string]string) ([]mcp.PromptMessage, error) {
			period := args["period"]
			return []mcp.PromptMessage{
				{
					Role: "user",
					Content: mcp.ContentItem{
						Type: "text",
						Text: "Please analyze all trial balance lines for period " + period,
					},
				},
			}, nil
		},
	})

	ctx := context.Background()

	// 1. Prompts list
	req := `{"jsonrpc":"2.0","id":1,"method":"prompts/list","params":{}}`
	out, err := srv.HandleMessage(ctx, []byte(req))
	if err != nil {
		t.Fatalf("prompts/list error: %v", err)
	}

	var listResp struct {
		Result mcp.PromptsListResult `json:"result"`
	}
	_ = json.Unmarshal(out, &listResp)
	if len(listResp.Result.Prompts) != 1 || listResp.Result.Prompts[0].Name != "audit_review" {
		t.Errorf("prompts/list unexpected: %+v", listResp.Result.Prompts)
	}

	// 2. Prompts get
	getReq := `{"jsonrpc":"2.0","id":2,"method":"prompts/get","params":{"name":"audit_review","arguments":{"period":"2026-Q4"}}}`
	getOut, err := srv.HandleMessage(ctx, []byte(getReq))
	if err != nil {
		t.Fatalf("prompts/get error: %v", err)
	}

	var getResp struct {
		Result mcp.PromptGetResult `json:"result"`
	}
	_ = json.Unmarshal(getOut, &getResp)
	if len(getResp.Result.Messages) == 0 || getResp.Result.Messages[0].Content.Text != "Please analyze all trial balance lines for period 2026-Q4" {
		t.Errorf("prompts/get unexpected: %+v", getResp.Result)
	}
}

func TestMCPLibrary_Stdio(t *testing.T) {
	srv := mcp.NewServer(mcp.ServerConfig{
		Name: "stdio-server",
	})

	in := bytes.NewBufferString(`{"jsonrpc":"2.0","id":100,"method":"ping","params":{}}` + "\n")
	var out bytes.Buffer

	if err := srv.ServeStdio(context.Background(), in, &out); err != nil {
		t.Fatalf("ServeStdio error: %v", err)
	}

	var resp struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("output json parse error: %v", err)
	}
	if resp.ID != 100 {
		t.Errorf("unexpected id %d", resp.ID)
	}
}
