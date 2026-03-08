from __future__ import annotations

from collections.abc import AsyncIterator
from typing import Any

from agents import Runner
from chatkit.agents import AgentContext, ThreadItemConverter, stream_agent_response
from chatkit.server import ChatKitServer
from chatkit.types import ThreadMetadata, ThreadStreamEvent, UserMessageItem

from agent import mcp_agent
from memory_store import MemoryStore


thread_item_converter = ThreadItemConverter()


class HomelabChatKitServer(ChatKitServer[dict[str, Any]]):
    def __init__(self):
        self.store = MemoryStore()
        super().__init__(self.store)

    async def respond(
        self,
        thread: ThreadMetadata,
        user_message: UserMessageItem | None,
        context: dict[str, Any],
    ) -> AsyncIterator[ThreadStreamEvent]:
        # Load thread history and convert to agent input
        items_page = await self.store.load_thread_items(
            thread.id, None, 20, "desc", context,
        )
        items = list(reversed(items_page.data))
        input_items = await thread_item_converter.to_agent_input(items)

        # Create ChatKit agent context with our request context (contains mcp_token)
        agent_context = AgentContext(
            thread=thread,
            store=self.store,
            request_context=context,
        )

        # Run the agent with streaming
        result = Runner.run_streamed(
            mcp_agent,
            input_items,
            context=agent_context,
        )

        async for event in stream_agent_response(agent_context, result):
            yield event


server = HomelabChatKitServer()
