import os

ENTRA_TENANT_ID = os.environ["ENTRA_TENANT_ID"]
ENTRA_CLIENT_ID = os.environ["ENTRA_CLIENT_ID"]
ENTRA_CLIENT_SECRET = os.environ["ENTRA_CLIENT_SECRET"]
MCP_SERVER_SCOPE = os.environ["MCP_SERVER_SCOPE"]
MCP_SERVER_URL = os.environ.get("MCP_SERVER_URL", "http://mcp-server:3001")
OPENAI_API_KEY = os.environ["OPENAI_API_KEY"]
REDIRECT_URI = os.environ.get("REDIRECT_URI", "https://localhost/auth/callback")
AUTHORITY = f"https://login.microsoftonline.com/{ENTRA_TENANT_ID}"
