# viant/mcp-ui

Typed Go SDK for the MCP UI / MCP Apps extension.

This module is the dedicated extension layer for UI-specific MCP behavior. It
exists to keep the boundaries clean between:

- `github.com/viant/mcp-protocol` — official MCP schema mirror
- `github.com/viant/mcp` — transport/runtime
- `github.com/viant/mcp-ui` — UI extension capability, metadata, resources,
  browser envelope contracts, and compatibility helpers

## Why this repo exists

UI support for MCP needs its own layer.

This repo owns:

- typed `_meta.ui` helpers
- typed UI extension capability helpers
- `ui://...` resource builders
- browser host/guest `postMessage` envelope types
- exact compatibility switching between negotiated UI resources and explicit
  embedded fallback

This repo intentionally does **not** own:

- core MCP schema definitions
- MCP transport implementations
- Agently-specific server behavior
- Forge-specific rendering behavior

## Scope

Today this module focuses on the extension contract itself.

That means:

- typed capability negotiation
- typed `_meta.ui` metadata
- generic, exact `ui://...` resource construction
- browser host/guest envelope types
- exact compatibility switching between negotiated UI resources and explicit
  embedded fallback

It does **not** try to be:

- a browser host implementation
- an MCP server framework
- a product-specific UI toolkit
- a replacement for `mcp-protocol` or `mcp`

## Status

The SDK surface is active and usable.

Implemented in the current worktree:

- capability helpers in `capabilities`
- typed metadata helpers in `meta`
- resource builders in `resource`
- browser envelope types in `appproto`
- compatibility switching in `compat`

The current compatibility rule is exact (the negotiated MIME list must include `text/html;profile=mcp-app`):

- if UI extension capability is negotiated, use `_meta.ui.resourceUri` plus
  `resources/read`
- if UI extension capability is **not** negotiated, embedded fallback is
  allowed **only** when `_meta.ui.fallback = "embedded"` is explicitly present

The `appproto` package is an explicit legacy compatibility API. It does not implement official `ui/initialize` or the App/AppBridge JSON-RPC lifecycle.

Current maturity in this repo:

- the SDK contract is implemented and test-backed
- compatibility switching is implemented and test-backed
- browser-host rollout, production review, and product-specific policy remain
  downstream concerns for consuming projects

In other words: this repo is ready to be used as an extension SDK, while
deployment posture is decided by downstream hosts.

## Specification pin

This SDK targets the stable MCP Apps extension (2026-01-26) specified by SEP-1865.

