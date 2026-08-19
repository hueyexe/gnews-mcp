// Package gnews fetches Google News RSS and renders compact markdown for LLMs.
//
// No API key: it reads Google News' public RSS feeds. Output is intentionally
// minimal — a short header plus one clickable headline per line — so agents pay
// few tokens for the signal while keeping a link to follow.
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

	"golang.org/x/net/html"
)

const (
	searchURL = "https://news.google.com/rss/search"
	topURL    = "https://news.google.com/rss"
	userAgent = "gnews-mcp/1.1 (+https://github.com/hueyexe/gnews-mcp)"
)

var errBadStatus = errors.New("unexpected HTTP status")

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
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

// ReadArticle fetches a URL and extracts its title and paragraphs as markdown.
// Best effort: paywalled or JavaScript-rendered pages may return little or no text.
func ReadArticle(ctx context.Context, rawURL string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch article: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: %d", errBadStatus, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parse html: %w", err)
	}

	return extractArticle(doc), nil
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
		writeHeadline(&b, title, strings.TrimSpace(it.Link), compactDate(it.PubDate))
		count++
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func writeHeadline(b *strings.Builder, title, link, date string) {
	if link == "" {
		fmt.Fprintf(b, "- %s", title)
	} else {
		fmt.Fprintf(b, "- [%s](%s)", title, link)
	}
	if date != "" {
		fmt.Fprintf(b, " (%s)", date)
	}
	fmt.Fprintln(b)
}

func compactDate(s string) string {
	t, err := time.Parse(time.RFC1123, strings.TrimSpace(s))
	if err != nil {
		return ""
	}
	return t.Format("2006-01-02")
}

// extractArticle walks an HTML tree, keeping the title and paragraph text.
func extractArticle(doc *html.Node) string {
	var title string
	var paragraphs []string
	collect(doc, &title, &paragraphs)

	var b strings.Builder
	if title != "" {
		fmt.Fprintf(&b, "# %s\n\n", title)
	}
	for _, p := range paragraphs {
		fmt.Fprintf(&b, "%s\n\n", p)
	}
	out := strings.TrimRight(b.String(), "\n")
	if out == "" {
		return "(no readable article text)"
	}
	return out
}

func collect(n *html.Node, title *string, paragraphs *[]string) {
	if n.Type == html.ElementNode {
		switch n.Data {
		case "title":
			if *title == "" {
				*title = textContent(n)
			}
			return
		case "p":
			if t := strings.TrimSpace(textContent(n)); t != "" {
				*paragraphs = append(*paragraphs, t)
			}
			return
		case "script", "style", "noscript", "nav", "footer", "aside":
			return
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collect(c, title, paragraphs)
	}
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			b.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(b.String())
}
