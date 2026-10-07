// Package meta provides typed manipulation of tool and resource `_meta`
// payloads for the MCP UI extension proposed by SEP-1865 (MCP Apps).
//
// All `_meta` key spellings are sourced byte-for-byte from the upstream MCP
// Apps spec commit pinned in the module README. Helpers in this package
// perform exact identity reads/writes only; there is no fallback guessing.
package meta

import (
	"github.com/viant/mcp-protocol/schema"
)

// Top-level `_meta` namespace for the MCP UI extension.
const (
	// NamespaceUI is the parent key under which all UI extension metadata
	// lives in a `_meta` map.
	NamespaceUI = "ui"
)

// Tool-level `_meta.ui.*` keys (joined dotted paths are documented for
// readers; the constants below are bare leaf names because the helpers
// operate on the nested map under NamespaceUI).
const (
	// FieldResourceUri is the leaf key under `_meta.ui` advertising the
	// `ui://...` URI that a tool's result can be rendered through.
	// Full path: `_meta.ui.resourceUri`.
	FieldResourceUri = "resourceUri"

	// FieldAllowedTools is the leaf key under `_meta.ui` carrying the per-
	// resource tool allowlist that guest-originated `tools/call` requests
	// must satisfy. Full path: `_meta.ui.allowedTools`.
	FieldAllowedTools = "allowedTools"

	// FieldAllowedToolBundles is the leaf key under `_meta.ui` carrying the
	// exact tool bundle ids the host must use when executing guest-originated
	// `tools/call` requests through an approval-aware server path.
	// Full path: `_meta.ui.allowedToolBundles`.
	FieldAllowedToolBundles = "allowedToolBundles"

	// FieldContentHash is the leaf key under `_meta.ui` carrying the
	// canonical content hash of a UI resource payload.
	// Full path: `_meta.ui.contentHash`.
	FieldContentHash = "contentHash"

	// FieldProtocolVersion is the leaf key under `_meta.ui` carrying the
	// `viant/mcp-ui` semantic version that produced the resource. Full
	// path: `_meta.ui.protocolVersion`.
	FieldProtocolVersion = "protocolVersion"

	// FieldRendererURL is the leaf key under `_meta.ui` carrying an explicit
	// same-origin renderer page URL the host may load instead of raw srcdoc when
	// the resource requires a richer authenticated runtime.
	FieldRendererURL = "rendererUrl"

	// FieldSandbox is the leaf key under `_meta.ui` carrying an explicit iframe
	// sandbox policy string for this resource. Empty means host default.
	FieldSandbox = "sandbox"

	// FieldCSP is the leaf key under `_meta.ui` carrying an explicit srcdoc CSP
	// policy override for browser hosts that assemble HTML via iframe.srcdoc.
	// Empty means host default deny-by-default CSP.
	FieldCSP = "csp"

	// FieldCSPPolicy is the leaf key under `_meta.ui` carrying a structured CSP
	// relaxation policy for browser hosts that assemble HTML via iframe.srcdoc.
	// Empty means host default deny-by-default CSP with no relaxations.
	FieldCSPPolicy = "cspPolicy"

	// FieldFallback is the leaf key under `_meta.ui` opting a resource or
	// tool in to embedded-resource compatibility behavior. Full path:
	// `_meta.ui.fallback`. The only currently meaningful value is the
	// constant FallbackEmbedded.
	FieldFallback = "fallback"
)

// FallbackEmbedded is the only fallback value documented by the plan today.
const FallbackEmbedded = "embedded"

// ToolUI is the typed representation of `_meta.ui` as it appears on a
// schema.Tool. Fields with their zero values are omitted on write.
type ToolUI struct {
	// Visibility is the official audience list. Nil defaults to model and app;
	// an explicitly empty slice denies both.
	Visibility []string

	// ResourceUri is the canonical `ui://...` resource advertised by the
	// tool. Required when emitting UI metadata.
	ResourceUri string

	// AllowedTools is the per-resource tool allowlist mirror that may be
	// advertised on a tool for discovery purposes. The authoritative copy
	// lives on the resource itself.
	AllowedTools []string

	// Fallback opts the tool in to embedded-resource fallback when peers
	// do not negotiate UI extension capability. Empty means no fallback.
	Fallback string
}

