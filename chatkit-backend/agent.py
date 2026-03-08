from agents import Agent, function_tool, RunContextWrapper
from mcp_client import call_mcp_tool


@function_tool
async def echo(ctx: RunContextWrapper, message: str) -> str:
    """Echo a message through the MCP server"""
    token = ctx.context.request_context["mcp_token"]
    return await call_mcp_tool("echo", {"message": message}, token)


@function_tool
async def get_server_time(ctx: RunContextWrapper) -> str:
    """Get the current time from the MCP server"""
    token = ctx.context.request_context["mcp_token"]
    return await call_mcp_tool("get_server_time", {}, token)


@function_tool
async def whoami(ctx: RunContextWrapper) -> str:
    """Get the authenticated user info from the MCP server"""
    token = ctx.context.request_context["mcp_token"]
    return await call_mcp_tool("whoami", {}, token)


@function_tool
async def save_note(ctx: RunContextWrapper, title: str, content: str) -> str:
    """Save a personal note (scoped to the authenticated user)"""
    token = ctx.context.request_context["mcp_token"]
    return await call_mcp_tool("save_note", {"title": title, "content": content}, token)


@function_tool
async def list_notes(ctx: RunContextWrapper) -> str:
    """List all personal notes for the authenticated user"""
    token = ctx.context.request_context["mcp_token"]
    return await call_mcp_tool("list_notes", {}, token)


@function_tool
async def delete_note(ctx: RunContextWrapper, title: str) -> str:
    """Delete a personal note by title"""
    token = ctx.context.request_context["mcp_token"]
    return await call_mcp_tool("delete_note", {"title": title}, token)


mcp_agent = Agent(
    name="MCP Assistant",
    instructions=(
        "You help users interact with the MCP server tools. "
        "You can echo messages, check the time, identify the user, "
        "and manage personal notes. Use the available tools when asked."
    ),
    tools=[echo, get_server_time, whoami, save_note, list_notes, delete_note],
    model="gpt-4o-mini",
)
