# Step 12: Clone go-sdk Fork & Verify Go Build

## Actions

### 1. Clone the fork

```bash
cd /home/matt/Projects/MattChat
git clone https://github.com/radar07/go-sdk.git
cd go-sdk
git checkout enterprise-managed-authorization
git log --oneline -5   # note HEAD commit hash to pin later
cd ..
```

### 2. Resolve Go dependencies

```bash
cd homelab-mcp-server
go mod tidy
```

This will download all dependencies and populate `go.sum`.

### 3. Verify build

```bash
go build ./...
```

Should compile without errors. If there are API changes in the fork, adjust `main.go` accordingly.

### 4. Test standalone

```bash
# Set placeholder values (server won't validate tokens without real Entra setup)
export ENTRA_TENANT_ID=placeholder
export MCP_SERVER_CLIENT_ID=placeholder

go run main.go
```

In another terminal:
```bash
# Should return metadata JSON
curl http://localhost:3001/.well-known/oauth-protected-resource

# Should return 401 (no bearer token)
curl -v http://localhost:3001/mcp
```

### Note

The `go-sdk/` directory should be added to `.gitignore` since it's a cloned dependency (or alternatively, tracked as a git submodule). The Dockerfile copies it from the build context.
