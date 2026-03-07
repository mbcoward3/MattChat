# Homelab MCP Server

Go MCP server with Entra ID JWT authentication. Exposes 6 tools via Streamable HTTP transport:

- **echo** — echo a message back
- **get_server_time** — current server time
- **whoami** — authenticated user's identity from JWT claims
- **save_note** — save a personal note (per-user, keyed by JWT `sub`)
- **list_notes** — list your notes
- **delete_note** — delete a note by title

## Prerequisites

- Go 1.25+
- An Azure Entra ID tenant with app registrations (see `PLAN.md` Phases 0-1)

## Setup

The server depends on a fork of the MCP Go SDK that adds enterprise auth support. You need to clone it as a sibling directory.

```bash
# From the project root (MattChat/)
git clone https://github.com/radar07/go-sdk.git
cd go-sdk
git checkout enterprise-managed-authorization
cd ..
```

Then resolve dependencies:

```bash
cd homelab-mcp-server
go mod tidy
```

The `go.mod` has a `replace` directive that points to `../go-sdk`, so both directories must exist side by side:

```
MattChat/
├── go-sdk/                      # cloned fork
└── homelab-mcp-server/          # this directory
    ├── go.mod                   # replace directive -> ../go-sdk
    └── main.go
```

## Running

```bash
export ENTRA_TENANT_ID=your-tenant-id
export MCP_SERVER_CLIENT_ID=your-mcp-server-client-id

# Optional: override listen address (default :3001)
# export LISTEN_ADDR=:3001

go run main.go
```

## Verifying

```bash
# Should return OAuth protected resource metadata (no auth required)
curl http://localhost:3001/.well-known/oauth-protected-resource

# Should return 401 (no bearer token)
curl -v http://localhost:3001/mcp
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ENTRA_TENANT_ID` | Yes | — | Azure AD tenant ID for JWKS + issuer validation |
| `MCP_SERVER_CLIENT_ID` | Yes | — | App registration client ID (JWT audience) |
| `LISTEN_ADDR` | No | `:3001` | Address to listen on |
