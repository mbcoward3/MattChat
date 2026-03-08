import secrets

from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, RedirectResponse, Response
from starlette.middleware.sessions import SessionMiddleware

from chatkit_server import server as chatkit_server
from entra_auth import build_login_url, get_mcp_token, handle_callback

app = FastAPI()
app.add_middleware(SessionMiddleware, secret_key=secrets.token_hex(32))


# --- Auth routes ---


@app.get("/login")
async def login(request: Request):
    state = secrets.token_urlsafe(16)
    request.session["oauth_state"] = state
    url = build_login_url(state=state)
    return RedirectResponse(url)


@app.get("/auth/callback")
async def auth_callback(request: Request):
    code = request.query_params.get("code")
    if not code:
        return Response("Missing auth code", status_code=400)

    tokens = handle_callback(code)
    request.session["user"] = tokens["id_token_claims"]
    request.session["access_token"] = tokens["access_token"]
    return RedirectResponse("/")


@app.get("/logout")
async def logout(request: Request):
    request.session.clear()
    return RedirectResponse("/login")


# --- ChatKit route ---


@app.post("/api/chat")
async def chat(request: Request):
    if "user" not in request.session:
        return Response(status_code=401)

    access_token = request.session.get("access_token")
    mcp_token = get_mcp_token(access_token)

    # Pass the MCP token as context for the ChatKit server
    request.state.chatkit_context = {"mcp_token": mcp_token}
    return await chatkit_server.handle_request(request)


# --- Frontend ---


@app.get("/")
async def index(request: Request):
    if "user" not in request.session:
        return RedirectResponse("/login")
    with open("static/index.html") as f:
        return HTMLResponse(f.read())
