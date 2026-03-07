module homelab-mcp-server

go 1.25.0

// The go-sdk fork must be cloned into ../go-sdk from:
//   git clone https://github.com/radar07/go-sdk.git
//   cd go-sdk && git checkout enterprise-managed-authorization
replace github.com/modelcontextprotocol/go-sdk => ../go-sdk

require (
	github.com/MicahParks/keyfunc/v3 v3.8.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/modelcontextprotocol/go-sdk v0.0.0-00010101000000-000000000000
)

require (
	github.com/MicahParks/jwkset v0.11.0 // indirect
	github.com/google/jsonschema-go v0.4.2 // indirect
	github.com/segmentio/asm v1.1.3 // indirect
	github.com/segmentio/encoding v0.5.3 // indirect
	github.com/yosida95/uritemplate/v3 v3.0.2 // indirect
	golang.org/x/oauth2 v0.35.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/time v0.9.0 // indirect
)
