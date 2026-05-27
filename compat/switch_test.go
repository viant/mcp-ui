package compat_test

import (
	"context"
	"errors"
	"testing"

	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-ui/capabilities"
	"github.com/viant/mcp-ui/compat"
	"github.com/viant/mcp-ui/meta"
)

func uiCapableClient() *schema.ClientCapabilities {
	caps := &schema.ClientCapabilities{}
	capabilities.SetClientCapability(caps, capabilities.Capability{
		ProtocolVersion: "1.0.0",
		MimeTypes:       []string{capabilities.ResourceMimeType},
	})
	return caps
}

func TestClientUICapable(t *testing.T) {
	if compat.ClientUICapable(nil) {
		t.Fatal("nil client caps must not be reported as UI capable")
	}
	if compat.ClientUICapable(&schema.ClientCapabilities{}) {
		t.Fatal("empty client caps must not be reported as UI capable")
	}
	if !compat.ClientUICapable(uiCapableClient()) {
		t.Fatal("client with UI extension capability must be reported as UI capable")
	}
}

func TestServerUICapable(t *testing.T) {
	if compat.ServerUICapable(nil) {
		t.Fatal("nil server caps must not be reported as UI capable")
	}
	srvCaps := &schema.ServerCapabilities{}
	capabilities.SetServerCapability(srvCaps, capabilities.Capability{
		ProtocolVersion: "1.0.0",
		MimeTypes:       []string{capabilities.ResourceMimeType},
	})
	if !compat.ServerUICapable(srvCaps) {
		t.Fatal("server with UI extension capability must be reported as UI capable")
	}
}

func TestSelectToolMode(t *testing.T) {
	testCases := []struct {
		description string
		caps        *schema.ClientCapabilities
		toolUI      meta.ToolUI
		expected    compat.Mode
	}{
		{
			description: "no resource uri returns ModeNone even when UI capable",
			caps:        uiCapableClient(),
			toolUI:      meta.ToolUI{},
			expected:    compat.ModeNone,
		},
		{
			description: "no resource uri with embedded opt-in still returns ModeNone",
			caps:        nil,
			toolUI:      meta.ToolUI{Fallback: meta.FallbackEmbedded},
			expected:    compat.ModeNone,
		},
		{
			description: "UI capable client with resource uri uses capability path",
			caps:        uiCapableClient(),
			toolUI:      meta.ToolUI{ResourceUri: "ui://x/widget"},
			expected:    compat.ModeCapability,
		},
		{
			description: "UI capable client ignores embedded fallback opt-in",
			caps:        uiCapableClient(),
			toolUI:      meta.ToolUI{ResourceUri: "ui://x/widget", Fallback: meta.FallbackEmbedded},
			expected:    compat.ModeCapability,
		},
		{
			description: "non UI capable client with embedded opt-in returns ModeEmbeddedFallback",
			caps:        nil,
			toolUI:      meta.ToolUI{ResourceUri: "ui://x/widget", Fallback: meta.FallbackEmbedded},
			expected:    compat.ModeEmbeddedFallback,
		},
		{
			description: "non UI capable client without opt-in returns ModeNone",
			caps:        nil,
			toolUI:      meta.ToolUI{ResourceUri: "ui://x/widget"},
			expected:    compat.ModeNone,
		},
		{
			description: "empty client caps without opt-in returns ModeNone",
			caps:        &schema.ClientCapabilities{},
			toolUI:      meta.ToolUI{ResourceUri: "ui://x/widget"},
			expected:    compat.ModeNone,
		},
		{
			description: "unknown fallback value is not honored",
			caps:        nil,
			toolUI:      meta.ToolUI{ResourceUri: "ui://x/widget", Fallback: "iframe"},
			expected:    compat.ModeNone,
		},
	}

	for _, tc := range testCases {
		got := compat.SelectToolMode(tc.caps, tc.toolUI)
		if got != tc.expected {
			t.Errorf("%s: got %s want %s", tc.description, got, tc.expected)
		}
	}
}

func TestSelectResourceMode(t *testing.T) {
	testCases := []struct {
		description string
		caps        *schema.ClientCapabilities
		ui          meta.ResourceUI
		expected    compat.Mode
	}{
		{
			description: "UI capable uses capability path",
			caps:        uiCapableClient(),
			ui:          meta.ResourceUI{},
			expected:    compat.ModeCapability,
		},
		{
			description: "non UI capable with embedded opt-in returns ModeEmbeddedFallback",
			caps:        nil,
			ui:          meta.ResourceUI{Fallback: meta.FallbackEmbedded},
			expected:    compat.ModeEmbeddedFallback,
		},
		{
			description: "non UI capable without opt-in returns ModeNone",
			caps:        nil,
			ui:          meta.ResourceUI{},
			expected:    compat.ModeNone,
		},
		{
			description: "unknown fallback value is not honored",
			caps:        nil,
			ui:          meta.ResourceUI{Fallback: "something-else"},
			expected:    compat.ModeNone,
		},
	}
	for _, tc := range testCases {
		got := compat.SelectResourceMode(tc.caps, tc.ui)
		if got != tc.expected {
			t.Errorf("%s: got %s want %s", tc.description, got, tc.expected)
		}
	}
}

