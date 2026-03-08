import msal
from config import (
    ENTRA_CLIENT_ID,
    ENTRA_CLIENT_SECRET,
    AUTHORITY,
    MCP_SERVER_SCOPE,
    REDIRECT_URI,
)

_msal_app = None


def get_msal_app() -> msal.ConfidentialClientApplication:
    global _msal_app
    if _msal_app is None:
        _msal_app = msal.ConfidentialClientApplication(
            client_id=ENTRA_CLIENT_ID,
            client_credential=ENTRA_CLIENT_SECRET,
            authority=AUTHORITY,
        )
    return _msal_app


def build_login_url(state: str) -> str:
    app = get_msal_app()
    return app.get_authorization_request_url(
        scopes=["openid", "profile", "email", MCP_SERVER_SCOPE],
        redirect_uri=REDIRECT_URI,
        state=state,
    )


def handle_callback(auth_code: str) -> dict:
    app = get_msal_app()
    result = app.acquire_token_by_authorization_code(
        code=auth_code,
        scopes=["openid", "profile", "email", MCP_SERVER_SCOPE],
        redirect_uri=REDIRECT_URI,
    )
    if "error" in result:
        raise RuntimeError(f"Auth error: {result['error']} - {result.get('error_description', '')}")
    return {
        "id_token_claims": result.get("id_token_claims"),
        "access_token": result.get("access_token"),
    }


def get_mcp_token(user_assertion: str) -> str:
    app = get_msal_app()
    result = app.acquire_token_on_behalf_of(
        user_assertion=user_assertion,
        scopes=[MCP_SERVER_SCOPE],
    )
    if "error" in result:
        raise RuntimeError(f"OBO error: {result['error']} - {result.get('error_description', '')}")
    return result["access_token"]
