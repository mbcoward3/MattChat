from openai_chatkit import ChatKitServer, InMemoryStore
from openai_chatkit.streams import stream_agent_response
from agent import mcp_agent


class HomelabChatKitServer(ChatKitServer):
    async def respond(self, thread, input, context):
        mcp_token = context.get("mcp_token")
        agent_context = {"mcp_token": mcp_token}

        async for event in stream_agent_response(
            agent=mcp_agent,
            input=input,
            thread=thread,
            context=agent_context,
        ):
            yield event


server = HomelabChatKitServer(store=InMemoryStore())
