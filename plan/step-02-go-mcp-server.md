# Step 2: Go MCP Server — Core Implementation

## Files: `homelab-mcp-server/go.mod`, `homelab-mcp-server/main.go`

### `go.mod`

- Module name: `homelab-mcp-server`
- Requires `github.com/modelcontextprotocol/go-sdk` (via replace directive pointing to `../go-sdk`)
- JWT library: `github.com/golang-jwt/jwt/v5`
- JWKS fetching: `github.com/MicahParks/keyfunc/v3` (or manual JWKS fetch)
- Run `go mod tidy` after the fork is cloned

### `main.go` — Components

#### 1. Six MCP Tools

- **`echo`** — accepts `{ "message": string }`, returns the message as-is
- **`get_server_time`** — no params, returns `time.Now().Format(time.RFC3339)`
- **`whoami`** — no params, extracts JWT claims (`preferred_username`, `name`, `sub`) from request context, returns them as JSON
- **`save_note`** — accepts `{ "title": string, "content": string }`, stores in an in-memory `map[userSub][]Note` keyed by the JWT `sub` claim. Each note has `title`, `content`, and `created_at` (timestamp). Returns confirmation message.
- **`list_notes`** — no params, returns all notes for the authenticated user as a JSON array of `[{title, content, created_at}, ...]`. Returns empty array if none.
- **`delete_note`** — accepts `{ "title": string }`, deletes the matching note for the authenticated user. Returns confirmation or "not found".

The notes store is a `sync.RWMutex`-protected `map[string][]Note` where the key is the user's `sub` claim from the JWT. This proves auth-based data isolation — User A cannot see User B's notes.

#### 2. Protected Resource Metadata Endpoint

```
GET /.well-known/oauth-protected-resource
```

Returns:
```json
{
  "resource": "http://localhost:3001",
  "authorization_servers": [
    "https://login.microsoftonline.com/{ENTRA_TENANT_ID}/v2.0"
  ],
  "scopes_supported": ["mcp.tools"],
  "bearer_methods_supported": ["header"]
}
```

#### 3. JWT Validation Middleware

- Fetches JWKS from `https://login.microsoftonline.com/{tenant}/discovery/v2.0/keys`
- Caches JWKS with periodic refresh (e.g., every 1 hour)
- Validates:
  - Signature (RSA, using JWKS)
  - `iss` — must match `https://login.microsoftonline.com/{tenant}/v2.0`
  - `aud` — must match `MCP_SERVER_CLIENT_ID` env var
  - `exp` — must not be expired
  - `scp` — must contain `mcp.tools`
- On success: stores extracted user claims in request context
- On failure: returns 401 with `WWW-Authenticate: Bearer` header

#### 4. Streamable HTTP Transport

- Uses go-sdk's MCP server with streamable HTTP handler
- Listens on `:3001`
- JWT middleware wraps MCP endpoints; `/.well-known/*` is unauthenticated

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `ENTRA_TENANT_ID` | Azure AD tenant for JWKS + issuer validation |
| `MCP_SERVER_CLIENT_ID` | App registration client ID (audience claim) |
