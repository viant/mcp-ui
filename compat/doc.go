// Package compat provides compatibility helpers for the MCP UI extension
// proposed by SEP-1865 (MCP Apps).
//
// The package owns the exact, doc-defined switching rule between:
//
//   - the negotiated UI extension path — server emits `_meta.ui.resourceUri`
//     and the host calls `resources/read` to fetch the canonical payload; and
//   - the embedded-resource fallback path — server appends a UI EmbeddedResource
//     directly to the tool result, used only when the peer does NOT advertise
//     UI extension capability AND the tool/resource has explicitly opted in
//     via `_meta.ui.fallback = "embedded"`.
//
// There are no heuristics: a tool/resource that does not advertise both
// `_meta.ui.resourceUri` and `_meta.ui.fallback = "embedded"` cannot receive
// an embedded-resource fallback through these helpers, even if the peer is
// UI-incapable.
package compat
