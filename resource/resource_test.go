package resource_test

import (
	"strings"
	"testing"

	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-ui/capabilities"
	metaui "github.com/viant/mcp-ui/meta"
	"github.com/viant/mcp-ui/resource"
)

func TestValidateUIURI(t *testing.T) {
	valid := "ui://agently.wk_ab12cd34ef56ab78/demo/show_widget"
	parsed, err := resource.ValidateUIURI(valid)
	if err != nil {
		t.Fatalf("ValidateUIURI failed: %v", err)
	}
	if parsed.ServerScope != "agently.wk_ab12cd34ef56ab78" || parsed.Kind != resource.KindDemo || parsed.ResourceID != "show_widget" {
		t.Fatalf("parsed mismatch: %#v", parsed)
	}
	if _, err := resource.ValidateUIURI("http://x/demo/y"); err == nil {
		t.Fatal("expected scheme validation failure")
	}
	if _, err := resource.ValidateUIURI("ui://x/unknown/y"); err == nil {
		t.Fatal("expected kind validation failure")
	}
}

func TestContentHashStable(t *testing.T) {
	a := resource.ContentHash("<html>ok</html>")
	b := resource.ContentHash("<html>ok</html>")
	c := resource.ContentHash("<html>changed</html>")
	if a != b {
		t.Fatalf("expected stable hash, got %q vs %q", a, b)
	}
	if a == c {
		t.Fatalf("expected different hash for different content")
	}
}

func TestNewHTMLResource(t *testing.T) {
	desc := "demo"
	title := "Demo"
	ui := metaui.ResourceUI{
		AllowedTools:    []string{"message:add"},
		ContentHash:     "sha256:abc",
		ProtocolVersion: "1.0.0",
	}
	res, err := resource.NewHTMLResource(
		"ui://agently.wk_ab12cd34ef56ab78/demo/show_widget",
		"show_widget",
		&desc,
		&title,
		&schema.Annotations{},
		ui,
	)
	if err != nil {
		t.Fatalf("NewHTMLResource failed: %v", err)
	}
	if res.Name != "show_widget" || res.Uri == "" || res.MimeType == nil || *res.MimeType != capabilities.ResourceMimeType {
		t.Fatalf("resource mismatch: %#v", res)
	}
	gotUI, ok := metaui.GetResourceUI(res)
	if !ok || gotUI.ContentHash != "sha256:abc" {
		t.Fatalf("resource ui metadata missing: %#v", gotUI)
	}
}

func TestNewHTMLContents(t *testing.T) {
	ui := metaui.ResourceUI{
		AllowedTools:    []string{"message:add"},
		ContentHash:     "sha256:abc",
		ProtocolVersion: "1.0.0",
	}
	contents, err := resource.NewHTMLContents(
		"ui://agently.wk_ab12cd34ef56ab78/demo/show_widget",
		"<!DOCTYPE html><html><body>ok</body></html>",
		ui,
	)
	if err != nil {
		t.Fatalf("NewHTMLContents failed: %v", err)
	}
	if contents.MimeType == nil || *contents.MimeType != capabilities.ResourceMimeType {
		t.Fatalf("mime mismatch: %#v", contents)
	}
	gotUI, ok := metaui.GetTextResourceContentsUI(contents)
	if !ok || gotUI.ProtocolVersion != "1.0.0" {
		t.Fatalf("ui metadata mismatch: %#v", gotUI)
	}
}

func TestNewReadResultHTMLContents(t *testing.T) {
	ui := metaui.ResourceUI{
		ContentHash:     "sha256:abc",
		ProtocolVersion: "1.0.0",
	}
	contents, err := resource.NewReadResultHTMLContents(
		"ui://agently.wk_ab12cd34ef56ab78/demo/show_widget",
		"<html>ok</html>",
		ui,
	)
	if err != nil {
		t.Fatalf("NewReadResultHTMLContents failed: %v", err)
	}
	if contents.Text != "<html>ok</html>" {
		t.Fatalf("text mismatch: %#v", contents)
	}
	gotUI, ok := metaui.GetReadResultContentsUI(contents)
	if !ok || gotUI.ContentHash != "sha256:abc" {
		t.Fatalf("ui metadata mismatch: %#v", gotUI)
	}
}

func TestNewEmbeddedHTMLResource(t *testing.T) {
	ui := metaui.ResourceUI{
		ContentHash:     "sha256:abc",
		ProtocolVersion: "1.0.0",
		Fallback:        metaui.FallbackEmbedded,
	}
	embedded, err := resource.NewEmbeddedHTMLResource(
		"ui://agently.wk_ab12cd34ef56ab78/demo/show_widget",
		"<html>ok</html>",
		nil,
		ui,
	)
	if err != nil {
		t.Fatalf("NewEmbeddedHTMLResource failed: %v", err)
	}
	if embedded.Type != "resource" || embedded.Resource.MimeType == nil || *embedded.Resource.MimeType != capabilities.ResourceMimeType {
		t.Fatalf("embedded resource mismatch: %#v", embedded)
	}
	gotUI, ok := metaui.GetEmbeddedResourceUI(embedded)
	if !ok || gotUI.Fallback != metaui.FallbackEmbedded {
		t.Fatalf("embedded ui metadata mismatch: %#v", gotUI)
	}
}

func TestNewResourceRejectsEmptyInputs(t *testing.T) {
	if _, err := resource.NewHTMLResource("", "x", nil, nil, nil, metaui.ResourceUI{}); err == nil {
		t.Fatal("expected uri error")
	}
	if _, err := resource.NewHTMLResource("ui://agently.wk_ab12cd34ef56ab78/demo/show_widget", "", nil, nil, nil, metaui.ResourceUI{}); err == nil {
		t.Fatal("expected name error")
	}
	if _, err := resource.NewHTMLContents("ui://agently.wk_ab12cd34ef56ab78/demo/show_widget", "", metaui.ResourceUI{}); err == nil {
		t.Fatal("expected empty html error")
	}
	if _, err := resource.NewReadResultHTMLContents("ui://agently.wk_ab12cd34ef56ab78/demo/show_widget", "", metaui.ResourceUI{}); err == nil {
		t.Fatal("expected empty html error")
	}
	if _, err := resource.NewEmbeddedHTMLResource("ui://agently.wk_ab12cd34ef56ab78/demo/show_widget", "", nil, metaui.ResourceUI{}); err == nil {
		t.Fatal("expected empty html error")
	}
}

func TestNewResourceRejectsOversizedHTML(t *testing.T) {
	oversized := strings.Repeat("a", resource.MaxHTMLBytes+1)
	uri := "ui://agently.wk_ab12cd34ef56ab78/demo/show_widget"
	if _, err := resource.NewHTMLContents(uri, oversized, metaui.ResourceUI{}); err == nil {
		t.Fatal("expected oversized html error for NewHTMLContents")
	}
	if _, err := resource.NewReadResultHTMLContents(uri, oversized, metaui.ResourceUI{}); err == nil {
		t.Fatal("expected oversized html error for NewReadResultHTMLContents")
	}
	if _, err := resource.NewEmbeddedHTMLResource(uri, oversized, nil, metaui.ResourceUI{}); err == nil {
		t.Fatal("expected oversized html error for NewEmbeddedHTMLResource")
	}
}
