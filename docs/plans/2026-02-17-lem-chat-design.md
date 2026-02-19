# LEM Chat — Web Components Design

**Date**: 2026-02-17
**Status**: Approved

## Summary

Standalone chat UI built with vanilla Web Components (Custom Elements + Shadow DOM). Connects to the MLX inference server's OpenAI-compatible SSE streaming endpoint. Zero framework dependencies. Single JS file output, embeddable anywhere.

## Components

| Element | Purpose |
|---------|---------|
| `<lem-chat>` | Container. Conversation state, SSE connection, config via attributes |
| `<lem-messages>` | Scrollable message list with auto-scroll anchoring |
| `<lem-message>` | Single message bubble. Streams tokens for assistant messages |
| `<lem-input>` | Text input, Enter to send, Shift+Enter for newline |

## Data Flow

```
User types in <lem-input>
  → dispatches 'lem-send' CustomEvent
  → <lem-chat> catches it
  → adds user message to <lem-messages>
  → POST /v1/chat/completions {stream: true, messages: [...history]}
  → reads SSE chunks via fetch + ReadableStream
  → appends tokens to streaming <lem-message>
  → on [DONE], finalises message
```

## Configuration

```html
<lem-chat endpoint="http://localhost:8090" model="qwen3-8b"></lem-chat>
```

Attributes: `endpoint`, `model`, `system-prompt`, `max-tokens`, `temperature`

## Theming

Shadow DOM with CSS custom properties:

```css
--lem-bg: #1a1a1e;
--lem-msg-user: #2a2a3e;
--lem-msg-assistant: #1e1e2a;
--lem-accent: #5865f2;
--lem-text: #e0e0e0;
--lem-font: system-ui;
```

## Markdown

Minimal inline parsing: fenced code blocks, inline code, bold, italic. No library.

## File Structure

```
lem-chat/
├── index.html          # Demo page
├── src/
│   ├── lem-chat.ts     # Main container + SSE client
│   ├── lem-messages.ts # Message list with scroll anchoring
│   ├── lem-message.ts  # Single message with streaming
│   ├── lem-input.ts    # Text input
│   ├── markdown.ts     # Minimal markdown → HTML
│   └── styles.ts       # CSS template literals
├── package.json        # typescript + esbuild
└── tsconfig.json
```

Build: `esbuild src/lem-chat.ts --bundle --outfile=dist/lem-chat.js`

## Not in v1

- Model selection UI
- Conversation persistence
- File/image upload
- Syntax highlighting
- Typing indicators
- User avatars
