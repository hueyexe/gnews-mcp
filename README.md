# gnews-mcp

Google News headlines for AI agents. One static binary that any MCP-capable
agent (Claude, Cursor, OpenCode, Hermes) can mount as a tool. No API key.

Headlines come back as markdown, not XML or JSON, so the agent reads them
directly.

## Tools

| Tool | Returns |
|------|---------|
| `search_news` | Headlines for a free-text query |
| `ticker_news` | Headlines for a stock ticker |
| `top_stories` | Current top stories |
| `read_article` | The text of a given URL |

Each headline is one line: a clickable link and a date.

```markdown
### MU stock

- [Why Micron Stock and Its Memory Peers Are Falling - Barron's](https://news.google.com/rss/articles/CBM...) (2026-08-18)
- [AI chip stocks were riding high. Here's why Micron and others are now pulling back. - MarketWatch](https://...) (2026-08-18)
```

`read_article` is best effort. Paywalled or JavaScript-rendered pages may yield
little or no text.

## Install

Build the binary:

```bash
go build -o gnews-mcp .
```

Then register it as a stdio MCP server. Claude Desktop and Cursor use this
`mcpServers` shape:

```json
{
  "mcpServers": {
    "gnews": {
      "command": "/absolute/path/to/gnews-mcp",
      "args": []
    }
  }
}
```

Config file locations:

- Claude Desktop: `~/Library/Application Support/Claude/claude_desktop_config.json`
  (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows)
- Cursor: `~/.cursor/mcp.json` (global) or `.cursor/mcp.json` (per project)
- OpenCode: the `mcp` key in `opencode.json`
- Hermes and other clients: the same `mcpServers` block in their config

`command` must be an absolute path, or the binary must sit on the agent's PATH.

## Build and test

```bash
make build    # go build -o bin/gnews-mcp .
make test     # go test -race ./...
make lint     # golangci-lint run ./...
```

## File map

- `main.go` — MCP server: tool registration + JSON-RPC over stdio
- `internal/gnews/` — RSS fetch, markdown rendering, article extraction

## Does not own

- Headlines only; no sentiment, no full-article crawling beyond `read_article`
- No history or state between calls
- No Google authentication; it reads the public RSS feed (personal,
  non-commercial use, see Google News RSS terms)