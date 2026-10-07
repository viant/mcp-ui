// Package capabilities provides UI extension capability helpers for the MCP
// Apps stable extension specified by SEP-1865 (MCP Apps).
//
// All identifiers used here are sourced byte-for-byte from the upstream MCP
// Apps spec commit pinned in the module README.
package capabilities

import (
	"github.com/viant/mcp-protocol/schema"
)

// ExtensionName is the capability key used to advertise UI extension
// support in the extensions capability slot of both client and server
// capabilities. The value is pinned to the upstream MCP Apps spec.
const ExtensionName = "io.modelcontextprotocol/ui"

// ResourceMimeType is the MIME type that identifies an MCP UI / MCP Apps
// HTML resource. The value is pinned byte-for-byte to the upstream MCP Apps
// spec commit recorded in the module README.
const ResourceMimeType = "text/html;profile=mcp-app"

// Capability is the negotiated payload that producers place under
// Extensions[ExtensionName] in client and server capabilities.
//
// The stable payload declares MIME types:
//
//	{"mimeTypes":["text/html;profile=mcp-app"]}
type Capability struct {
	// ProtocolVersion is a retained legacy field. Leave it empty for stable
	// outer negotiation; ui/initialize separately negotiates the view version.
	ProtocolVersion string `json:"protocolVersion,omitempty"`

	// MimeTypes is the set of UI resource MIME types the producer supports.
	MimeTypes []string `json:"mimeTypes"`
}

// payloadKey* constants name the fields inside the Extensions payload.
const (
	payloadKeyProtocolVersion = "protocolVersion"
	payloadKeyMimeTypes       = "mimeTypes"
)

// SetClientCapability declares UI extension support on the given client
// capabilities by writing cap under Extensions[ExtensionName].
//
// The function preserves any other Extensions entries already present and
// allocates the Extensions map if it is nil. It panics if caps is nil
// because there is no sensible behavior for a missing capabilities holder.
func SetClientCapability(caps *schema.ClientCapabilities, capability Capability) {
	if caps == nil {
		panic("capabilities: SetClientCapability called with nil caps")
	}
	if caps.Extensions == nil {
		caps.Extensions = map[string]map[string]interface{}{}
	}
	caps.Extensions[ExtensionName] = capabilityToMap(capability)
}

// SetServerCapability declares UI extension support on the given server
// capabilities by writing cap under Extensions[ExtensionName].
//
// Behavior mirrors SetClientCapability.
func SetServerCapability(caps *schema.ServerCapabilities, capability Capability) {
	if caps == nil {
		panic("capabilities: SetServerCapability called with nil caps")
	}
	if caps.Extensions == nil {
		caps.Extensions = map[string]map[string]interface{}{}
	}
	caps.Extensions[ExtensionName] = capabilityToMap(capability)
}

// GetClientCapability reads the UI extension capability from the given
// client capabilities. It returns the parsed Capability and true if the
// entry exists; otherwise it returns the zero Capability and false.
//
// caps may be nil; that case returns the zero value and false.
func GetClientCapability(caps *schema.ClientCapabilities) (Capability, bool) {
	if caps == nil {
		return Capability{}, false
	}
	raw, ok := caps.Extensions[ExtensionName]
	if !ok {
		return Capability{}, false
	}
	return mapToCapability(raw), true
}

// GetServerCapability reads the UI extension capability from the given
// server capabilities. Semantics mirror GetClientCapability.
func GetServerCapability(caps *schema.ServerCapabilities) (Capability, bool) {
	if caps == nil {
		return Capability{}, false
	}
	raw, ok := caps.Extensions[ExtensionName]
	if !ok {
		return Capability{}, false
	}
	return mapToCapability(raw), true
}

func capabilityToMap(capability Capability) map[string]interface{} {
	mimeTypes := make([]interface{}, len(capability.MimeTypes))
	for i, mt := range capability.MimeTypes {
		mimeTypes[i] = mt
	}
	result := map[string]interface{}{payloadKeyMimeTypes: mimeTypes}
	// ProtocolVersion was a viant legacy field. Stable extension negotiation
	// declares MIME types; ui/initialize separately negotiates the date version.
	if capability.ProtocolVersion != "" {
		result[payloadKeyProtocolVersion] = capability.ProtocolVersion
	}
	return result
}

func mapToCapability(raw map[string]interface{}) Capability {
	out := Capability{}
	if v, ok := raw[payloadKeyProtocolVersion].(string); ok {
		out.ProtocolVersion = v
	}
	if raw[payloadKeyMimeTypes] != nil {
		switch v := raw[payloadKeyMimeTypes].(type) {
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					out.MimeTypes = append(out.MimeTypes, s)
				}
			}
		case []string:
			out.MimeTypes = append(out.MimeTypes, v...)
		}
	}
	return out
}
