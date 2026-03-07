# Homelab Checklist: Self-Hosted ChatKit + Entra ID + Go MCP Server with Enterprise Auth

## Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                         User's Browser                               │
│  ┌──────────────────┐                                                │
│  │  ChatKit JS       │  (minimal HTML + chatkit-js web component)    │
│  │  Widget           │  No direct OpenAI connection — talks only     │
│  │                   │  to your Python backend                       │
│  └────────┬─────────┘                                                │
└───────────┼──────────────────────────────────────────────────────────┘
            │ HTTPS (via Caddy)
            ▼
┌────────────────────────────────────────────────────────────────┐
│  ChatKit Python SDK Server (FastAPI)           :8000           │
│                                                                │
│  ┌─────────────────┐  ┌──────────────────┐  ┌──────────────┐ │
│  │ ChatKitServer    │  │ OpenAI Agents SDK│  │ MSAL / Entra │ │
│  │ - respond()      │  │ - LLM inference  │  │ - OIDC login │ │
│  │ - thread mgmt    │  │ - tool routing   │  │ - OBO token  │ │
│  │ - SSE streaming  │  │ - chat history   │  │   exchange   │ │
│  └────────┬────────┘  └────────┬─────────┘  └──────┬───────┘ │
│           │                    │                     │         │
│           │     Agent calls MCP tool ──────┐        │         │
│           │                                │        │         │
└───────────┼────────────────────────────────┼────────┼─────────┘
            │                                │        │
            │                                ▼        │
            │                    ┌──────────────────┐ │
            │                    │ MCP Client (httpx)│ │
            │                    │ Bearer: <token>   │─┘ (token from Entra OBO)
            │                    └────────┬─────────┘
            │                             │ MCP Streamable HTTP
            │                             ▼
            │              ┌────────────────────────┐
            │              │  Go MCP Server          │
            │              │  (go-sdk PR #770)       │
            │              │  :3001                   │
            │              │                          │
            │              │  - Echo, GetTime tools   │
            │              │  - JWT validation        │  ◄── Entra JWKS
            │              │  - /.well-known/oauth-   │
            │              │    protected-resource    │
            │              │  - JWT bearer grant      │
            │              └──────────────────────────┘
            │
            ▼
┌──────────────────────────┐
│   Microsoft Entra ID     │
│   (Free Azure tenant)    │
│                          │
│   - User SSO (OIDC)     │
│   - OBO token exchange   │
│   - Policy enforcement   │
│   - JWKS endpoint        │
└──────────────────────────┘
```

### What Each Component Does

|Component                |Role             |Manages                                                          |
|-------------------------|-----------------|-----------------------------------------------------------------|
|ChatKit JS (browser)     |Thin UI shell    |Renders chat, streams responses, sends user input                |
|ChatKit Python SDK Server|**The brain**    |Auth, chat history, thread state, agent orchestration, MCP client|
|OpenAI Agents SDK        |LLM layer        |Decides when to call tools, generates responses                  |
|Go MCP Server            |Tool provider    |Exposes tools, validates JWTs, executes tool logic               |
|Entra ID                 |Identity provider|SSO, token exchange, access policy, JWKS                         |

### Key Design Decisions

- **No Agent Builder workflow** — your Python server IS the backend. ChatKit JS connects directly to it.
- **No OpenAI-hosted backend** — all chat state, history, and connectivity managed by you.
- **Agents SDK for inference only** — the Python SDK’s `stream_agent_response` helper wires the agent’s tool calls to your MCP tools.
- **Python is the MCP client** — the Python backend authenticates to the Go MCP server using Entra OBO tokens, then forwards tool calls from the agent.
- **In-memory thread store** — minimal; no database needed for the POC.

-----

## Phase 0: Accounts & Prerequisites

### 0.1 — OpenAI Account

- [ ] Go to https://platform.openai.com, create account or sign in
- [ ] Add a payment method (Agents SDK requires API access)
- [ ] Generate an API key: API Keys → Create new secret key
- [ ] Note the key — used by the Python backend for Agents SDK inference

### 0.2 — Azure Free Account + Entra Tenant

- [ ] Open an **incognito browser window**
- [ ] Go to https://azure.microsoft.com/free
- [ ] Sign up with a **personal Microsoft account** (outlook.com / hotmail.com)
  - If you don’t have one, create one at https://outlook.live.com first
  - Do NOT use your work email (it attaches to your company’s tenant)
- [ ] Complete signup (credit card required, no charges for free tier)
- [ ] Go to https://portal.azure.com
- [ ] Search “Microsoft Entra ID” → you should land in your own tenant
- [ ] Note your **Tenant ID** (Overview → Properties → Tenant ID)
- [ ] Note your tenant domain (e.g., `youralias.onmicrosoft.com`)
- [ ] Create a test user: Users → New user → Create new user
  - Username: `testuser@youralias.onmicrosoft.com`
  - Set a password you’ll remember
  - This simulates an enterprise employee

### 0.3 — Local Dev Tools (Windows)

- [ ] Docker Desktop for Windows installed and running
- [ ] Go 1.23+ installed
- [ ] Python 3.11+ installed
- [ ] Git installed
- [ ] A code editor

-----

## Phase 1: Entra ID App Registrations

Two app registrations: one for the Python backend (MCP client), one for the Go MCP server (resource server).

### 1.1 — Register the Go MCP Server App (Resource Server)

- [ ] Entra admin center → Identity → Applications → App registrations → New registration
- [ ] Name: `MCP Server (Homelab)`
- [ ] Supported account types: “Accounts in this organizational directory only”
- [ ] No redirect URI (resource server doesn’t handle login)
- [ ] Click Register
- [ ] **Note: Application (client) ID** → this is your MCP server’s audience
- [ ] **Note: Directory (tenant) ID**
- [ ] Go to “Expose an API”
  - [ ] Set Application ID URI → accept default `api://<client-id>` or set custom
  - [ ] Add a scope:
    - Scope name: `mcp.tools`
    - Who can consent: Admins and users
    - Display name: “Access MCP tools”
    - State: Enabled
  - [ ] **Note the full scope**: `api://<mcp-server-client-id>/mcp.tools`

### 1.2 — Register the Python Backend App (MCP Client)

- [ ] New registration
- [ ] Name: `ChatKit Backend (Homelab)`
- [ ] Supported account types: “Accounts in this organizational directory only”
- [ ] Redirect URI: Web → `https://localhost/auth/callback`
- [ ] Click Register
- [ ] **Note: Application (client) ID**
- [ ] Certificates & secrets → New client secret
  - Description: `homelab-dev`
  - **Note the secret value immediately**
- [ ] API permissions → Add a permission:
  - My APIs → “MCP Server (Homelab)” → Delegated → `mcp.tools` → Add
  - Microsoft Graph → Delegated → `openid`, `profile`, `email`, `offline_access` → Add
  - Click **“Grant admin consent for [tenant]”**

### 1.3 — Pre-Authorize the Backend for OBO

- [ ] Go back to the MCP Server app registration
- [ ] “Expose an API” → Authorized client applications → Add
  - Client ID: paste the ChatKit Backend’s client ID
  - Check the `mcp.tools` scope
- [ ] This pre-authorizes the Python backend to exchange tokens for MCP server access

-----

## Phase 2: Go MCP Server (PR #770)

### 2.1 — Clone and Build

- [ ] Clone the fork with enterprise auth:
  
  ```powershell
  git clone https://github.com/radar07/go-sdk.git
  cd go-sdk
  git checkout enterprise-managed-authorization
  git log --oneline -5  # note the HEAD commit hash to pin later
  ```
- [ ] Verify it builds: `go build ./...`

### 2.2 — Create Your MCP Server

- [ ] Create project directory:
  
  ```powershell
  mkdir homelab-mcp-server
  cd homelab-mcp-server
  go mod init homelab-mcp-server
  ```
- [ ] Add replace directive in `go.mod`:
  
  ```
  replace github.com/modelcontextprotocol/go-sdk => ../go-sdk
  ```
  
  Then run `go mod tidy`
- [ ] Create `main.go` implementing:
  
  **a) Example tools** (keep it dead simple):
  - `echo` — returns whatever you send it
  - `get_server_time` — returns current server time
  - `whoami` — extracts and returns the authenticated user’s claims from the JWT (proves auth works end-to-end)
  
  **b) Protected Resource Metadata endpoint**:
  
  ```
  GET /.well-known/oauth-protected-resource
  →
  {
    "resource": "http://localhost:3001",
    "authorization_servers": [
      "https://login.microsoftonline.com/{tenant-id}/v2.0"
    ],
    "scopes_supported": ["mcp.tools"],
    "bearer_methods_supported": ["header"]
  }
  ```
  
  **c) JWT validation middleware**:
  - Fetch JWKS from `https://login.microsoftonline.com/{tenant-id}/discovery/v2.0/keys`
  - Validate: signature, `iss` (must be your tenant), `aud` (must be your MCP server app ID), `exp`, `scp` contains `mcp.tools`
  - Extract user claims (`preferred_username`, `name`, `sub`) and make available to tool handlers
  
  **d) Streamable HTTP transport** on `:3001`
  
  **e) JWT bearer grant endpoint** (from PR #770’s `oauthex/jwt_bearer.go`):
  - Accepts ID-JAG or access tokens
  - Validates against Entra’s JWKS
  - Issues MCP access token (or pass-through if using OBO tokens directly)
- [ ] Test locally:
  
  ```powershell
  set ENTRA_TENANT_ID=your-tenant-id
  set MCP_SERVER_CLIENT_ID=your-mcp-server-client-id
  go run main.go
  ```
- [ ] Verify metadata endpoint:
  
  ```powershell
  curl http://localhost:3001/.well-known/oauth-protected-resource
  ```
- [ ] Verify unauthenticated request returns 401:
  
  ```powershell
  curl http://localhost:3001/mcp
  # Should get 401 with WWW-Authenticate header
  ```

### 2.3 — Dockerfile

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go-sdk/ ./go-sdk/
COPY homelab-mcp-server/ ./homelab-mcp-server/
WORKDIR /src/homelab-mcp-server
RUN go build -o /mcp-server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=build /mcp-server /mcp-server
EXPOSE 3001
ENTRYPOINT ["/mcp-server"]
```

-----

## Phase 3: ChatKit Python SDK Server

This is the core component. It’s a self-hosted ChatKit server that:

- Handles user authentication via Entra OIDC
- Manages chat threads and history (in-memory)
- Runs an OpenAI Agent that can call MCP tools
- Acts as the MCP client, authenticating to the Go MCP server with Entra OBO tokens

### 3.1 — Project Setup

```powershell
mkdir chatkit-backend
cd chatkit-backend
python -m venv .venv
.venv\Scripts\activate
pip install openai-chatkit openai-agents fastapi uvicorn httpx msal python-jose[cryptography]
pip freeze > requirements.txt
```

Package purposes:

- `openai-chatkit` — ChatKit Python SDK (ChatKitServer, thread management, SSE streaming)
- `openai-agents` — OpenAI Agents SDK (agent orchestration, tool calling)
- `fastapi` / `uvicorn` — Web framework + ASGI server
- `httpx` — Async HTTP client (for MCP server calls)
- `msal` — Microsoft Authentication Library (OIDC + OBO token exchange)
- `python-jose` — JWT decoding/validation

### 3.2 — Configuration (`config.py`)

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

### 3.3 — Entra Auth Module (`entra_auth.py`)

Implement these functions:

- [ ] **`get_msal_app()`** — Creates an MSAL `ConfidentialClientApplication` with your client ID/secret
- [ ] **`build_login_url(state)`** — Generates the Entra OIDC auth URL with scopes `openid profile email` + `api://<mcp-server>/mcp.tools`
  - Request the MCP server scope during initial login so you can do OBO later
- [ ] **`handle_callback(auth_code)`** — Exchanges auth code for tokens via MSAL’s `acquire_token_by_authorization_code`
  - Returns: ID token, access token (for Graph), and the user’s access token scoped to the MCP server
- [ ] **`get_mcp_token(user_assertion)`** — OBO token exchange:
  
  ```python
  result = msal_app.acquire_token_on_behalf_of(
      user_assertion=user_assertion,
      scopes=[MCP_SERVER_SCOPE]
  )
  return result["access_token"]
  ```
  
  This is the key function — takes the user’s token and gets back a token scoped to `api://<mcp-server>/mcp.tools`

### 3.4 — MCP Client Wrapper (`mcp_client.py`)

- [ ] Implement an async function that calls the Go MCP server:
  
  ```python
  async def call_mcp_tool(tool_name: str, arguments: dict, bearer_token: str) -> dict:
      async with httpx.AsyncClient() as client:
          # Use MCP Streamable HTTP protocol
          response = await client.post(
              f"{MCP_SERVER_URL}/mcp",
              json={
                  "jsonrpc": "2.0",
                  "method": "tools/call",
                  "params": {"name": tool_name, "arguments": arguments},
                  "id": 1
              },
              headers={"Authorization": f"Bearer {bearer_token}"}
          )
          return response.json()
  ```
- [ ] Or use the official MCP Python SDK’s `ClientSession` with `StreamableHTTPTransport` for a more proper implementation (add `mcp` to requirements if so)

### 3.5 — Agent with MCP Tools (`agent.py`)

- [ ] Define Agents SDK tools that wrap MCP calls:
  
  ```python
  from agents import Agent, function_tool, RunContextWrapper
  
  @function_tool
  async def echo(ctx: RunContextWrapper, message: str) -> str:
      """Echo a message through the MCP server"""
      token = ctx.context["mcp_token"]  # Entra OBO token
      result = await call_mcp_tool("echo", {"message": message}, token)
      return result["result"]["content"][0]["text"]
  
  @function_tool
  async def get_server_time(ctx: RunContextWrapper) -> str:
      """Get the current time from the MCP server"""
      token = ctx.context["mcp_token"]
      result = await call_mcp_tool("get_server_time", {}, token)
      return result["result"]["content"][0]["text"]
  
  @function_tool
  async def whoami(ctx: RunContextWrapper) -> str:
      """Get the authenticated user info from the MCP server"""
      token = ctx.context["mcp_token"]
      result = await call_mcp_tool("whoami", {}, token)
      return result["result"]["content"][0]["text"]
  
  mcp_agent = Agent(
      name="MCP Assistant",
      instructions="You help users interact with the MCP server tools. Use the available tools when asked.",
      tools=[echo, get_server_time, whoami],
      model="gpt-4o-mini"
  )
  ```

### 3.6 — ChatKit Server (`chatkit_server.py`)

- [ ] Implement the minimal ChatKitServer:
  
  ```python
  from openai_chatkit import ChatKitServer, InMemoryStore
  from openai_chatkit.streams import stream_agent_response
  
  class HomelabChatKitServer(ChatKitServer):
      async def respond(self, thread, input, context):
          # Get the MCP token from the user's session
          mcp_token = get_mcp_token_for_user(thread.user_id)
  
          # Run the agent with MCP token in context
          agent_context = {"mcp_token": mcp_token}
  
          async for event in stream_agent_response(
              agent=mcp_agent,
              input=input,
              thread=thread,
              context=agent_context
          ):
              yield event
  
  server = HomelabChatKitServer(
      store=InMemoryStore()  # threads + messages stored in memory
  )
  ```

### 3.7 — FastAPI App (`app.py`)

- [ ] Wire everything together:
  
  ```python
  from fastapi import FastAPI, Request, Response
  from fastapi.responses import RedirectResponse, HTMLResponse
  from starlette.middleware.sessions import SessionMiddleware
  
  app = FastAPI()
  app.add_middleware(SessionMiddleware, secret_key="change-me-in-prod")
  
  # --- Auth routes ---
  
  @app.get("/login")
  async def login(request: Request):
      url = build_login_url(state="random-state")
      return RedirectResponse(url)
  
  @app.get("/auth/callback")
  async def auth_callback(request: Request):
      code = request.query_params.get("code")
      tokens = handle_callback(code)
      request.session["user"] = tokens["id_token_claims"]
      request.session["access_token"] = tokens["access_token"]
      return RedirectResponse("/")
  
  # --- ChatKit routes ---
  
  @app.post("/api/chat")
  async def chat(request: Request):
      """ChatKit Python SDK handles this endpoint"""
      if "user" not in request.session:
          return Response(status_code=401)
      # Forward to ChatKit server
      return await chatkit_server.handle_request(request)
  
  # --- Frontend ---
  
  @app.get("/")
  async def index(request: Request):
      if "user" not in request.session:
          return RedirectResponse("/login")
      return HTMLResponse(open("static/index.html").read())
  ```
  
  Note: The exact ChatKit SDK server integration with FastAPI may differ — check the `openai-chatkit` docs for the correct `handle_request` or ASGI mount pattern. The advanced samples repo shows the FastAPI wiring.

### 3.8 — Dockerfile

```dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8000"]
```

-----

## Phase 4: ChatKit Frontend

### 4.1 — Minimal HTML (`chatkit-backend/static/index.html`)

The frontend is dead simple because your Python server handles everything.

```html
<!DOCTYPE html>
<html>
<head>
  <title>Homelab ChatKit + Enterprise Auth</title>
  <script src="https://cdn.platform.openai.com/deployments/chatkit/chatkit.js" async></script>
  <style>
    body { margin: 0; font-family: sans-serif; }
    #chat { height: 100vh; }
  </style>
</head>
<body>
  <div id="chat"></div>
  <script>
    const el = document.getElementById('chat');
    const widget = document.createElement('chatkit-widget');

    // Self-hosted mode: point directly at your Python backend
    // No OpenAI client secret needed — your server handles inference
    widget.setOptions({
      api: {
        url: '/api/chat',  // your FastAPI endpoint
      }
    });

    el.appendChild(widget);
  </script>
</body>
</html>
```

Note: The exact `setOptions` shape for self-hosted mode may differ from hosted mode. In self-hosted mode you don’t use `getClientSecret` — the ChatKit JS talks directly to your server’s endpoint. Check the ChatKit JS docs for the self-hosted API config.

-----

## Phase 5: TLS & Docker Compose

### 5.1 — Caddyfile

```
https://localhost {
  handle /api/* {
    reverse_proxy chatkit-backend:8000
  }
  handle /login {
    reverse_proxy chatkit-backend:8000
  }
  handle /auth/* {
    reverse_proxy chatkit-backend:8000
  }
  handle /* {
    reverse_proxy chatkit-backend:8000
  }
  tls internal
}
```

### 5.2 — docker-compose.yml

```yaml
services:
  caddy:
    image: caddy:2-alpine
    ports:
      - "443:443"
      - "80:80"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - chatkit-backend

  chatkit-backend:
    build: ./chatkit-backend
    environment:
      - ENTRA_TENANT_ID=${ENTRA_TENANT_ID}
      - ENTRA_CLIENT_ID=${ENTRA_CLIENT_ID}
      - ENTRA_CLIENT_SECRET=${ENTRA_CLIENT_SECRET}
      - MCP_SERVER_SCOPE=${MCP_SERVER_SCOPE}
      - MCP_SERVER_URL=http://mcp-server:3001
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - REDIRECT_URI=https://localhost/auth/callback
    depends_on:
      - mcp-server

  mcp-server:
    build:
      context: .
      dockerfile: homelab-mcp-server/Dockerfile
    environment:
      - ENTRA_TENANT_ID=${ENTRA_TENANT_ID}
      - MCP_SERVER_CLIENT_ID=${MCP_SERVER_CLIENT_ID}
    expose:
      - "3001"

volumes:
  caddy_data:
  caddy_config:
```

### 5.3 — .env

```env
ENTRA_TENANT_ID=your-tenant-id
ENTRA_CLIENT_ID=chatkit-backend-client-id
ENTRA_CLIENT_SECRET=chatkit-backend-secret-value
MCP_SERVER_SCOPE=api://mcp-server-client-id/mcp.tools
MCP_SERVER_CLIENT_ID=mcp-server-client-id
OPENAI_API_KEY=sk-...
```

-----

## Phase 6: Bring Up & Validate

### 6.1 — Start

```powershell
docker compose up --build
```

Watch for:

- `caddy`: “serving initial configuration”
- `chatkit-backend`: “Uvicorn running on 0.0.0.0:8000”
- `mcp-server`: “MCP server listening on :3001”

### 6.2 — Auth Flow

- [ ] Open `https://localhost` → accept self-signed cert
- [ ] Should redirect to `/login` → Entra login page
- [ ] Sign in with your test user (`testuser@youralias.onmicrosoft.com`)
- [ ] After login → redirect back to ChatKit UI
- [ ] Backend logs should show: OIDC callback success, tokens received

### 6.3 — Token Exchange

- [ ] Backend logs should show OBO token exchange:
  - “OBO exchange successful — got access token for MCP server”
- [ ] If it fails, check:
  - Admin consent granted for `mcp.tools` scope?
  - Backend is listed as authorized client of MCP server app?
  - Client ID and secret correct?

### 6.4 — MCP Tool Calls

- [ ] In the ChatKit widget, type: “What time is it?”
- [ ] Agent should call `get_server_time` tool
- [ ] Go MCP server logs: “JWT validation OK — user: testuser@…, tool: get_server_time”
- [ ] Response appears in chat
- [ ] Type: “Who am I?”
- [ ] Agent calls `whoami` → returns authenticated user’s claims from the JWT
- [ ] **This is the money shot** — proves the full chain: browser → Entra SSO → Python backend → OBO token → Go MCP server → JWT validation → user identity extracted
- [ ] Type: “Echo hello world”
- [ ] Agent calls `echo` → Go server returns “hello world”

### 6.5 — Authorization Denial

- [ ] In Entra → Enterprise applications → “MCP Server (Homelab)”
  - Set “Assignment required” = Yes
  - Remove test user from assignments
- [ ] Refresh ChatKit, try a tool call → should fail (OBO returns error)
- [ ] Re-add user → works again
- [ ] This proves Entra controls access to MCP server

-----

## Phase 7: Validation Checklist

### Auth

- [ ] User authenticates via Entra (real Microsoft login page)
- [ ] ID Token has correct claims (`iss`, `sub`, `preferred_username`)
- [ ] OBO exchange produces token with `aud` = MCP server app ID
- [ ] OBO token contains `scp: mcp.tools`

### Enterprise Policy

- [ ] Admin can revoke access by removing user assignment
- [ ] Revoked users get auth errors (not tool errors)
- [ ] All events visible in Entra sign-in logs

### MCP Protocol

- [ ] `/.well-known/oauth-protected-resource` returns correct metadata
- [ ] Unauthenticated requests get 401 with `WWW-Authenticate` header
- [ ] Valid bearer tokens grant access to tools
- [ ] Expired/invalid tokens get 401

### End-to-End

- [ ] `whoami` tool returns the actual Entra user identity
- [ ] Chat history persists within a session (in-memory thread store)
- [ ] Agent decides when to call tools vs respond directly
- [ ] All connectivity managed server-side (no direct browser-to-OpenAI calls for chat)

-----

## Known Gotchas & Workarounds

1. **Azure free account — use personal email.** Signing up with a work email attaches to your company’s tenant. Use a personal outlook.com account in incognito.
1. **PR #770 is under review.** Pin to a specific commit. The reviewer is working on a parallel client-side OAuth refactor (PR #785). Server-side JWT validation is stable; client-side abstractions may shift.
1. **Entra OBO vs full ID-JAG.** The spec describes RFC 8693 token exchange producing an `id-jag` token type. Entra’s native support for this exact type is still evolving. OBO (`grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer` with `requested_token_use=on_behalf_of`) is fully supported and achieves the same result. Use OBO.
1. **ChatKit Python SDK self-hosted wiring.** The exact FastAPI integration pattern may differ from what’s shown here. Reference the `openai/openai-chatkit-advanced-samples` repo — each example shows the FastAPI ↔ ChatKitServer wiring in detail.
1. **Agents SDK tool context.** Passing the MCP token through the agent’s `RunContext` requires checking the current Agents SDK docs for the exact pattern. The context dict approach shown above is illustrative — the actual API may use `RunContextWrapper` or similar.
1. **MCP client in Python.** For a more spec-compliant MCP client (vs raw httpx), install the official `mcp` Python package and use `ClientSession` with `StreamableHTTPTransport`. This handles JSON-RPC framing, capabilities negotiation, etc.
1. **Docker networking on Windows.** Container-to-container via service names works in Compose. For host access use `localhost` with published ports. Caddy handles TLS termination.
1. **No DCR anywhere in this flow.** The enterprise auth pattern uses pre-registered apps in Entra + OBO token exchange. DCR is not involved and not needed.

-----

## File Structure

```
homelab-chatkit-mcp/
├── docker-compose.yml
├── .env
├── Caddyfile
├── go-sdk/                      # cloned from radar07/go-sdk
│   └── (enterprise-managed-authorization branch)
├── homelab-mcp-server/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
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