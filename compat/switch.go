package compat

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-ui/capabilities"
	"github.com/viant/mcp-ui/meta"
	"github.com/viant/mcp-ui/resource"
)

// Mode is the rendering mode a server should apply to a UI-bearing tool
// result, given the negotiated UI extension capability and the tool's or
// resource's explicit fallback opt-in.
type Mode int

const (
	// ModeNone means no UI rendering applies. Either the tool advertises no
	// UI resource at all, or UI capability is absent and the tool/resource
	// has not opted in to embedded-resource fallback.
	ModeNone Mode = iota

	// ModeCapability means UI extension capability is negotiated and the
	// host follows the `_meta.ui.resourceUri` + `resources/read` flow on its
	// own. The server should NOT emit an embedded UI resource.
	ModeCapability

	// ModeEmbeddedFallback means UI extension capability is NOT negotiated
	// but the tool or resource has explicitly opted in via
	// `_meta.ui.fallback = "embedded"`. The server MUST emit a UI
	// EmbeddedResource in the tool result so a non-UI-capable host can still
	// render it.
	ModeEmbeddedFallback
)

// String reports a stable identifier for the mode, useful in logs and tests.
func (m Mode) String() string {
	switch m {
	case ModeCapability:
		return "capability"
	case ModeEmbeddedFallback:
		return "embedded-fallback"
	default:
		return "none"
	}
}

// ClientUICapable reports whether the supplied client capabilities advertise
// the supported HTML MIME under Extensions[capabilities.ExtensionName]. nil caps
// are treated as not capable.
func ClientUICapable(caps *schema.ClientCapabilities) bool {
	capability, ok := capabilities.GetClientCapability(caps)
	return ok && supportsHTML(capability)
}

// ServerUICapable reports whether the supplied server capabilities advertise
// the supported HTML MIME under Extensions[capabilities.ExtensionName].
func ServerUICapable(caps *schema.ServerCapabilities) bool {
	capability, ok := capabilities.GetServerCapability(caps)
	return ok && supportsHTML(capability)
}

func supportsHTML(capability capabilities.Capability) bool {
	for _, mime := range capability.MimeTypes {
		if mime == capabilities.ResourceMimeType {
			return true
		}
	}
	return false
}

// ToolOptsInToEmbeddedFallback reports whether the tool's `_meta.ui.fallback`
// equals `"embedded"`. No other values are recognized.
func ToolOptsInToEmbeddedFallback(ui meta.ToolUI) bool {
	return ui.Fallback == meta.FallbackEmbedded
}

// ResourceOptsInToEmbeddedFallback reports whether the resource's
// `_meta.ui.fallback` equals `"embedded"`.
func ResourceOptsInToEmbeddedFallback(ui meta.ResourceUI) bool {
	return ui.Fallback == meta.FallbackEmbedded
}

// SelectToolMode returns the rendering mode a server should apply to a tool
// result. The decision is exact:
//
//   - if the tool advertises no `_meta.ui.resourceUri` => ModeNone.
//   - if UI extension capability is negotiated => ModeCapability.
//   - else if the tool opts in via `_meta.ui.fallback = "embedded"` =>
//     ModeEmbeddedFallback.
//   - otherwise => ModeNone.
func SelectToolMode(clientCaps *schema.ClientCapabilities, toolUI meta.ToolUI) Mode {
	if toolUI.ResourceUri == "" {
		return ModeNone
	}
	if ClientUICapable(clientCaps) {
		return ModeCapability
	}
	if ToolOptsInToEmbeddedFallback(toolUI) {
		return ModeEmbeddedFallback
	}
	return ModeNone
}

// SelectResourceMode is the resource-side analogue used when the opt-in is
// declared on the resource itself rather than on the tool that advertises it.
// The decision is exact in the same way as SelectToolMode.
func SelectResourceMode(clientCaps *schema.ClientCapabilities, resourceUI meta.ResourceUI) Mode {
	if ClientUICapable(clientCaps) {
		return ModeCapability
	}
	if ResourceOptsInToEmbeddedFallback(resourceUI) {
		return ModeEmbeddedFallback
	}
	return ModeNone
}

// ResourceReader is the narrow interface a host runtime implements so the
// compat helpers can fetch the canonical `ui://...` payload without coupling
// this package to any specific server implementation.
//
// Implementations should resolve uri exactly; partial matches are not
// permitted.
type ResourceReader interface {
	ReadResource(ctx context.Context, uri string) (*schema.ReadResourceResult, error)
}

// BuildEmbeddedFallback fetches the canonical resource by uri via reader and
// wraps the first text contents entry as a UI EmbeddedResource suitable for
// appending to CallToolResult.Content.
//
// The helper preserves the canonical `_meta.ui` (contentHash, protocolVersion,
// etc.) of the resource read result rather than re-assembling HTML. The MIME
// type is taken from the resource contents when present; otherwise it
// defaults to capabilities.ResourceMimeType.
//
// Returns an error if reader is nil, the read returns no text contents, or
// the read itself fails.
func BuildEmbeddedFallback(ctx context.Context, reader ResourceReader, uri string) (*schema.EmbeddedResource, error) {
	if reader == nil {
		return nil, errors.New("compat: nil resource reader")
	}
	if _, err := resource.ValidateUIURI(uri); err != nil {
		return nil, err
	}
	result, err := reader.ReadResource(ctx, uri)
	if err != nil {
		return nil, err
	}
	if result == nil || len(result.Contents) == 0 {
		return nil, fmt.Errorf("compat: no contents for resource %q", uri)
	}
	var matches []schema.ReadResourceResultContentsElem
	for _, contents := range result.Contents {
		if contents.Uri == uri {
			matches = append(matches, contents)
		}
	}
	if len(matches) != 1 {
		return nil, errors.New("compat: missing or ambiguous resource identity")
	}
	first := matches[0]
	if first.MimeType != nil && *first.MimeType != "" && *first.MimeType != capabilities.ResourceMimeType {
		return nil, errors.New("compat: unsupported resource MIME type")
	}
	html := first.Text
	if html == "" && first.Blob != "" {
		if len(first.Blob) > base64.StdEncoding.EncodedLen(resource.MaxHTMLBytes) {
			return nil, errors.New("compat: resource exceeds budget")
		}
		decoded, err := base64.StdEncoding.DecodeString(first.Blob)
		if err != nil || !utf8.Valid(decoded) {
			return nil, errors.New("compat: invalid HTML blob")
		}
		html = string(decoded)
	}
	if html == "" || len(html) > resource.MaxHTMLBytes {
		return nil, fmt.Errorf("compat: empty text contents for resource %q", uri)
	}
	mimeType := capabilities.ResourceMimeType
	if first.MimeType != nil && *first.MimeType != "" {
		mimeType = *first.MimeType
	}
	embedded := &schema.EmbeddedResource{
		Type: "resource",
		Resource: schema.EmbeddedResourceResource{
			Uri:      first.Uri,
			MimeType: &mimeType,
			Text:     html,
		},
	}
	if ui, ok := meta.GetReadResultContentsUI(&first); ok {
		meta.SetEmbeddedResourceUI(embedded, ui)
	}
	return embedded, nil
}