// ResourceUI is the typed representation of `_meta.ui` as it appears on a
// schema.Resource or on a schema.ResourceContents entry.
// ResourceCSP contains the official stable resource-owned domain declarations.
type ResourceCSP struct {
	ConnectDomains  []string `json:"connectDomains,omitempty"`
	ResourceDomains []string `json:"resourceDomains,omitempty"`
	FrameDomains    []string `json:"frameDomains,omitempty"`
	BaseUriDomains  []string `json:"baseUriDomains,omitempty"`
}

type ResourceUI struct {
	// OfficialCSP encodes _meta.ui.csp as the stable object. CSP/CSPPolicy below
	// are legacy adapter fields and must not be treated as normative metadata.
	OfficialCSP   *ResourceCSP
	Domain        string
	PrefersBorder *bool

	// AllowedTools is the authoritative per-resource tool allowlist that
	// guest-originated `tools/call` requests are validated against.
	AllowedTools []string

	// AllowedToolBundles is the exact bundle selection the host must apply
	// when executing guest-originated `tools/call` requests so approval and
	// other tool-surface metadata comes from the canonical bundle layer.
	AllowedToolBundles []string

	// ContentHash is the canonical content hash of the resource payload.
	ContentHash string

	// ProtocolVersion is the `viant/mcp-ui` semantic version that produced
	// the resource.
	ProtocolVersion string

	// RendererURL is an optional same-origin page URL that the host may load
	// instead of srcdoc when the resource needs a richer authenticated runtime.
	RendererURL string

	// Sandbox is an optional iframe sandbox string for this resource.
	Sandbox string

	// CSP is an optional explicit srcdoc CSP policy override. Empty means host
	// default deny-by-default nonce-based policy.
	CSP string

	// CSPPolicy is the preferred structured srcdoc CSP relaxation contract for
	// explicitly approved advanced widgets. Empty means no relaxation.
	CSPPolicy *CSPPolicy

	// Fallback opts the resource in to embedded-resource fallback. Empty
	// means no fallback.
	Fallback string
}

// CSPPolicy is the explicit, bounded per-resource CSP relaxation contract for
// srcdoc-rendered widgets. It is intentionally narrow: only the allowlists and
// strict-dynamic toggle the plan currently calls out are modeled here.
type CSPPolicy struct {
	ScriptSrc           []string `json:"scriptSrc,omitempty" yaml:"scriptSrc,omitempty"`
	StyleSrc            []string `json:"styleSrc,omitempty" yaml:"styleSrc,omitempty"`
	ConnectSrc          []string `json:"connectSrc,omitempty" yaml:"connectSrc,omitempty"`
	ImgSrc              []string `json:"imgSrc,omitempty" yaml:"imgSrc,omitempty"`
	FontSrc             []string `json:"fontSrc,omitempty" yaml:"fontSrc,omitempty"`
	ScriptStrictDynamic bool     `json:"scriptStrictDynamic,omitempty" yaml:"scriptStrictDynamic,omitempty"`
}

// SetToolUI writes ui under tool.Meta[NamespaceUI]. It preserves any other
// `_meta.*` entries already on the tool and any other keys under
// `_meta.ui.*` that are not modeled by ToolUI. Passing the zero ToolUI is
// equivalent to ClearToolUI and removes the `_meta.ui` entry.
//
// tool must be non-nil.
func SetToolUI(tool *schema.Tool, ui ToolUI) {
	if tool == nil {
		panic("meta: SetToolUI called with nil tool")
	}
	if isZeroToolUI(ui) {
		clearMetaUI(&tool.Meta)
		return
	}
	if tool.Meta == nil {
		tool.Meta = map[string]interface{}{}
	}
	sub := nestedMap(tool.Meta, NamespaceUI)
	writeToolUI(sub, ui)
	tool.Meta[NamespaceUI] = sub
}

