# Implementation Tracker

## Progress

| Step | Description | Status |
|------|-------------|--------|
| 1 | Project Scaffold | Done |
| 2 | Go MCP Server — Core (6 tools: echo, time, whoami, notes CRUD) | Done |
| 3 | Go MCP Server — Dockerfile | Done |
| 4 | Python Backend — Dependencies & Config | Done |
| 5 | Python Backend — Auth Layer | Done |
| 6 | Python Backend — MCP Client | Done |
| 7 | Python Backend — Agent | Done |
| 8 | Python Backend — ChatKit Server | Done |
| 9 | Python Backend — FastAPI App | Done |
| 10 | Frontend — ChatKit Widget | Done |
| 11 | Infrastructure — Caddy, Compose, Env | Done |
| 12 | Clone go-sdk Fork & Verify Build | Done |
| 13 | Python Backend — Verify & Adapt | Done |

## Notes

- All code steps complete
- Key adaptation in Step 13: ChatKit SDK imports as `chatkit` not `openai_chatkit`, no built-in InMemoryStore (wrote `memory_store.py`), tool context path is `ctx.context.request_context["mcp_token"]`, FastAPI wiring uses `server.process(body, context)` pattern
- Added `itsdangerous` to requirements (needed by Starlette SessionMiddleware)
- Entra app registrations (PLAN.md Phases 0-1): Done separately
- Next: fill in `.env` with real credentials, run `docker compose up --build`
