package mcp

import (
	"context"
)

// PromptArgument defines an input argument for a prompt template.
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptDefinition defines an interactive prompt workflow exposed over MCP.
type PromptDefinition struct {
	Name        string                                                                    `json:"name"`
	Description string                                                                    `json:"description"`
	Arguments   []PromptArgument                                                          `json:"arguments,omitempty"`
	Permission  string                                                                    `json:"permission,omitempty"`
	Handler     func(ctx context.Context, args map[string]string) ([]PromptMessage, error) `json:"-"`
}

// PromptMessage models a message returned by prompts/get.
type PromptMessage struct {
	Role    string      `json:"role"`
	Content ContentItem `json:"content"`
}

// PromptItem is the schema representation returned by prompts/list.
type PromptItem struct {
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Arguments   []PromptArgument `json:"arguments,omitempty"`
}

// PromptsListResult is returned by prompts/list.
type PromptsListResult struct {
	Prompts []PromptItem `json:"prompts"`
}

// PromptGetParams defines parameters for prompts/get.
type PromptGetParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments,omitempty"`
}

// PromptGetResult is returned by prompts/get.
type PromptGetResult struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}