// GetToolUI reads `_meta.ui` from the tool and returns the typed view plus
// true if `_meta.ui` is present. A present-but-empty `_meta.ui` returns
// ToolUI{}, true.
func GetToolUI(tool *schema.Tool) (ToolUI, bool) {
	if tool == nil {
		return ToolUI{}, false
	}
	sub, ok := readNamespace(tool.Meta)
	if !ok {
		return ToolUI{}, false
	}
	return readToolUI(sub), true
}

// SetResourceUI writes ui under resource.Meta[NamespaceUI]. It preserves
// any other `_meta.*` entries already on the resource. Passing the zero
// ResourceUI removes the `_meta.ui` entry.
//
// resource must be non-nil.
func SetResourceUI(resource *schema.Resource, ui ResourceUI) {
	if resource == nil {
		panic("meta: SetResourceUI called with nil resource")
	}
	if isZeroResourceUI(ui) {
		clearMetaUI(&resource.Meta)
		return
	}
	if resource.Meta == nil {
		resource.Meta = map[string]interface{}{}
	}
	sub := nestedMap(resource.Meta, NamespaceUI)
	writeResourceUI(sub, ui)
	resource.Meta[NamespaceUI] = sub
}

// GetResourceUI reads `_meta.ui` from the resource.
func GetResourceUI(resource *schema.Resource) (ResourceUI, bool) {
	if resource == nil {
		return ResourceUI{}, false
	}
	sub, ok := readNamespace(resource.Meta)
	if !ok {
		return ResourceUI{}, false
	}
	return readResourceUI(sub), true
}

// SetResourceContentsUI writes ui under contents.Meta[NamespaceUI].
//
// contents must be non-nil.
func SetResourceContentsUI(contents *schema.ResourceContents, ui ResourceUI) {
	if contents == nil {
		panic("meta: SetResourceContentsUI called with nil contents")
	}
	if isZeroResourceUI(ui) {
		clearMetaUI(&contents.Meta)
		return
	}
	if contents.Meta == nil {
		contents.Meta = map[string]interface{}{}
	}
	sub := nestedMap(contents.Meta, NamespaceUI)
	writeResourceUI(sub, ui)
	contents.Meta[NamespaceUI] = sub
}

// GetResourceContentsUI reads `_meta.ui` from the resource contents entry.
func GetResourceContentsUI(contents *schema.ResourceContents) (ResourceUI, bool) {
	if contents == nil {
		return ResourceUI{}, false
	}
	sub, ok := readNamespace(contents.Meta)
	if !ok {
		return ResourceUI{}, false
	}
	return readResourceUI(sub), true
}

// SetTextResourceContentsUI writes ui under contents.Meta[NamespaceUI].
func SetTextResourceContentsUI(contents *schema.TextResourceContents, ui ResourceUI) {
	if contents == nil {
		panic("meta: SetTextResourceContentsUI called with nil contents")
	}
	if isZeroResourceUI(ui) {
		clearMetaUI(&contents.Meta)
		return
	}
	if contents.Meta == nil {
		contents.Meta = map[string]interface{}{}
	}
	sub := nestedMap(contents.Meta, NamespaceUI)
	writeResourceUI(sub, ui)
	contents.Meta[NamespaceUI] = sub
}

// GetTextResourceContentsUI reads `_meta.ui` from a text resource contents
// entry.
func GetTextResourceContentsUI(contents *schema.TextResourceContents) (ResourceUI, bool) {
	if contents == nil {
		return ResourceUI{}, false
	}
	sub, ok := readNamespace(contents.Meta)
	if !ok {
		return ResourceUI{}, false
	}
	return readResourceUI(sub), true
}

