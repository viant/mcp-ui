package meta_test

import (
	"reflect"
	"testing"

	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-ui/meta"
)

func TestKeyConstants(t *testing.T) {
	type kv struct{ k, v string }
	cases := []kv{
		{"NamespaceUI", meta.NamespaceUI},
		{"FieldResourceUri", meta.FieldResourceUri},
		{"FieldAllowedTools", meta.FieldAllowedTools},
		{"FieldAllowedToolBundles", meta.FieldAllowedToolBundles},
		{"FieldContentHash", meta.FieldContentHash},
		{"FieldProtocolVersion", meta.FieldProtocolVersion},
		{"FieldRendererURL", meta.FieldRendererURL},
		{"FieldSandbox", meta.FieldSandbox},
		{"FieldCSPPolicy", meta.FieldCSPPolicy},
		{"FieldCSP", meta.FieldCSP},
		{"FieldFallback", meta.FieldFallback},
	}
	expected := map[string]string{
		"NamespaceUI":             "ui",
		"FieldResourceUri":        "resourceUri",
		"FieldAllowedTools":       "allowedTools",
		"FieldAllowedToolBundles": "allowedToolBundles",
		"FieldContentHash":        "contentHash",
		"FieldProtocolVersion":    "protocolVersion",
		"FieldRendererURL":        "rendererUrl",
		"FieldSandbox":            "sandbox",
		"FieldCSPPolicy":          "cspPolicy",
		"FieldCSP":                "csp",
		"FieldFallback":           "fallback",
	}
	for _, c := range cases {
		if expected[c.k] != c.v {
			t.Fatalf("constant %s drifted: got %q want %q", c.k, c.v, expected[c.k])
		}
	}
	if meta.FallbackEmbedded != "embedded" {
		t.Fatalf("FallbackEmbedded drifted: %q", meta.FallbackEmbedded)
	}
}

func TestToolUI_RoundTrip(t *testing.T) {
	tool := &schema.Tool{Name: "demo.show_widget"}

	want := meta.ToolUI{
		ResourceUri:  "ui://agently.wkabc/demo/show_widget",
		AllowedTools: []string{"chat.reply", "ui/view"},
		Fallback:     meta.FallbackEmbedded,
	}
	meta.SetToolUI(tool, want)

	got, ok := meta.GetToolUI(tool)
	if !ok {
		t.Fatalf("expected _meta.ui to be present")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch: got %#v want %#v", got, want)
	}
}

func TestToolUI_PreservesOtherMetaKeys(t *testing.T) {
	tool := &schema.Tool{
		Name: "demo.show_widget",
		Meta: map[string]interface{}{
			"vendor.other/feature": map[string]interface{}{"k": "v"},
			"ui":                   map[string]interface{}{"unrelatedFutureField": "stay"},
		},
	}
	meta.SetToolUI(tool, meta.ToolUI{ResourceUri: "ui://x/demo/y"})

	other, ok := tool.Meta["vendor.other/feature"].(map[string]interface{})
	if !ok || other["k"] != "v" {
		t.Fatalf("unrelated _meta entry was mutated: %#v", tool.Meta)
	}
	ui := tool.Meta["ui"].(map[string]interface{})
	if ui["unrelatedFutureField"] != "stay" {
		t.Fatalf("unrelated _meta.ui subkey was lost: %#v", ui)
	}
	if ui[meta.FieldResourceUri] != "ui://x/demo/y" {
		t.Fatalf("resourceUri not written: %#v", ui)
	}
}

func TestToolUI_EmptyClearsNamespace(t *testing.T) {
	tool := &schema.Tool{
		Name: "demo",
		Meta: map[string]interface{}{
			"ui": map[string]interface{}{meta.FieldResourceUri: "ui://x/demo/y"},
		},
	}
	meta.SetToolUI(tool, meta.ToolUI{})
	if _, present := tool.Meta["ui"]; present {
		t.Fatalf("expected _meta.ui to be removed, got %#v", tool.Meta)
	}
}

