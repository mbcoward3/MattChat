import logging

from mcp import ClientSession
from mcp.client.streamable_http import streamablehttp_client
from config import MCP_SERVER_URL

logger = logging.getLogger(__name__)


async def call_mcp_tool(tool_name: str, arguments: dict, bearer_token: str) -> str:
    try:
        logger.info(f"Calling MCP tool '{tool_name}' at {MCP_SERVER_URL}/mcp")
        async with streamablehttp_client(
            f"{MCP_SERVER_URL}/mcp",
            headers={"Authorization": f"Bearer {bearer_token}"},
        ) as (read, write, _):
            async with ClientSession(read, write) as session:
                await session.initialize()
                result = await session.call_tool(tool_name, arguments)
                logger.info(f"MCP tool '{tool_name}' returned: {result}")
                return result.content[0].text
    except Exception as e:
        logger.error(f"MCP tool '{tool_name}' failed: {e}", exc_info=True)
        raise