// SetReadResultContentsUI writes ui under contents.Meta[NamespaceUI].
func SetReadResultContentsUI(contents *schema.ReadResourceResultContentsElem, ui ResourceUI) {
	if contents == nil {
		panic("meta: SetReadResultContentsUI called with nil contents")
	}
	if isZeroResourceUI(ui) {
		clearMetaUI(&contents.Meta)
		return
	}
	if contents.Meta == nil {
		contents.Meta = map[string]interface{}{}
	}
	sub := nestedMap(contents.Meta, NamespaceUI)
	writeResourceUI(sub, ui)
	contents.Meta[NamespaceUI] = sub
}

// GetReadResultContentsUI reads `_meta.ui` from a resources/read contents
// entry.
func GetReadResultContentsUI(contents *schema.ReadResourceResultContentsElem) (ResourceUI, bool) {
	if contents == nil {
		return ResourceUI{}, false
	}
	sub, ok := readNamespace(contents.Meta)
	if !ok {
		return ResourceUI{}, false
	}
	return readResourceUI(sub), true
}

// SetEmbeddedResourceUI writes ui under both the top-level embedded resource
// metadata and the embedded resource payload metadata.
func SetEmbeddedResourceUI(resource *schema.EmbeddedResource, ui ResourceUI) {
	if resource == nil {
		panic("meta: SetEmbeddedResourceUI called with nil resource")
	}
	if isZeroResourceUI(ui) {
		clearMetaUI(&resource.Meta)
		clearMetaUI(&resource.Resource.Meta)
		return
	}
	if resource.Meta == nil {
		resource.Meta = map[string]interface{}{}
	}
	if resource.Resource.Meta == nil {
		resource.Resource.Meta = map[string]interface{}{}
	}
	sub := nestedMap(resource.Meta, NamespaceUI)
	writeResourceUI(sub, ui)
	resource.Meta[NamespaceUI] = sub
	subResource := nestedMap(resource.Resource.Meta, NamespaceUI)
	writeResourceUI(subResource, ui)
	resource.Resource.Meta[NamespaceUI] = subResource
}

// GetEmbeddedResourceUI reads `_meta.ui` from the embedded resource payload
// metadata, falling back to the top-level embedded resource metadata only if
// the payload metadata is absent.
func GetEmbeddedResourceUI(resource *schema.EmbeddedResource) (ResourceUI, bool) {
	if resource == nil {
		return ResourceUI{}, false
	}
	if sub, ok := readNamespace(resource.Resource.Meta); ok {
		return readResourceUI(sub), true
	}
	sub, ok := readNamespace(resource.Meta)
	if !ok {
		return ResourceUI{}, false
	}
	return readResourceUI(sub), true
}

func writeToolUI(sub map[string]interface{}, ui ToolUI) {
	if ui.Visibility != nil {
		sub["visibility"] = toInterfaceSlice(ui.Visibility)
	} else {
		delete(sub, "visibility")
	}
	if ui.ResourceUri != "" {
		sub[FieldResourceUri] = ui.ResourceUri
	} else {
		delete(sub, FieldResourceUri)
	}
	if len(ui.AllowedTools) > 0 {
		sub[FieldAllowedTools] = toInterfaceSlice(ui.AllowedTools)
	} else {
		delete(sub, FieldAllowedTools)
	}
	if ui.Fallback != "" {
		sub[FieldFallback] = ui.Fallback
	} else {
		delete(sub, FieldFallback)
	}
}

func readToolUI(sub map[string]interface{}) ToolUI {
	out := ToolUI{}
	if value, ok := sub["visibility"]; ok {
		out.Visibility = readStringSlice(value)
		if out.Visibility == nil {
			out.Visibility = []string{}
		}
	}
	if v, ok := sub[FieldResourceUri].(string); ok {
		out.ResourceUri = v
	}
	out.AllowedTools = readStringSlice(sub[FieldAllowedTools])
	if v, ok := sub[FieldFallback].(string); ok {
		out.Fallback = v
	}
	return out
}

