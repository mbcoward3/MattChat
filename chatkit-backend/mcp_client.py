from mcp import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from config import MCP_SERVER_URL


async def call_mcp_tool(tool_name: str, arguments: dict, bearer_token: str) -> str:
    async with streamablehttp_client(
        f"{MCP_SERVER_URL}/mcp",
        headers={"Authorization": f"Bearer {bearer_token}"},
    ) as (read, write, _):
        async with ClientSession(read, write) as session:
            await session.initialize()
            result = await session.call_tool(tool_name, arguments)
            return result.content[0].text