func TestGetToolUI_AbsentReturnsFalse(t *testing.T) {
	if _, ok := meta.GetToolUI(nil); ok {
		t.Fatal("nil tool must return ok=false")
	}
	if _, ok := meta.GetToolUI(&schema.Tool{Name: "x"}); ok {
		t.Fatal("missing meta must return ok=false")
	}
	if _, ok := meta.GetToolUI(&schema.Tool{Name: "x", Meta: map[string]interface{}{"other": 1}}); ok {
		t.Fatal("missing _meta.ui must return ok=false")
	}
}

func TestResourceUI_RoundTrip(t *testing.T) {
	res := &schema.Resource{Name: "demo/show_widget", Uri: "ui://agently.wkabc/demo/show_widget"}

	want := meta.ResourceUI{
		AllowedTools:       []string{"chat.reply"},
		AllowedToolBundles: []string{"mcp_ui_preview_queue"},
		ContentHash:        "sha256:abc",
		ProtocolVersion:    "1.0.0",
		RendererURL:        "/mcp-ui-forge-window.html?windowKey=order",
		Sandbox:            "allow-scripts allow-same-origin",
		CSPPolicy: &meta.CSPPolicy{
			ScriptSrc:           []string{"https://cdn.example"},
			ConnectSrc:          []string{"https://api.example"},
			ScriptStrictDynamic: true,
		},
		CSP:      "default-src 'none'; script-src 'nonce-{{nonce}}'",
		Fallback: "",
	}
	meta.SetResourceUI(res, want)

	got, ok := meta.GetResourceUI(res)
	if !ok {
		t.Fatalf("expected _meta.ui to be present")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch: got %#v want %#v", got, want)
	}
}

func TestResourceContentsUI_RoundTrip(t *testing.T) {
	contents := &schema.ResourceContents{Uri: "ui://agently.wkabc/demo/show_widget"}

	want := meta.ResourceUI{
		AllowedTools:       []string{"chat.reply", "ui/view"},
		AllowedToolBundles: []string{"ui_host_bundle"},
		ContentHash:        "sha256:deadbeef",
		ProtocolVersion:    "1.0.0",
		RendererURL:        "/mcp-ui-forge-window.html?windowKey=order",
		Sandbox:            "allow-scripts allow-same-origin",
		CSPPolicy: &meta.CSPPolicy{
			ScriptSrc:           []string{"https://cdn.example"},
			ConnectSrc:          []string{"https://api.example"},
			ScriptStrictDynamic: true,
		},
		CSP: "default-src 'none'; script-src 'nonce-{{nonce}}'",
	}
	meta.SetResourceContentsUI(contents, want)

	got, ok := meta.GetResourceContentsUI(contents)
	if !ok {
		t.Fatalf("expected _meta.ui to be present")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch: got %#v want %#v", got, want)
	}
}

func TestResourceUI_PreservesOtherMetaKeys(t *testing.T) {
	res := &schema.Resource{
		Name: "x",
		Uri:  "ui://x/demo/y",
		Meta: map[string]interface{}{
			"vendor.other/feature": "keep",
			"ui":                   map[string]interface{}{"unrelated": 1},
		},
	}
	meta.SetResourceUI(res, meta.ResourceUI{ContentHash: "sha256:1"})
	if res.Meta["vendor.other/feature"] != "keep" {
		t.Fatalf("unrelated _meta entry was mutated: %#v", res.Meta)
	}
	ui := res.Meta["ui"].(map[string]interface{})
	if ui["unrelated"] != 1 {
		t.Fatalf("unrelated _meta.ui subkey was lost: %#v", ui)
	}
	if ui[meta.FieldContentHash] != "sha256:1" {
		t.Fatalf("contentHash not written: %#v", ui)
	}
}