func writeResourceUI(sub map[string]interface{}, ui ResourceUI) {
	if ui.Domain != "" {
		sub["domain"] = ui.Domain
	} else {
		delete(sub, "domain")
	}
	if ui.PrefersBorder != nil {
		sub["prefersBorder"] = *ui.PrefersBorder
	} else {
		delete(sub, "prefersBorder")
	}
	if len(ui.AllowedTools) > 0 {
		sub[FieldAllowedTools] = toInterfaceSlice(ui.AllowedTools)
	} else {
		delete(sub, FieldAllowedTools)
	}
	if len(ui.AllowedToolBundles) > 0 {
		sub[FieldAllowedToolBundles] = toInterfaceSlice(ui.AllowedToolBundles)
	} else {
		delete(sub, FieldAllowedToolBundles)
	}
	if ui.ContentHash != "" {
		sub[FieldContentHash] = ui.ContentHash
	} else {
		delete(sub, FieldContentHash)
	}
	if ui.ProtocolVersion != "" {
		sub[FieldProtocolVersion] = ui.ProtocolVersion
	} else {
		delete(sub, FieldProtocolVersion)
	}
	if ui.RendererURL != "" {
		sub[FieldRendererURL] = ui.RendererURL
	} else {
		delete(sub, FieldRendererURL)
	}
	if ui.Sandbox != "" {
		sub[FieldSandbox] = ui.Sandbox
	} else {
		delete(sub, FieldSandbox)
	}
	if ui.CSPPolicy != nil && !isZeroCSPPolicy(ui.CSPPolicy) {
		sub[FieldCSPPolicy] = writeCSPPolicy(ui.CSPPolicy)
	} else {
		delete(sub, FieldCSPPolicy)
	}
	if ui.OfficialCSP != nil {
		sub[FieldCSP] = map[string]interface{}{
			"connectDomains":  toInterfaceSlice(ui.OfficialCSP.ConnectDomains),
			"resourceDomains": toInterfaceSlice(ui.OfficialCSP.ResourceDomains),
			"frameDomains":    toInterfaceSlice(ui.OfficialCSP.FrameDomains),
			"baseUriDomains":  toInterfaceSlice(ui.OfficialCSP.BaseUriDomains),
		}
	} else if ui.CSP != "" {
		sub[FieldCSP] = ui.CSP
	} else {
		delete(sub, FieldCSP)
	}
	if ui.Fallback != "" {
		sub[FieldFallback] = ui.Fallback
	} else {
		delete(sub, FieldFallback)
	}
}

func readResourceUI(sub map[string]interface{}) ResourceUI {
	out := ResourceUI{}
	out.Domain, _ = sub["domain"].(string)
	if value, ok := sub["prefersBorder"].(bool); ok {
		out.PrefersBorder = &value
	}
	if csp, ok := sub[FieldCSP].(map[string]interface{}); ok {
		out.OfficialCSP = &ResourceCSP{ConnectDomains: readStringSlice(csp["connectDomains"]), ResourceDomains: readStringSlice(csp["resourceDomains"]), FrameDomains: readStringSlice(csp["frameDomains"]), BaseUriDomains: readStringSlice(csp["baseUriDomains"])}
	}
	out.AllowedTools = readStringSlice(sub[FieldAllowedTools])
	out.AllowedToolBundles = readStringSlice(sub[FieldAllowedToolBundles])
	if v, ok := sub[FieldContentHash].(string); ok {
		out.ContentHash = v
	}
	if v, ok := sub[FieldProtocolVersion].(string); ok {
		out.ProtocolVersion = v
	}
	if v, ok := sub[FieldRendererURL].(string); ok {
		out.RendererURL = v
	}
	if v, ok := sub[FieldSandbox].(string); ok {
		out.Sandbox = v
	}
	if policy, ok := readCSPPolicy(sub[FieldCSPPolicy]); ok {
		out.CSPPolicy = policy
	}
	if v, ok := sub[FieldCSP].(string); ok {
		out.CSP = v
	}
	if v, ok := sub[FieldFallback].(string); ok {
		out.Fallback = v
	}
	return out
}

