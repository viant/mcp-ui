package appproto_test

import (
	"encoding/json"
	"testing"

	"github.com/viant/mcp-ui/appproto"
)

func TestConstants(t *testing.T) {
	if appproto.Version != "1.0.0" {
		t.Fatalf("Version drifted: %q", appproto.Version)
	}
	want := map[string]string{
		"tools":    appproto.MethodToolsCall,
		"message":  appproto.MethodMessage,
		"open":     appproto.MethodOpenLink,
		"ready":    appproto.MethodHostReady,
		"input":    appproto.MethodToolInput,
		"partial":  appproto.MethodToolInputPartial,
		"result":   appproto.MethodToolResult,
		"teardown": appproto.MethodTeardown,
		"size":     appproto.MethodSizeChanged,
	}
	expected := map[string]string{
		"tools":    "mcpui:tools-call",
		"message":  "mcpui:message",
		"open":     "mcpui:open-link",
		"ready":    "mcpui:host-ready",
		"input":    "mcpui:tool-input",
		"partial":  "mcpui:tool-input-partial",
		"result":   "mcpui:tool-result",
		"teardown": "mcpui:teardown",
		"size":     "mcpui:size-changed",
	}
	for k, got := range want {
		if got != expected[k] {
			t.Fatalf("%s drifted: got %q want %q", k, got, expected[k])
		}
	}
}

func TestEnvelopeRoundTrip_ToolsCall(t *testing.T) {
	env := appproto.NewEnvelope(appproto.MethodToolsCall, appproto.ToolsCallParams{
		WindowID:    "w1",
		ResourceURI: "ui://agently.wk_ab/view/order",
		Name:        "ui/window:get",
		Arguments:   map[string]interface{}{"windowId": "w1"},
	})
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded appproto.Envelope[appproto.ToolsCallParams]
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Version != appproto.Version || decoded.Method != appproto.MethodToolsCall {
		t.Fatalf("decoded envelope mismatch: %#v", decoded)
	}
	if decoded.Params.Name != "ui/window:get" || decoded.Params.WindowID != "w1" {
		t.Fatalf("decoded params mismatch: %#v", decoded.Params)
	}
}

func TestEnvelopeRoundTrip_HostReady(t *testing.T) {
	env := appproto.NewEnvelope(appproto.MethodHostReady, appproto.HostReadyParams{
		WindowID:        "w1",
		ResourceURI:     "ui://agently.wk_ab/demo/show_widget",
		AllowedTools:    []string{"message:add"},
		ProtocolVersion: appproto.Version,
	})
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded appproto.Envelope[appproto.HostReadyParams]
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Params.ProtocolVersion != appproto.Version {
		t.Fatalf("protocol version mismatch: %#v", decoded.Params)
	}
	if len(decoded.Params.AllowedTools) != 1 || decoded.Params.AllowedTools[0] != "message:add" {
		t.Fatalf("allowed tools mismatch: %#v", decoded.Params.AllowedTools)
	}
}

func TestEnvelopeRoundTrip_ToolResult(t *testing.T) {
	env := appproto.NewEnvelope(appproto.MethodToolResult, appproto.ToolResultParams{
		WindowID:          "w1",
		ResourceURI:       "ui://agently.wk_ab/demo/show_widget",
		ToolName:          "demo/show_widget",
		Content:           []map[string]interface{}{{"type": "text", "text": "ok"}},
		StructuredContent: map[string]interface{}{"ok": true},
		Meta:              map[string]interface{}{"source": "test"},
		ProtocolVersion:   appproto.Version,
	})
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded appproto.Envelope[appproto.ToolResultParams]
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Method != appproto.MethodToolResult {
		t.Fatalf("method mismatch: %#v", decoded)
	}
	if decoded.Params.Meta["source"] != "test" {
		t.Fatalf("meta mismatch: %#v", decoded.Params.Meta)
	}
}
