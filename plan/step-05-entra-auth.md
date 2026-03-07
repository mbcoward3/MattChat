# Step 5: Python Backend — Auth Layer

## File: `chatkit-backend/entra_auth.py`

Handles all Entra ID authentication: OIDC login flow and OBO token exchange.

### Functions

#### `get_msal_app() -> msal.ConfidentialClientApplication`

Creates and returns an MSAL confidential client app configured with:
- `client_id` = `ENTRA_CLIENT_ID`
- `client_credential` = `ENTRA_CLIENT_SECRET`
- `authority` = `AUTHORITY` (derived from tenant ID)

Can be cached as a module-level singleton.

#### `build_login_url(state: str) -> str`

Generates the Entra OIDC authorization URL via `msal_app.get_authorization_request_url()`:
- Scopes: `["openid", "profile", "email", MCP_SERVER_SCOPE]`
  - Including `MCP_SERVER_SCOPE` during initial login ensures the access token can be used for OBO later
- `redirect_uri` = `REDIRECT_URI`
- `state` = passed in (for CSRF protection)

Returns the full URL to redirect the user's browser to.

#### `handle_callback(auth_code: str) -> dict`

Exchanges the authorization code for tokens via `msal_app.acquire_token_by_authorization_code()`:
- `code` = `auth_code`
- `scopes` = `["openid", "profile", "email", MCP_SERVER_SCOPE]`
- `redirect_uri` = `REDIRECT_URI`

Returns dict with:
- `id_token_claims` — decoded ID token (user info)
- `access_token` — the access token (used as user assertion for OBO)

Raises on error (check `result.get("error")`).

#### `get_mcp_token(user_assertion: str) -> str`

Performs OBO (On-Behalf-Of) token exchange:
```python
result = msal_app.acquire_token_on_behalf_of(
    user_assertion=user_assertion,
    scopes=[MCP_SERVER_SCOPE]
)
```

Returns `result["access_token"]` — a token scoped to `api://<mcp-server>/mcp.tools` that the Go MCP server will accept.

This is the critical function: it takes the user's access token and gets back a token the MCP server trusts, carrying the user's identity.
