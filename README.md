# gnews-mcp

Google News headlines, delivered to AI agents as compact markdown over the
[Model Context Protocol](https://modelcontextprotocol.io).

No API key. A single static binary that any MCP-capable agent (Claude, Cursor,
OpenCode, Hermes, …) can mount as a tool.

## What

Three tools:

| Tool | Returns |
|------|---------|
| `search_news` | Recent Google News headlines for a free-text query |
| `ticker_news` | Recent headlines for a stock ticker (auto-appends "stock") |
| `top_stories` | Current top stories |

Output is intentionally minimal — a short header plus one headline per line —
so agents pay the fewest tokens for the signal:

```markdown
### MU stock

- Why Micron Stock and Its Memory Peers Are Falling - Barron's
- AI chip stocks were riding high. Here's why Micron and others are now pulling back. - MarketWatch
```

## Why

LLM agents want *just the headlines* to reason over — not HTML, not XML, not
JSON stuffed with metadata they don't use. Most news integrations require a paid
API key or dump verbose structures. This is the minimal middle: a public RSS
feed, parsed and re-rendered as token-efficient markdown.

## How to run

```bash
go build -o gnews-mcp .
```

Then register it as a stdio MCP server in your agent's config — e.g. Claude
Desktop `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "gnews": { "command": "/path/to/gnews-mcp" }
  }
}
```

## File map

- `main.go` — MCP server: tool registration + JSON-RPC over stdio.
- `internal/gnews/` — RSS fetch + markdown rendering (stdlib `net/http` + `encoding/xml`).

## Does not own

- It does not fetch full article bodies or do sentiment analysis — headlines only.
- It keeps no history or state between calls.
- It does not authenticate with Google; it reads the public RSS feed (personal,
  non-commercial use — see the Google News RSS terms).