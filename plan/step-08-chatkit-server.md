# Step 8: Python Backend — ChatKit Server

## File: `chatkit-backend/chatkit_server.py`

Subclasses `ChatKitServer` from `openai-chatkit` to wire the agent to the ChatKit thread model.

### Implementation

```python
from openai_chatkit import ChatKitServer, InMemoryStore
from openai_chatkit.streams import stream_agent_response
from agent import mcp_agent

class HomelabChatKitServer(ChatKitServer):
    async def respond(self, thread, input, context):
        # Get the MCP token stashed in the thread/session context
        mcp_token = context.get("mcp_token")

        agent_context = {"mcp_token": mcp_token}

        async for event in stream_agent_response(
            agent=mcp_agent,
            input=input,
            thread=thread,
            context=agent_context
        ):
            yield event

server = HomelabChatKitServer(
    store=InMemoryStore()
)
```

### Key Behaviors

- **Thread management**: `InMemoryStore` keeps threads and messages in memory — no database needed for the POC
- **Streaming**: `stream_agent_response` yields SSE events as the agent generates responses and makes tool calls
- **Context passing**: The MCP token flows from the user's session → ChatKit respond → agent context → tool functions → MCP client → Go server

### Caveat

The `openai-chatkit` package API (class names, method signatures, imports) may differ from what's shown. After installing in Step 13, verify:
- Import paths (`openai_chatkit` vs `chatkit`)
- `ChatKitServer` base class and `respond()` signature
- `stream_agent_response` helper location and params
- How `context` is passed to `respond()`

If the package doesn't exist or has an incompatible API, fall back to a plain FastAPI SSE endpoint that runs the agent directly (see Step 9 for the fallback approach).
