# Step 4: Python Backend — Dependencies & Config

## Files: `chatkit-backend/requirements.txt`, `chatkit-backend/config.py`

### `requirements.txt`

```
openai-chatkit
openai-agents
fastapi
uvicorn[standard]
httpx
msal
python-jose[cryptography]
mcp
```

| Package | Purpose |
|---------|---------|
| `openai-chatkit` | ChatKit Python SDK — server, thread management, SSE streaming |
| `openai-agents` | OpenAI Agents SDK — agent orchestration, tool calling |
| `fastapi` | Web framework |
| `uvicorn[standard]` | ASGI server |
| `httpx` | Async HTTP client (used by `mcp` package internally) |
| `msal` | Microsoft Authentication Library — OIDC + OBO token exchange |
| `python-jose[cryptography]` | JWT decoding/validation |
| `mcp` | Official MCP Python SDK — `ClientSession` + `StreamableHTTPTransport` |

### `config.py`

Reads all configuration from environment variables:

```python
import os

ENTRA_TENANT_ID = os.environ["ENTRA_TENANT_ID"]
ENTRA_CLIENT_ID = os.environ["ENTRA_CLIENT_ID"]           # ChatKit backend app
ENTRA_CLIENT_SECRET = os.environ["ENTRA_CLIENT_SECRET"]
MCP_SERVER_SCOPE = os.environ["MCP_SERVER_SCOPE"]          # api://<id>/mcp.tools
MCP_SERVER_URL = os.environ.get("MCP_SERVER_URL", "http://mcp-server:3001")
OPENAI_API_KEY = os.environ["OPENAI_API_KEY"]
REDIRECT_URI = os.environ.get("REDIRECT_URI", "https://localhost/auth/callback")
AUTHORITY = f"https://login.microsoftonline.com/{ENTRA_TENANT_ID}"
```

No defaults for secrets — app fails fast if they're missing.