type stubReader struct {
	result *schema.ReadResourceResult
	err    error
	gotUri string
}

func (s *stubReader) ReadResource(_ context.Context, uri string) (*schema.ReadResourceResult, error) {
	s.gotUri = uri
	return s.result, s.err
}

func ptr(s string) *string { return &s }

func TestBuildEmbeddedFallback_PreservesCanonicalMetaAndMime(t *testing.T) {
	uri := "ui://x/widget"
	html := "<!DOCTYPE html><html><body>hi</body></html>"
	mime := capabilities.ResourceMimeType
	contents := schema.ReadResourceResultContentsElem{
		Uri:      uri,
		MimeType: &mime,
		Text:     html,
	}
	meta.SetReadResultContentsUI(&contents, meta.ResourceUI{
		ContentHash:     "sha256:abc",
		ProtocolVersion: "1.0.0",
	})
	reader := &stubReader{
		result: &schema.ReadResourceResult{
			Contents: []schema.ReadResourceResultContentsElem{contents},
		},
	}

	embedded, err := compat.BuildEmbeddedFallback(context.Background(), reader, uri)
	if err != nil {
		t.Fatalf("BuildEmbeddedFallback returned error: %v", err)
	}
	if reader.gotUri != uri {
		t.Fatalf("reader received uri %q want %q", reader.gotUri, uri)
	}
	if embedded.Type != "resource" {
		t.Fatalf("Type=%q want %q", embedded.Type, "resource")
	}
	if embedded.Resource.Uri != uri {
		t.Fatalf("Resource.Uri=%q want %q", embedded.Resource.Uri, uri)
	}
	if embedded.Resource.MimeType == nil || *embedded.Resource.MimeType != mime {
		t.Fatalf("Resource.MimeType=%v want %q", embedded.Resource.MimeType, mime)
	}
	if embedded.Resource.Text != html {
		t.Fatalf("Resource.Text drift: got %q", embedded.Resource.Text)
	}
	ui, ok := meta.GetEmbeddedResourceUI(embedded)
	if !ok {
		t.Fatal("embedded resource missing canonical _meta.ui")
	}
	if ui.ContentHash != "sha256:abc" || ui.ProtocolVersion != "1.0.0" {
		t.Fatalf("canonical _meta.ui not preserved: %#v", ui)
	}
}

func TestBuildEmbeddedFallback_DefaultsMimeWhenAbsent(t *testing.T) {
	uri := "ui://x/widget"
	html := "<!DOCTYPE html><html><body>hi</body></html>"
	reader := &stubReader{
		result: &schema.ReadResourceResult{
			Contents: []schema.ReadResourceResultContentsElem{
				{Uri: uri, Text: html},
			},
		},
	}
	embedded, err := compat.BuildEmbeddedFallback(context.Background(), reader, uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if embedded.Resource.MimeType == nil || *embedded.Resource.MimeType != capabilities.ResourceMimeType {
		t.Fatalf("default MIME not applied: %#v", embedded.Resource.MimeType)
	}
}

func TestBuildEmbeddedFallback_NilReaderRejected(t *testing.T) {
	if _, err := compat.BuildEmbeddedFallback(context.Background(), nil, "ui://x/widget"); err == nil {
		t.Fatal("expected error for nil reader")
	}
}

func TestBuildEmbeddedFallback_EmptyUriRejected(t *testing.T) {
	reader := &stubReader{result: &schema.ReadResourceResult{}}
	if _, err := compat.BuildEmbeddedFallback(context.Background(), reader, ""); err == nil {
		t.Fatal("expected error for empty uri")
	}
}

func TestBuildEmbeddedFallback_NoContentsRejected(t *testing.T) {
	reader := &stubReader{result: &schema.ReadResourceResult{}}
	if _, err := compat.BuildEmbeddedFallback(context.Background(), reader, "ui://x/widget"); err == nil {
		t.Fatal("expected error when read returns no contents")
	}
}

func TestBuildEmbeddedFallback_EmptyTextRejected(t *testing.T) {
	uri := "ui://x/widget"
	mime := capabilities.ResourceMimeType
	reader := &stubReader{
		result: &schema.ReadResourceResult{
			Contents: []schema.ReadResourceResultContentsElem{{Uri: uri, MimeType: &mime}},
		},
	}
	if _, err := compat.BuildEmbeddedFallback(context.Background(), reader, uri); err == nil {
		t.Fatal("expected error when text contents are empty")
	}
}

func TestBuildEmbeddedFallback_PropagatesReaderError(t *testing.T) {
	want := errors.New("boom")
	reader := &stubReader{err: want}
	_, err := compat.BuildEmbeddedFallback(context.Background(), reader, "ui://x/widget")
	if !errors.Is(err, want) {
		t.Fatalf("error not propagated: got %v want %v", err, want)
	}
}

var _ = ptr
