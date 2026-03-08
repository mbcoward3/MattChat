import secrets

from chatkit.server import StreamingResult
from fastapi import FastAPI, Request
from fastapi.responses import HTMLResponse, RedirectResponse, Response, StreamingResponse
from starlette.middleware.sessions import SessionMiddleware
from starlette.responses import JSONResponse

from chatkit_server import server as chatkit_server
from config import AUTHORITY
from entra_auth import build_login_url, handle_callback

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
    return RedirectResponse(
        f"{AUTHORITY}/oauth2/v2.0/logout"
        f"?post_logout_redirect_uri=https://localhost/login"
    )


# --- ChatKit route ---


@app.post("/api/chat")
async def chat(request: Request):
    if "user" not in request.session:
        return Response(status_code=401)

    mcp_token = request.session.get("access_token")

    body = await request.body()
    context = {"mcp_token": mcp_token, "request": request}
    result = await chatkit_server.process(body, context)

    if isinstance(result, StreamingResult):
        return StreamingResponse(result, media_type="text/event-stream")
    return Response(content=result.json, media_type="application/json")


# --- Frontend ---


@app.get("/")
async def index(request: Request):
    if "user" not in request.session:
        return RedirectResponse("/login")
    with open("static/index.html") as f:
        return HTMLResponse(f.read())
