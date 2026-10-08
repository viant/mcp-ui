// Package resource provides `ui://...` resource builders for the MCP UI
// extension proposed by SEP-1865 (MCP Apps).
//
// All MIME and `_meta` values are sourced byte-for-byte from the upstream
// MCP Apps spec commit pinned in the module README. Validation is explicit
// and identity-based; no heuristics or fallback guesses.
package resource

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
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

// UIURI preserves an exact ui:// resource identity. Kind and ResourceID are
// populated only for the optional scoped legacy naming convention.
type UIURI struct {
	Raw         string
	ServerScope string
	Kind        string
	ResourceID  string
}

// ValidateUIURI accepts generic official ui:// resource identities. The stable
// extension does not impose a product's resource-kind or path convention.
func ValidateUIURI(raw string) (UIURI, error) {
	if len(raw) > 4096 || !utf8.ValidString(raw) || strings.IndexFunc(raw, unicode.IsSpace) >= 0 || strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return UIURI{}, fmt.Errorf("ui uri: invalid or oversized identity")
	}
	parsed, err := url.Parse(raw)
	if err != nil || !strings.HasPrefix(raw, Scheme+"://") || parsed.Scheme != Scheme || parsed.Host == "" || parsed.User != nil {
		return UIURI{}, fmt.Errorf("ui uri: exact ui:// identity required")
	}
	result := UIURI{Raw: raw, ServerScope: parsed.Host}
	if scoped, err := ValidateScopedUIURI(raw); err == nil {
		result.Kind, result.ResourceID = scoped.Kind, scoped.ResourceID
	}
	return result, nil
}

// ValidateScopedUIURI parses the legacy, application-owned naming convention:
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
func ValidateScopedUIURI(raw string) (UIURI, error) {
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
