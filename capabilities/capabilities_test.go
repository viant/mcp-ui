package capabilities_test

import (
	"reflect"
	"testing"

	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-ui/capabilities"
)

func TestExtensionConstants(t *testing.T) {
	if capabilities.ExtensionName != "io.modelcontextprotocol/ui" {
		t.Fatalf("ExtensionName drifted from upstream pin: %q", capabilities.ExtensionName)
	}
	if capabilities.ResourceMimeType != "text/html;profile=mcp-app" {
		t.Fatalf("ResourceMimeType drifted from upstream pin: %q", capabilities.ResourceMimeType)
	}
}

func TestSetClientCapability_AllocatesAndPreservesOtherEntries(t *testing.T) {
	caps := &schema.ClientCapabilities{
		Extensions: map[string]map[string]interface{}{
			"vendor.other/feature": {"hello": "world"},
		},
	}

	want := capabilities.Capability{
		ProtocolVersion: "1.0.0",
		MimeTypes:       []string{capabilities.ResourceMimeType},
	}
	capabilities.SetClientCapability(caps, want)

	if other, ok := caps.Extensions["vendor.other/feature"]; !ok || other["hello"] != "world" {
		t.Fatalf("unrelated experimental entry was mutated: %#v", caps.Extensions)
	}
	got, ok := capabilities.GetClientCapability(caps)
	if !ok {
		t.Fatalf("expected capability to be present")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch: got %#v want %#v", got, want)
	}
}

func TestSetClientCapability_NilExperimentalMap(t *testing.T) {
	caps := &schema.ClientCapabilities{}
	capabilities.SetClientCapability(caps, capabilities.Capability{
		ProtocolVersion: "1.0.0",
		MimeTypes:       []string{capabilities.ResourceMimeType},
	})
	if caps.Extensions == nil {
		t.Fatal("Experimental map should have been allocated")
	}
	if _, ok := caps.Extensions[capabilities.ExtensionName]; !ok {
		t.Fatal("expected capability entry under ExtensionName")
	}
}

func TestSetServerCapability_RoundTrip(t *testing.T) {
	caps := &schema.ServerCapabilities{}
	want := capabilities.Capability{
		ProtocolVersion: "1.2.3",
		MimeTypes:       []string{capabilities.ResourceMimeType, "text/markdown"},
	}
	capabilities.SetServerCapability(caps, want)

	got, ok := capabilities.GetServerCapability(caps)
	if !ok {
		t.Fatalf("expected capability to be present")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch: got %#v want %#v", got, want)
	}
}

func TestGetCapability_AbsentReturnsFalse(t *testing.T) {
	if _, ok := capabilities.GetClientCapability(nil); ok {
		t.Fatal("nil caps must return ok=false")
	}
	if _, ok := capabilities.GetServerCapability(nil); ok {
		t.Fatal("nil caps must return ok=false")
	}
	if _, ok := capabilities.GetClientCapability(&schema.ClientCapabilities{}); ok {
		t.Fatal("empty caps must return ok=false")
	}
	if _, ok := capabilities.GetServerCapability(&schema.ServerCapabilities{}); ok {
		t.Fatal("empty caps must return ok=false")
	}
}

func TestSetClientCapability_DoesNotMutateUnrelatedSubkeys(t *testing.T) {
	caps := &schema.ClientCapabilities{
		Extensions: map[string]map[string]interface{}{
			capabilities.ExtensionName: {"legacy": "should-be-replaced"},
			"vendor.other/feature":     {"keep": 42},
		},
	}
	capabilities.SetClientCapability(caps, capabilities.Capability{
		ProtocolVersion: "1.0.0",
		MimeTypes:       []string{capabilities.ResourceMimeType},
	})
	if v, ok := caps.Extensions["vendor.other/feature"]["keep"]; !ok || v != 42 {
		t.Fatalf("unrelated entry mutated: %#v", caps.Extensions["vendor.other/feature"])
	}
	if _, hasLegacy := caps.Extensions[capabilities.ExtensionName]["legacy"]; hasLegacy {
		t.Fatalf("old extension payload was not replaced cleanly: %#v", caps.Extensions[capabilities.ExtensionName])
	}
}
