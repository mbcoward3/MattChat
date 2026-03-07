# Step 6: Python Backend — MCP Client

## File: `chatkit-backend/mcp_client.py`

Uses the official `mcp` Python package with `ClientSession` + `streamablehttp_client`.

### Function

#### `call_mcp_tool(tool_name: str, arguments: dict, bearer_token: str) -> str`

```python
from mcp import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from config import MCP_SERVER_URL

async def call_mcp_tool(tool_name: str, arguments: dict, bearer_token: str) -> str:
    async with streamablehttp_client(
        f"{MCP_SERVER_URL}/mcp",
        headers={"Authorization": f"Bearer {bearer_token}"}
    ) as (read, write, _):
        async with ClientSession(read, write) as session:
            await session.initialize()
            result = await session.call_tool(tool_name, arguments)
            # Extract text from the first content block
            return result.content[0].text
```

### Notes

- The `streamablehttp_client` handles JSON-RPC framing and MCP protocol negotiation automatically
- The bearer token is passed via headers — the Go MCP server's JWT middleware extracts and validates it
- Each call opens a new connection; for production, connection pooling or session reuse would improve performance, but for the POC this is sufficient
- Error handling: if the MCP server returns 401, the token is expired/invalid — surface this to the caller
