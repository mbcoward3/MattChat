# Step 1: Project Scaffold

Create the directory structure, `.gitignore`, and `.env.example`.

## Directory Structure

```
MattChat/
├── docker-compose.yml
├── .env.example
├── .gitignore
├── Caddyfile
├── PLAN.md                         # already exists
├── homelab-mcp-server/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
└── chatkit-backend/
    ├── Dockerfile
    ├── requirements.txt
    ├── config.py
    ├── entra_auth.py
    ├── mcp_client.py
    ├── agent.py
    ├── chatkit_server.py
    ├── app.py
    └── static/
        └── index.html
```

## Actions

1. Create directories: `homelab-mcp-server/`, `chatkit-backend/`, `chatkit-backend/static/`
2. Create `.gitignore` covering:
   - Python: `__pycache__/`, `*.pyc`, `.venv/`, `*.egg-info/`
   - Go: binary output
   - Secrets: `.env` (but NOT `.env.example`)
   - IDE: `.vscode/`, `.idea/`
   - Docker: volume data
3. Create `.env.example` with placeholder values and descriptive comments for all required env vars:
   - `ENTRA_TENANT_ID`
   - `ENTRA_CLIENT_ID`
   - `ENTRA_CLIENT_SECRET`
   - `MCP_SERVER_SCOPE`
   - `MCP_SERVER_CLIENT_ID`
   - `OPENAI_API_KEY`

## Prerequisites

The `go-sdk` fork (`radar07/go-sdk`, branch `enterprise-managed-authorization`) must be cloned into the project root before the Go server can build. This is handled in Step 12.
