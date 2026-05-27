// Package resource provides `ui://...` resource builders for the MCP UI
// extension proposed by SEP-1865 (MCP Apps).
//
// All MIME and `_meta` values are sourced byte-for-byte from the upstream
// MCP Apps spec commit pinned in the module README. Validation is explicit
// and identity-based; no heuristics or fallback guesses.
package resource

import (
	"fmt"
	"strings"
)

// Scheme is the URI scheme of all MCP UI resources.
const Scheme = "ui"

// Resource-kind enumeration recognized by Phase 1.
const (
	KindDemo   = "demo"
	KindView   = "view"
	KindFeed   = "feed"
	KindWidget = "widget"
)

// allowedKinds lists the resource kinds accepted by ValidateUIURI in Phase 1.
var allowedKinds = map[string]struct{}{
	KindDemo:   {},
	KindView:   {},
	KindFeed:   {},
	KindWidget: {},
}

// UIURI is the parsed form of a `ui://<server-scope>/<resource-kind>/<resource-id>`
// URI. The original textual form is preserved as Raw.
type UIURI struct {
	Raw         string
	ServerScope string
	Kind        string
	ResourceID  string
}

// ValidateUIURI parses and validates a `ui://...` resource URI according to
// the grammar documented in the enhancement plan:
//
//	ui://<server-scope>/<resource-kind>/<resource-id>
//
// Rules enforced:
//
//   - scheme MUST be `ui`
//   - exactly three non-empty path segments
//   - resource-kind MUST be one of demo/view/feed/widget
//   - no query string and no fragment are allowed
//   - server-scope and resource-id MUST be non-empty
//
// Validation is exact; there is no fuzzy matching and no fallback guessing.
func ValidateUIURI(raw string) (UIURI, error) {
	if raw == "" {
		return UIURI{}, fmt.Errorf("ui uri: empty")
	}
	const prefix = Scheme + "://"
	if !strings.HasPrefix(raw, prefix) {
		return UIURI{}, fmt.Errorf("ui uri: scheme must be %q, got %q", Scheme, raw)
	}
	if strings.ContainsAny(raw, "?#") {
		return UIURI{}, fmt.Errorf("ui uri: query/fragment not allowed: %q", raw)
	}
	rest := raw[len(prefix):]
	parts := strings.Split(rest, "/")
	if len(parts) != 3 {
		return UIURI{}, fmt.Errorf("ui uri: expected ui://<server-scope>/<kind>/<resource-id>, got %q", raw)
	}
	serverScope, kind, resourceID := parts[0], parts[1], parts[2]
	if serverScope == "" {
		return UIURI{}, fmt.Errorf("ui uri: server-scope must be non-empty: %q", raw)
	}
	if resourceID == "" {
		return UIURI{}, fmt.Errorf("ui uri: resource-id must be non-empty: %q", raw)
	}
	if _, ok := allowedKinds[kind]; !ok {
		return UIURI{}, fmt.Errorf("ui uri: resource-kind %q not in {demo,view,feed,widget}", kind)
	}
	return UIURI{
		Raw:         raw,
		ServerScope: serverScope,
		Kind:        kind,
		ResourceID:  resourceID,
	}, nil
}

// MustValidateUIURI is the panicking variant of ValidateUIURI for use in
// tests and package-level initializers that must be statically valid.
func MustValidateUIURI(raw string) UIURI {
	u, err := ValidateUIURI(raw)
	if err != nil {
		panic(err)
	}
	return u
}
