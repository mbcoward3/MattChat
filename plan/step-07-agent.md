# Step 7: Python Backend — Agent

## File: `chatkit-backend/agent.py`

Uses the `openai-agents` SDK to define an agent with six tool wrappers that delegate to the Go MCP server.

### Tool Definitions

Each tool is a `@function_tool` that:
1. Extracts the `mcp_token` from the agent's run context
2. Calls `mcp_client.call_mcp_tool()` with the token
3. Returns the result text

```python
from agents import Agent, function_tool, RunContextWrapper
from mcp_client import call_mcp_tool

@function_tool
async def echo(ctx: RunContextWrapper, message: str) -> str:
    """Echo a message through the MCP server"""
    token = ctx.context["mcp_token"]
    return await call_mcp_tool("echo", {"message": message}, token)

@function_tool
async def get_server_time(ctx: RunContextWrapper) -> str:
    """Get the current time from the MCP server"""
    token = ctx.context["mcp_token"]
    return await call_mcp_tool("get_server_time", {}, token)

@function_tool
async def whoami(ctx: RunContextWrapper) -> str:
    """Get the authenticated user info from the MCP server"""
    token = ctx.context["mcp_token"]
    return await call_mcp_tool("whoami", {}, token)

@function_tool
async def save_note(ctx: RunContextWrapper, title: str, content: str) -> str:
    """Save a personal note (scoped to the authenticated user)"""
    token = ctx.context["mcp_token"]
    return await call_mcp_tool("save_note", {"title": title, "content": content}, token)

@function_tool
async def list_notes(ctx: RunContextWrapper) -> str:
    """List all personal notes for the authenticated user"""
    token = ctx.context["mcp_token"]
    return await call_mcp_tool("list_notes", {}, token)

@function_tool
async def delete_note(ctx: RunContextWrapper, title: str) -> str:
    """Delete a personal note by title"""
    token = ctx.context["mcp_token"]
    return await call_mcp_tool("delete_note", {"title": title}, token)
```

### Agent Instance

```python
mcp_agent = Agent(
    name="MCP Assistant",
    instructions="You help users interact with the MCP server tools. You can echo messages, check the time, identify the user, and manage personal notes. Use the available tools when asked.",
    tools=[echo, get_server_time, whoami, save_note, list_notes, delete_note],
    model="gpt-4o-mini"
)
```

### Notes

- The `RunContextWrapper` API may differ from what's shown — verify against `openai-agents` docs after install
- `ctx.context` is a dict passed when running the agent, containing the user's MCP bearer token
- The agent decides when to call tools vs respond directly (LLM-driven tool selection)
- Notes are per-user (keyed by JWT `sub` claim on the Go server side) — the agent doesn't need to handle user scoping