- Repository: `modelcontextprotocol/ext-apps`
- Commit SHA: `9a37ad71827d076af06978fa7f7f510449687061`
- Pinned source:
  [apps.mdx](https://github.com/modelcontextprotocol/ext-apps/blob/9a37ad71827d076af06978fa7f7f510449687061/specification/2026-01-26/apps.mdx)
- SEP discussion:
  [modelcontextprotocol/discussions/1865](https://github.com/modelcontextprotocol/modelcontextprotocol/discussions/1865)

Official identifiers and fields follow that stable document. The `appproto`
package's `mcpui:*` envelopes and the additional legacy metadata fields are
explicit compatibility contracts, not the official host/view JSON-RPC protocol.
Official web hosts should use the upstream App/AppBridge SDK and an isolated
sandbox proxy; this Go module does not implement that browser lifecycle.

## Key exported concepts

Pinned identifiers:

- extension capability key: `io.modelcontextprotocol/ui`
- UI resource MIME type: `text/html;profile=mcp-app`
- canonical tool metadata field: `_meta.ui.resourceUri`

Current metadata surface includes:

- `resourceUri`
- official `visibility` audiences
- official object `csp` via `ResourceUI.OfficialCSP`
- official `domain`, `prefersBorder` and optional browser `permissions`
- `allowedTools`
- `allowedToolBundles`
- `contentHash`
- `protocolVersion`
- `rendererUrl`
- `sandbox`
- legacy raw `csp`
- structured `cspPolicy`
- `fallback`

## Package map

### `capabilities`

Capability helpers for the UI extension. They read/write
`capabilities.extensions["io.modelcontextprotocol/ui"]` with `mimeTypes`.
The optional `Capability.ProtocolVersion` field is retained for compatibility;
it is omitted when empty and does not negotiate the separate view handshake.
The stable host/view protocol version is negotiated by `ui/initialize`.

Use it to:

- advertise UI support on client/server capabilities
- detect whether the peer negotiated UI support

Main exports:

- `ExtensionName`
- `ResourceMimeType`
- `Capability`
- `SetClientCapability`
- `SetServerCapability`
- `GetClientCapability`
- `GetServerCapability`

### `meta`

Typed `_meta.ui` helpers for tools and resources.

Use it to:

- advertise `ui://...` resources on tools
- attach render/security hints to resources
- carry structured CSP policy and explicit embedded-fallback intent

Main exports:

- `ToolUI`
- `ResourceUI`
- `CSPPolicy`
- `SetToolUI`
- `GetToolUI`
- `SetResourceUI`
- `GetResourceUI`
- `SetReadResultContentsUI`
- `GetReadResultContentsUI`

### `resource`

Builders for `ui://...` resources and resource contents.

Use it to:

- create MCP resource declarations
- create `resources/read` HTML contents
- create embedded HTML fallback resources
- compute canonical content hashes

Main exports:

- `ValidateUIURI` for generic official resource identities
- `ValidateScopedUIURI` only for the legacy `server/kind/id` naming convention
- `ContentHash`
- `NewHTMLResource`
- `NewHTMLContents`
- `NewReadResultHTMLContents`
- `NewEmbeddedHTMLResource`

### `appproto`

Browser host/guest `postMessage` contract.

This is explicitly **not** MCP JSON-RPC.

Use it to type and validate:

- `mcpui:tools-call`
- `mcpui:message`
- `mcpui:open-link`
- `mcpui:host-ready`
- `mcpui:tool-input`
- `mcpui:tool-input-partial`
- `mcpui:tool-result`
- `mcpui:teardown`

### `compat`

Exact compatibility switching helpers.

Use it to:

- decide whether a tool/resource should use negotiated UI-resource flow
- decide whether explicit embedded fallback is allowed
- build embedded fallback from canonical `resources/read` output

## Actionable example

The example below shows a full minimal flow:

1. declare UI capability on the server
2. advertise a `ui://...` resource on a tool
3. publish the matching resource
4. choose negotiated UI-resource flow vs embedded fallback

```go
package main

import (
	"context"
	"fmt"

	"github.com/viant/mcp-protocol/schema"
	"github.com/viant/mcp-ui/capabilities"
	"github.com/viant/mcp-ui/compat"
	"github.com/viant/mcp-ui/meta"
	"github.com/viant/mcp-ui/resource"
)

func main() {
	const (
		uri  = "ui://example.demo/widget/show_status"
		html = "<!DOCTYPE html><html><body><h1>Status OK</h1></body></html>"
	)

	// 1. Server advertises UI capability.
	serverCaps := &schema.ServerCapabilities{}
	capabilities.SetServerCapability(serverCaps, capabilities.Capability{
		MimeTypes:       []string{capabilities.ResourceMimeType},
	})

	// 2. Tool advertises a UI resource and explicit embedded fallback.
	tool := &schema.Tool{Name: "show_status"}
	meta.SetToolUI(tool, meta.ToolUI{
		ResourceUri: uri,
		Fallback:    meta.FallbackEmbedded,
	})

	// 3. Resource carries canonical UI metadata.
	uiMeta := meta.ResourceUI{
		ContentHash:     resource.ContentHash(html),
		ProtocolVersion: "1.0.0",
		CSPPolicy: &meta.CSPPolicy{
			ScriptSrc: []string{"https://cdn.example"},
		},
	}

	res, err := resource.NewHTMLResource(
		uri,
		"show_status",
		ptr("Demo status widget"),
		ptr("Demo Status"),
		nil,
		uiMeta,
	)
	if err != nil {
		panic(err)
	}

	readContents, err := resource.NewReadResultHTMLContents(uri, html, uiMeta)
	if err != nil {
		panic(err)
	}

	// 4. Client decides whether to use negotiated UI-resource flow
	//    or explicit embedded fallback.
	clientCaps := &schema.ClientCapabilities{}
	mode := compat.SelectToolMode(clientCaps, meta.ToolUI{
		ResourceUri: uri,
		Fallback:    meta.FallbackEmbedded,
	})

	fmt.Println("tool:", tool.Name)
	fmt.Println("resource:", res.Uri)
	fmt.Println("mode:", mode.String())
	fmt.Println("read mime:", *readContents.MimeType)
	fmt.Println("read hash:", uiMeta.ContentHash)

	// If the peer had no UI capability and you needed fallback:
	reader := compatReader{contents: *readContents}
	embedded, err := compat.BuildEmbeddedFallback(context.Background(), reader, uri)
	if err != nil {
		panic(err)
	}
	fmt.Println("embedded type:", embedded.Type)
}

type compatReader struct {
	contents schema.ReadResourceResultContentsElem
}

func (r compatReader) ReadResource(context.Context, string) (*schema.ReadResourceResult, error) {
	return &schema.ReadResourceResult{
		Contents: []schema.ReadResourceResultContentsElem{r.contents},
	}, nil
}

func ptr[T any](v T) *T { return &v }
```

The stable specification permits resource IDs such as `ui://weather-dashboard` or
`ui://weather-server/dashboard-template`; product naming rules must not restrict
these identities. `compat.BuildEmbeddedFallback` selects one exact matching URI,
rejects ambiguous or incompatible contents, and bounds UTF-8 HTML/blob payloads.
Legacy fallback may default an omitted MIME; the official resource path still
requires the exact MCP Apps MIME. Hosts own HTML validation, resource review, CSP
enforcement, sandboxing and permission decisions.

## Development

Local replace in a consumer repo:

```go
replace github.com/viant/mcp-ui => ../mcp-ui
```

Published dependencies are pinned in `go.mod`; local replacements are development-only.

## Testing

Run the package tests from this repo root:

```bash
go test ./...
```

For the package-level suites used most often during development:

```bash
go test ./capabilities ./meta ./resource ./compat ./appproto
```

## Boundaries

This repo:

- depends on `github.com/viant/mcp-protocol`
- must not depend on:
  - `github.com/viant/agently-core`
  - `github.com/viant/agently`
  - `github.com/viant/forge`

That separation is intentional: this SDK defines the extension contract, while
hosts and products consume it downstream.

## License

Apache 2.0. See [LICENSE](LICENSE).
