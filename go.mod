module github.com/viant/mcp-ui

go 1.23.8

require github.com/viant/mcp-protocol v0.11.0

require (
	github.com/go-viper/mapstructure/v2 v2.2.1 // indirect
	github.com/viant/jsonrpc v0.7.5 // indirect
)

replace github.com/viant/mcp-protocol => ../mcp-protocol
