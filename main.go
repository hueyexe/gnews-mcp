// Command gnews-mcp exposes Google News headlines to agents as an MCP server
// over stdio. Tools: search_news, ticker_news, top_stories.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/hueyexe/gnews-mcp/internal/gnews"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gnews-mcp:", err)
		os.Exit(1)
	}
}

func run() error {
	s := server.NewMCPServer("gnews-mcp", "1.0.0", server.WithToolCapabilities(true))

	s.AddTool(
		mcp.NewTool(
			"search_news",
			mcp.WithDescription("Search Google News and return recent headlines as compact markdown."),
			mcp.WithString("query", mcp.Required(), mcp.Description("Search query, e.g. \"Nvidia earnings\".")),
			mcp.WithNumber("limit", mcp.Description("Max headlines (default 10, max 50).")),
		),
		handleSearchNews,
	)
	s.AddTool(
		mcp.NewTool(
			"ticker_news",
			mcp.WithDescription("Return recent Google News headlines for a stock ticker as compact markdown."),
			mcp.WithString("ticker", mcp.Required(), mcp.Description("Stock ticker, e.g. \"MU\".")),
			mcp.WithNumber("limit", mcp.Description("Max headlines (default 10, max 50).")),
		),
		handleTickerNews,
	)
	s.AddTool(
		mcp.NewTool(
			"top_stories",
			mcp.WithDescription("Return current top stories from Google News as compact markdown."),
			mcp.WithNumber("limit", mcp.Description("Max headlines (default 10, max 50).")),
		),
		handleTopStories,
	)

	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("serve stdio: %w", err)
	}
	return nil
}

//nolint:gocritic // request-by-value signature is mandated by the mcp-go SDK handler type
func handleSearchNews(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := req.GetString("query", "")
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}
	out, err := gnews.Search(ctx, query, intArg(req.GetArguments(), 10))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(out), nil
}

//nolint:gocritic // request-by-value signature is mandated by the mcp-go SDK handler type
func handleTickerNews(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ticker := req.GetString("ticker", "")
	if ticker == "" {
		return mcp.NewToolResultError("ticker is required"), nil
	}
	out, err := gnews.TickerNews(ctx, ticker, intArg(req.GetArguments(), 10))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(out), nil
}

//nolint:gocritic // request-by-value signature is mandated by the mcp-go SDK handler type
func handleTopStories(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	out, err := gnews.TopStories(ctx, intArg(req.GetArguments(), 10))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return textResult(out), nil
}

func intArg(args map[string]any, def int) int {
	if args == nil {
		return def
	}
	if v, ok := args["limit"].(float64); ok {
		return int(v)
	}
	return def
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: text},
		},
	}
}
