# AGENTS.md

## Project

`gnews-mcp` is a single-binary MCP server that returns Google News headlines as
compact markdown for LLM agents. No API key; RSS fetching and parsing use the
Go stdlib, and article text extraction uses `golang.org/x/net/html`.

## Architecture

- `main.go` registers four MCP tools (`search_news`, `ticker_news`,
  `top_stories`, `read_article`) and serves them over stdio via
  `github.com/mark3labs/mcp-go`.
- `internal/gnews` fetches Google News RSS (`net/http`) and renders markdown
  (`encoding/xml`). `render` is the testable core; `extractArticle` is the
  best-effort article reader.

## Commands

```bash
make build    # go build -o bin/gnews-mcp .
make test     # go test -race ./...
make lint     # golangci-lint run ./...
make fmt      # golangci-lint fmt ./...
```

## Standards

- Go 1.26, strict `golangci-lint` v2 (`.golangci.yml`), zero tolerance.
- No `panic`/`log.Fatal`/`os.Exit` outside `main()`; return errors to `main()`.
- Sentinel errors + `%w` wrapping; `context` threaded through all fetches.
- The mcp-go handler signature `(ctx, mcp.CallToolRequest)` is SDK-mandated and
  passes the request by value; the `//nolint:gocritic` on handlers is justified.