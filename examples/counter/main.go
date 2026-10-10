package main

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"github.com/viant/jsonrpc"
	"github.com/viant/mcp-protocol/schema"
	proto "github.com/viant/mcp-protocol/server"
	"github.com/viant/mcp-ui/capabilities"
	"github.com/viant/mcp-ui/meta"
	"github.com/viant/mcp-ui/resource"
	"github.com/viant/mcp/server"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

//go:embed counter.html
var html string

const uri = "ui://counter.demo/counter.html"

var state struct {
	sync.Mutex
	Count int
}

func result(increment bool) *schema.CallToolResult {
	state.Lock()
	defer state.Unlock()
	if increment {
		state.Count++
	}
	value := map[string]interface{}{"count": state.Count}
	data, _ := json.Marshal(value)
	log.Printf("counter increment=%t count=%d", increment, state.Count)
	return &schema.CallToolResult{Content: []schema.CallToolResultContentElem{schema.TextContent{Type: "text", Text: string(data)}}, StructuredContent: value}
}
func main() {
	handler := proto.WithDefaultHandler(context.Background(), func(h *proto.DefaultHandler) error {
		h.ServerCapabilities = &schema.ServerCapabilities{}
		capabilities.SetServerCapability(h.ServerCapabilities, capabilities.Capability{MimeTypes: []string{capabilities.ResourceMimeType}})
		res, err := resource.NewHTMLResource(uri, "counter", nil, nil, nil, meta.ResourceUI{})
		if err != nil {
			return err
		}
		h.RegisterResource(*res, func(ctx context.Context, r *schema.ReadResourceRequest) (*schema.ReadResourceResult, *jsonrpc.Error) {
			contents, err := resource.NewReadResultHTMLContents(uri, html, meta.ResourceUI{})
			if err != nil {
				return nil, jsonrpc.NewInternalError(err.Error(), nil)
			}
			log.Print("resources/read counter")
			return &schema.ReadResourceResult{Contents: []schema.ReadResourceResultContentsElem{*contents}}, nil
		})
		for _, name := range []string{"counter_view", "counter_increment"} {
			tool := schema.Tool{Name: name, InputSchema: schema.ToolInputSchema{Type: "object", Properties: schema.ToolInputSchemaProperties{}}}
			meta.SetToolUI(&tool, meta.ToolUI{ResourceUri: uri})
			increment := name == "counter_increment"
			h.RegisterTool(&proto.ToolEntry{Metadata: tool, Handler: func(ctx context.Context, r *schema.CallToolRequest) (*schema.CallToolResult, *jsonrpc.Error) {
				if h.ClientInitialize != nil {
					data, _ := json.Marshal(h.ClientInitialize)
					log.Printf("client-initialize %s", data)
				}
				return result(increment), nil
			}})
		}
		return nil
	})
	srv, err := server.New(server.WithNewHandler(handler), server.WithImplementation(schema.Implementation{Name: "counter-demo", Version: "1.0.0"}))
	if err != nil {
		log.Fatal(err)
	}
	srv.UseStreamableHTTP(true)
	httpServer := srv.HTTP(context.Background(), "127.0.0.1:29411")
	original := httpServer.Handler
	httpServer.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		if r.Method == "POST" {
			body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(body))
			var request struct {
				Method string                 `json:"method"`
				Params map[string]interface{} `json:"params"`
			}
			_ = json.Unmarshal(body, &request)
			audit := map[string]interface{}{"method": request.Method}
			if metadata, ok := request.Params["_meta"].(map[string]interface{}); ok {
				for _, key := range []string{"io.modelcontextprotocol/clientCapabilities", "io.modelcontextprotocol/clientInfo", "io.modelcontextprotocol/protocolVersion"} {
					if value, ok := metadata[key]; ok {
						audit[key] = value
					}
				}
			}
			if request.Method == "initialize" || request.Method == "discovery" || request.Method == "server/discover" {
				for _, key := range []string{"capabilities", "clientInfo", "protocolVersion"} {
					if value, ok := request.Params[key]; ok {
						audit[key] = value
					}
				}
			}
			data, _ := json.Marshal(audit)
			log.Printf("rpc %s", data)
		}
		original.ServeHTTP(w, r)
		log.Printf("http method=%s elapsed_us=%d", r.Method, time.Since(start).Microseconds())
	})
	log.Fatal(httpServer.ListenAndServe())
}