func isZeroToolUI(ui ToolUI) bool {
	return ui.ResourceUri == "" && ui.Fallback == "" && len(ui.AllowedTools) == 0 && ui.Visibility == nil
}

func isZeroResourceUI(ui ResourceUI) bool {
	return ui.OfficialCSP == nil && ui.Domain == "" && ui.PrefersBorder == nil && ui.ContentHash == "" && ui.ProtocolVersion == "" && ui.RendererURL == "" && ui.Sandbox == "" && ui.CSP == "" && (ui.CSPPolicy == nil || isZeroCSPPolicy(ui.CSPPolicy)) && ui.Fallback == "" && len(ui.AllowedTools) == 0 && len(ui.AllowedToolBundles) == 0
}

func writeCSPPolicy(policy *CSPPolicy) map[string]interface{} {
	out := map[string]interface{}{}
	if policy == nil {
		return out
	}
	if len(policy.ScriptSrc) > 0 {
		out["scriptSrc"] = toInterfaceSlice(policy.ScriptSrc)
	}
	if len(policy.StyleSrc) > 0 {
		out["styleSrc"] = toInterfaceSlice(policy.StyleSrc)
	}
	if len(policy.ConnectSrc) > 0 {
		out["connectSrc"] = toInterfaceSlice(policy.ConnectSrc)
	}
	if len(policy.ImgSrc) > 0 {
		out["imgSrc"] = toInterfaceSlice(policy.ImgSrc)
	}
	if len(policy.FontSrc) > 0 {
		out["fontSrc"] = toInterfaceSlice(policy.FontSrc)
	}
	if policy.ScriptStrictDynamic {
		out["scriptStrictDynamic"] = true
	}
	return out
}

func readCSPPolicy(raw interface{}) (*CSPPolicy, bool) {
	sub, ok := raw.(map[string]interface{})
	if !ok {
		return nil, false
	}
	policy := &CSPPolicy{
		ScriptSrc:           readStringSlice(sub["scriptSrc"]),
		StyleSrc:            readStringSlice(sub["styleSrc"]),
		ConnectSrc:          readStringSlice(sub["connectSrc"]),
		ImgSrc:              readStringSlice(sub["imgSrc"]),
		FontSrc:             readStringSlice(sub["fontSrc"]),
		ScriptStrictDynamic: readBool(sub["scriptStrictDynamic"]),
	}
	if isZeroCSPPolicy(policy) {
		return nil, true
	}
	return policy, true
}

func isZeroCSPPolicy(policy *CSPPolicy) bool {
	if policy == nil {
		return true
	}
	return len(policy.ScriptSrc) == 0 &&
		len(policy.StyleSrc) == 0 &&
		len(policy.ConnectSrc) == 0 &&
		len(policy.ImgSrc) == 0 &&
		len(policy.FontSrc) == 0 &&
		!policy.ScriptStrictDynamic
}

func nestedMap(m map[string]interface{}, key string) map[string]interface{} {
	if existing, ok := m[key].(map[string]interface{}); ok {
		return existing
	}
	return map[string]interface{}{}
}

func readNamespace(m map[string]interface{}) (map[string]interface{}, bool) {
	if m == nil {
		return nil, false
	}
	sub, ok := m[NamespaceUI].(map[string]interface{})
	if !ok {
		return nil, false
	}
	return sub, true
}

func clearMetaUI(meta *map[string]interface{}) {
	if meta == nil || *meta == nil {
		return
	}
	delete(*meta, NamespaceUI)
	if len(*meta) == 0 {
		*meta = nil
	}
}

func toInterfaceSlice(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i, v := range in {
		out[i] = v
	}
	return out
}

func readStringSlice(raw interface{}) []string {
	if raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		out := make([]string, len(v))
		copy(out, v)
		return out
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func readBool(raw interface{}) bool {
	value, ok := raw.(bool)
	return ok && value
}
