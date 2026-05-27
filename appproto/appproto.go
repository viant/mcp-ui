// Package appproto defines the local contract-of-record browser host/guest
// postMessage envelopes for the MCP UI extension layer.
//
// These message types are intentionally distinct from MCP JSON-RPC method names
// between MCP clients and MCP servers.
package appproto

// Version is the current local appproto envelope version.
const Version = "1.0.0"

const (
	MethodToolsCall        = "mcpui:tools-call"
	MethodMessage          = "mcpui:message"
	MethodOpenLink         = "mcpui:open-link"
	MethodHostReady        = "mcpui:host-ready"
	MethodToolInput        = "mcpui:tool-input"
	MethodToolInputPartial = "mcpui:tool-input-partial"
	MethodToolResult       = "mcpui:tool-result"
	MethodTeardown         = "mcpui:teardown"
	MethodSizeChanged      = "mcpui:size-changed"
)

type Envelope[T any] struct {
	Version string `json:"version"`
	Method  string `json:"method"`
	Params  T      `json:"params"`
}

type ToolsCallParams struct {
	WindowID    string                 `json:"windowId"`
	ResourceURI string                 `json:"resourceUri"`
	Name        string                 `json:"name"`
	Arguments   map[string]interface{} `json:"arguments,omitempty"`
}

type MessageParams struct {
	WindowID    string                 `json:"windowId"`
	ResourceURI string                 `json:"resourceUri"`
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type OpenLinkParams struct {
	WindowID    string `json:"windowId"`
	ResourceURI string `json:"resourceUri"`
	URL         string `json:"url"`
}

type HostReadyParams struct {
	WindowID        string   `json:"windowId"`
	ResourceURI     string   `json:"resourceUri"`
	AllowedTools    []string `json:"allowedTools,omitempty"`
	ProtocolVersion string   `json:"protocolVersion"`
}

type ToolInputParams struct {
	WindowID        string                 `json:"windowId"`
	ResourceURI     string                 `json:"resourceUri"`
	ToolName        string                 `json:"toolName"`
	ToolInput       map[string]interface{} `json:"toolInput,omitempty"`
	ProtocolVersion string                 `json:"protocolVersion"`
}

type ToolInputPartialParams struct {
	WindowID        string                 `json:"windowId"`
	ResourceURI     string                 `json:"resourceUri"`
	ToolName        string                 `json:"toolName"`
	ToolInput       map[string]interface{} `json:"toolInput,omitempty"`
	ProtocolVersion string                 `json:"protocolVersion"`
}

type ToolResultParams struct {
	WindowID          string                 `json:"windowId"`
	ResourceURI       string                 `json:"resourceUri"`
	ToolName          string                 `json:"toolName"`
	Content           interface{}            `json:"content,omitempty"`
	StructuredContent map[string]interface{} `json:"structuredContent,omitempty"`
	Meta              map[string]interface{} `json:"_meta,omitempty"`
	ProtocolVersion   string                 `json:"protocolVersion"`
}

type TeardownParams struct {
	WindowID    string `json:"windowId"`
	ResourceURI string `json:"resourceUri"`
	Reason      string `json:"reason,omitempty"`
}

type SizeChangedParams struct {
	WindowID    string `json:"windowId"`
	ResourceURI string `json:"resourceUri"`
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
}

func NewEnvelope[T any](method string, params T) Envelope[T] {
	return Envelope[T]{
		Version: Version,
		Method:  method,
		Params:  params,
	}
}
