# Step 13: Python Backend — Verify Setup & Adapt

## Actions

### 1. Create virtual environment and install deps

```bash
cd /home/matt/Projects/MattChat/chatkit-backend
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

### 2. Verify package availability

```bash
python3 -c "import msal; print('msal OK')"
python3 -c "import mcp; print('mcp OK')"
python3 -c "import agents; print('agents OK')"
python3 -c "import fastapi; print('fastapi OK')"
```

### 3. Verify ChatKit package

```bash
python3 -c "import openai_chatkit; print('chatkit OK')"
```

**If this fails** (package doesn't exist or has a different name):
- Search PyPI: `pip search openai-chatkit` or check https://pypi.org/project/openai-chatkit/
- Try alternative names: `chatkit`, `openai-chatkit-python`
- If unavailable, implement the **fallback approach**: replace `chatkit_server.py` with a direct FastAPI SSE streaming endpoint using the Agents SDK's `Runner.run_streamed()` — no ChatKit SDK dependency needed

### 4. Verify Agents SDK tool context API

```bash
python3 -c "from agents import Agent, function_tool, RunContextWrapper; print('agents API OK')"
```

If `RunContextWrapper` doesn't exist, check the actual context-passing API in the `agents` package and adjust `agent.py` accordingly.

### 5. Verify MCP client API

```bash
python3 -c "from mcp.client.streamable_http import streamablehttp_client; print('mcp client OK')"
```

If the import path differs, check the `mcp` package structure and adjust `mcp_client.py`.

### 6. Adapt code to actual APIs

After verifying all imports, fix any mismatches between the plan's assumed APIs and the actual package interfaces. Common adjustments:
- Import paths
- Method signatures
- Context passing mechanisms
- Streaming response formats

### 7. Smoke test (without real credentials)

```bash
# Will fail on missing env vars, but verifies no import/syntax errors
python3 -c "from app import app; print('app loads OK')" 2>&1 | head -5
```