func TestResourceUI_EmptyClearsNamespace(t *testing.T) {
	res := &schema.Resource{
		Name: "x",
		Uri:  "ui://x/demo/y",
		Meta: map[string]interface{}{
			"ui": map[string]interface{}{meta.FieldContentHash: "sha256:1"},
		},
	}
	meta.SetResourceUI(res, meta.ResourceUI{})
	if _, present := res.Meta["ui"]; present {
		t.Fatalf("expected _meta.ui to be removed, got %#v", res.Meta)
	}
}

func TestResourceUI_CSPOverrideIsExplicitOnly(t *testing.T) {
	// Setting a non-empty CSP must write exactly that string under
	// `_meta.ui.csp`. Setting an empty CSP must NOT write a `csp` key at all.
	res := &schema.Resource{Name: "demo/show_widget", Uri: "ui://agently.wkabc/demo/show_widget"}
	meta.SetResourceUI(res, meta.ResourceUI{
		ContentHash: "sha256:1",
		CSP:         "default-src 'none'; script-src 'nonce-{nonce}'",
	})
	ui := res.Meta["ui"].(map[string]interface{})
	if ui[meta.FieldCSP] != "default-src 'none'; script-src 'nonce-{nonce}'" {
		t.Fatalf("explicit csp not written verbatim: %#v", ui[meta.FieldCSP])
	}

	// Reset and confirm no csp key when CSP is empty.
	res.Meta = nil
	meta.SetResourceUI(res, meta.ResourceUI{ContentHash: "sha256:1"})
	ui2 := res.Meta["ui"].(map[string]interface{})
	if _, present := ui2[meta.FieldCSP]; present {
		t.Fatalf("csp key must be absent when no override is requested, got %#v", ui2)
	}

	// Reading back from a freshly built sub-map must return CSP="" when absent.
	res.Meta = map[string]interface{}{
		"ui": map[string]interface{}{
			meta.FieldContentHash: "sha256:1",
		},
	}
	got, ok := meta.GetResourceUI(res)
	if !ok {
		t.Fatal("expected ok")
	}
	if got.CSP != "" {
		t.Fatalf("expected CSP to be empty when absent, got %q", got.CSP)
	}
}

func TestResourceUI_CSPPolicyRoundTrip(t *testing.T) {
	res := &schema.Resource{Name: "demo/show_widget", Uri: "ui://agently.wkabc/demo/show_widget"}
	want := &meta.CSPPolicy{
		ScriptSrc:           []string{"https://cdn.example"},
		StyleSrc:            []string{"https://fonts.example"},
		ConnectSrc:          []string{"https://api.example"},
		ImgSrc:              []string{"https://img.example"},
		FontSrc:             []string{"https://fonts.example"},
		ScriptStrictDynamic: true,
	}
	meta.SetResourceUI(res, meta.ResourceUI{
		ContentHash: "sha256:1",
		CSPPolicy:   want,
	})
	got, ok := meta.GetResourceUI(res)
	if !ok {
		t.Fatal("expected ok")
	}
	if !reflect.DeepEqual(got.CSPPolicy, want) {
		t.Fatalf("csp policy round-trip mismatch: got %#v want %#v", got.CSPPolicy, want)
	}
	ui := res.Meta["ui"].(map[string]interface{})
	if _, present := ui[meta.FieldCSPPolicy]; !present {
		t.Fatalf("expected cspPolicy key to be written, got %#v", ui)
	}
}

func TestReadingForeignSliceShapes(t *testing.T) {
	tool := &schema.Tool{
		Name: "demo",
		Meta: map[string]interface{}{
			"ui": map[string]interface{}{
				meta.FieldResourceUri:  "ui://x/demo/y",
				meta.FieldAllowedTools: []interface{}{"a", "b"},
			},
		},
	}
	got, ok := meta.GetToolUI(tool)
	if !ok {
		t.Fatal("expected ok")
	}
	if !reflect.DeepEqual(got.AllowedTools, []string{"a", "b"}) {
		t.Fatalf("allowed tools not read from []interface{}: %#v", got.AllowedTools)
	}
}
