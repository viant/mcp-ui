# MCP counter demo

Example counter server using published `viant/mcp` APIs and the enclosing `viant/mcp-ui` source. The HTML guest uses the official MCP Apps SDK. Only standard tool/resource discovery, `tools/call`, `resources/read`, and `_meta.ui.resourceUri` are emitted by demo code. No legacy adapter, proprietary components, or policy metadata is required.

```sh
npm ci
npm run build
GOWORK=off go build -o runtime/counter .
./runtime/counter
```

Server endpoint: `http://127.0.0.1:29411/mcp`. Tools: `counter_view` and `counter_increment`. Resource: `ui://counter.demo/counter.html`, MIME `text/html;profile=mcp-app`.

The count is shared, mutex-protected server state and survives client/view reconnects. It resets when the demo server process restarts. Run `python3 probe.py` against the server to verify tool discovery, resource MIME, three increments, and reconnect; this deliberately increments the current server count.

Add `{"mcpServers":{"counterdemo":{"url":"http://127.0.0.1:29411/mcp"}}}` to a project-scoped `.cursor/mcp.json`, then enable `counterdemo` in Customize → MCPs. Invoke `counter_view`, expand the tool response, and approve the view/increment demo tool calls when prompted.

Agently needs a dedicated sandbox origin and an agent whose configured tool surface includes both counter tools. The verified owned runtime preserved copied default assets and added a scoped `Counter Demo` agent. Agently main includes the iframe repair and scoped Collapse/Reopen/Reload controls. The aa5b9850 baseline retained an empty srcdoc attribute that prevented sandbox loading.

The `agently/` example assets grant only the two counter tools. Copy them into an owned workspace and choose its configured model reference; preserve existing workspace entries. Runtime files and generated artifacts are excluded.
