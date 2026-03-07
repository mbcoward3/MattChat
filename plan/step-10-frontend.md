# Step 10: Frontend — ChatKit Widget

## File: `chatkit-backend/static/index.html`

Minimal HTML page that loads the ChatKit JS widget and points it at the Python backend.

### Implementation

```html
<!DOCTYPE html>
<html>
<head>
  <title>MattChat — Enterprise MCP Demo</title>
  <script src="https://cdn.platform.openai.com/deployments/chatkit/chatkit.js" async></script>
  <style>
    body { margin: 0; font-family: sans-serif; }
    #chat { height: 100vh; }
  </style>
</head>
<body>
  <div id="chat"></div>
  <script>
    const el = document.getElementById('chat');
    const widget = document.createElement('chatkit-widget');

    // Self-hosted mode: talks to our Python backend, not OpenAI directly
    widget.setOptions({
      api: {
        url: '/api/chat',
      }
    });

    el.appendChild(widget);
  </script>
</body>
</html>
```

### Notes

- **No OpenAI client secret** in the frontend — the Python backend handles all LLM inference
- **No direct browser-to-OpenAI calls** — all traffic goes through the backend
- The `setOptions` shape for self-hosted mode may differ from what's shown — verify against ChatKit JS docs
- The CDN URL for ChatKit JS should be confirmed (it may be versioned or have a different path)
