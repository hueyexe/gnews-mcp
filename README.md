# gnews-mcp

Google News headlines for AI agents. One static binary that any MCP-capable
agent can mount as a tool. No API key.

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

Build it:

```bash
go build -o gnews-mcp .
```

Then register it with your agent. `command` is an absolute path in every case,
or the binary sits on the agent's PATH.

### Claude Code

```bash
claude mcp add gnews -- /absolute/path/to/gnews-mcp
```

Or add it to `~/.claude.json` (user scope) or `.mcp.json` (project scope) under
`mcpServers`:

```json
{ "mcpServers": { "gnews": { "command": "/absolute/path/to/gnews-mcp", "args": [] } } }
```

### Codex

`~/.codex/config.toml`:

```toml
[mcp_servers.gnews]
command = "/absolute/path/to/gnews-mcp"
args = []
```

### Hermes

`~/.hermes/config.yaml`:

```yaml
mcp_servers:
  gnews:
    command: "/absolute/path/to/gnews-mcp"
    args: []
```

### OpenClaw

`~/.openclaw/openclaw.json`:

```json
{ "mcpServers": { "gnews": { "command": "/absolute/path/to/gnews-mcp", "args": [] } } }
```

Or: `openclaw mcp add gnews --command /absolute/path/to/gnews-mcp`

### OpenCode

`opencode.json`:

```json
{
  "mcp": {
    "gnews": {
      "type": "local",
      "command": ["/absolute/path/to/gnews-mcp"],
      "enabled": true
    }
  }
}
```

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