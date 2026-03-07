# Step 9: Python Backend — FastAPI App

## File: `chatkit-backend/app.py`

Wires auth, ChatKit server, and static files into a FastAPI application.

### Routes

#### Auth Routes

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/login` | Generates Entra OIDC login URL, redirects browser |
| `GET` | `/auth/callback` | Receives auth code from Entra, exchanges for tokens, stores in session |
| `GET` | `/logout` | Clears session, redirects to `/login` |

#### ChatKit Route

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/api/chat` | Auth-gated; delegates to ChatKit server's request handler |

#### Frontend Route

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/` | Serves `static/index.html` if authenticated; redirects to `/login` if not |

### Middleware

- `SessionMiddleware` from Starlette — stores user tokens in encrypted server-side session
- Session secret key from env or a generated default

### Auth Flow Detail

1. User hits `/` → no session → redirect to `/login`
2. `/login` → builds Entra OIDC URL → browser redirects to Microsoft login
3. User authenticates → Entra redirects to `/auth/callback?code=...`
4. `/auth/callback` → exchanges code for tokens via `entra_auth.handle_callback()`
5. Stores in session: `id_token_claims` (user info), `access_token` (for OBO)
6. Redirects to `/` → session exists → serves ChatKit UI
7. On chat message → `/api/chat` → retrieves `access_token` from session → calls `entra_auth.get_mcp_token()` for OBO → passes MCP token to ChatKit server

### ChatKit Integration

The exact method to forward requests to the ChatKit server depends on the `openai-chatkit` API:
- Option A: `server.handle_request(request)` — if the SDK provides a FastAPI-compatible handler
- Option B: Mount the ChatKit server as an ASGI sub-application
- Option C (fallback): If `openai-chatkit` doesn't work, implement SSE streaming directly with `StreamingResponse` and the Agents SDK `Runner.run_streamed()`
