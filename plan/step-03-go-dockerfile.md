# Step 3: Go MCP Server — Dockerfile

## File: `homelab-mcp-server/Dockerfile`

Multi-stage build that compiles the Go binary and produces a minimal runtime image.

### Build Stage

- Base: `golang:1.23-alpine`
- Copies both `go-sdk/` and `homelab-mcp-server/` into the build context
- The `replace` directive in `go.mod` references `../go-sdk`, so both must be present
- Runs `go build -o /mcp-server .`

### Runtime Stage

- Base: `alpine:latest`
- Installs `ca-certificates` (needed for HTTPS calls to Entra JWKS endpoint)
- Copies the compiled binary from the build stage
- Exposes port 3001
- Entrypoint: `/mcp-server`

### Build Context Note

The `docker-compose.yml` (Step 11) sets `context: .` (project root) so both `go-sdk/` and `homelab-mcp-server/` are available during build. The Dockerfile path is `homelab-mcp-server/Dockerfile`.
