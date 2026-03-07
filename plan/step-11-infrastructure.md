# Step 11: Infrastructure — Caddy, Docker Compose, Env

## Files: `Caddyfile`, `docker-compose.yml`, `.env.example`

### `Caddyfile`

TLS-terminating reverse proxy using Caddy's automatic self-signed cert for `localhost`.

```
https://localhost {
  reverse_proxy chatkit-backend:8000
  tls internal
}
```

All routes (`/`, `/login`, `/auth/*`, `/api/*`) proxy to the Python backend on port 8000. Caddy handles HTTPS so Entra's redirect URI works with `https://localhost`.

### `docker-compose.yml`

Three services:

| Service | Build | Ports | Depends On |
|---------|-------|-------|------------|
| `caddy` | `caddy:2-alpine` image | 443, 80 (host) | chatkit-backend |
| `chatkit-backend` | `./chatkit-backend` | 8000 (internal) | mcp-server |
| `mcp-server` | context `.`, dockerfile `homelab-mcp-server/Dockerfile` | 3001 (internal) | — |

Key details:
- `caddy` mounts `./Caddyfile` and uses named volumes for cert data
- `chatkit-backend` gets all env vars from `.env` file
- `mcp-server` gets `ENTRA_TENANT_ID` and `MCP_SERVER_CLIENT_ID` from `.env`
- `chatkit-backend` reaches the MCP server at `http://mcp-server:3001` (Docker internal DNS)

### `.env.example`

```env
# Microsoft Entra ID (Azure AD)
ENTRA_TENANT_ID=your-tenant-id-here
ENTRA_CLIENT_ID=your-chatkit-backend-client-id
ENTRA_CLIENT_SECRET=your-chatkit-backend-client-secret

# MCP Server app registration
MCP_SERVER_SCOPE=api://your-mcp-server-client-id/mcp.tools
MCP_SERVER_CLIENT_ID=your-mcp-server-client-id

# OpenAI
OPENAI_API_KEY=sk-your-openai-api-key
```

### Python Backend `Dockerfile`

```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
```
