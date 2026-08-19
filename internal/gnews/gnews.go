// Package gnews fetches Google News RSS and renders compact markdown for LLMs.
//
// No API key: it reads Google News' public RSS feeds. Output is intentionally
// minimal — a short header plus one headline per line — so agents pay as few
// tokens as possible for the signal.
package gnews

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	searchURL = "https://news.google.com/rss/search"
	topURL    = "https://news.google.com/rss"
	userAgent = "gnews-mcp/1.0 (+https://github.com/hueyexe/gnews-mcp)"
)

var errBadStatus = errors.New("unexpected HTTP status")

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title string `xml:"title"`
}

// Search returns up to limit Google News headlines for query as compact markdown.
func Search(ctx context.Context, query string, limit int) (string, error) {
	q := url.Values{}
	q.Set("q", query)
	q.Set("hl", "en-US")
	q.Set("gl", "US")
	q.Set("ceid", "US:en")
	return fetch(ctx, searchURL+"?"+q.Encode(), query, limit)
}

// TickerNews returns up to limit headlines for a stock ticker.
func TickerNews(ctx context.Context, ticker string, limit int) (string, error) {
	return Search(ctx, ticker+" stock", limit)
}

// TopStories returns up to limit current top stories as compact markdown.
func TopStories(ctx context.Context, limit int) (string, error) {
	q := url.Values{}
	q.Set("hl", "en-US")
	q.Set("gl", "US")
	q.Set("ceid", "US:en")
	return fetch(ctx, topURL+"?"+q.Encode(), "top stories", limit)
}

func fetch(ctx context.Context, rawURL, label string, limit int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s: %w", label, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d for %s", errBadStatus, resp.StatusCode, label)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}

	return render(body, label, limit)
}

// render parses an RSS payload and formats its headlines as compact markdown.
func render(body []byte, label string, limit int) (string, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return "", fmt.Errorf("parse rss: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "### %s\n\n", label)
	count := 0
	for _, it := range feed.Channel.Items {
		title := strings.TrimSpace(it.Title)
		if title == "" {
			continue
		}
		if count >= limit {
			break
		}
		fmt.Fprintf(&b, "- %s\n", title)
		count++
	}
	return strings.TrimRight(b.String(), "\n"), nil
}
