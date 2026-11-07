package mcp

import (
	"context"
)

// ResourceDefinition models a read-only document, schema, or system state.
type ResourceDefinition struct {
	URI         string                                                              `json:"uri"`
	Name        string                                                              `json:"name"`
	Description string                                                              `json:"description"`
	MimeType    string                                                              `json:"mimeType"`
	Permission  string                                                              `json:"permission,omitempty"`
	Reader      func(ctx context.Context, uri string) ([]ResourceContentItem, error) `json:"-"`
}

// ResourceItem is the serialized schema representation for resources/list.
type ResourceItem struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// ResourcesListResult is returned by resources/list.
type ResourcesListResult struct {
	Resources []ResourceItem `json:"resources"`
}

// ResourceReadParams defines parameters received in resources/read.
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// ResourceContentItem is an individual piece of content within a resource read result.
type ResourceContentItem struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ResourceReadResult is returned by resources/read.
type ResourceReadResult struct {
	Contents []ResourceContentItem `json:"contents"`
}
