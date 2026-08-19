package gnews

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const sampleRSS = `<?xml version="1.0"?><rss version="2.0"><channel>
<title>"MU stock" - Google News</title>
<item><title>Micron Stock Falls - Barron's</title><link>https://news.google.com/rss/articles/abc</link><pubDate>Tue, 18 Aug 2026 20:46:00 GMT</pubDate></item>
<item><title>Memory Demand Surges - Reuters</title><link>https://news.google.com/rss/articles/def</link><pubDate>Tue, 18 Aug 2026 19:00:00 GMT</pubDate></item>
</channel></rss>`

func TestRenderIncludesHeaderHeadlineLinkAndDate(t *testing.T) {
	t.Parallel()

	out, err := render([]byte(sampleRSS), "MU", 10)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "### MU") {
		t.Fatalf("missing header: %q", out)
	}
	if !strings.Contains(out, "[Micron Stock Falls - Barron's](https://news.google.com/rss/articles/abc)") {
		t.Fatalf("missing linked headline: %q", out)
	}
	if !strings.Contains(out, "(2026-08-18)") {
		t.Fatalf("missing compact date: %q", out)
	}
}

func TestRenderRespectsLimit(t *testing.T) {
	t.Parallel()

	out, err := render([]byte(sampleRSS), "MU", 1)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "Memory Demand Surges") {
		t.Fatalf("limit not respected: %q", out)
	}
}

func TestExtractArticleTitleAndParagraphs(t *testing.T) {
	t.Parallel()

	doc, err := html.Parse(strings.NewReader(
		"<html><head><title>Micron Falls</title></head><body><p>First paragraph.</p><p>Second paragraph.</p><script>var x=1;</script></body></html>",
	))
	if err != nil {
		t.Fatal(err)
	}

	out := extractArticle(doc)

	if !strings.Contains(out, "# Micron Falls") {
		t.Fatalf("missing title: %q", out)
	}
	if !strings.Contains(out, "First paragraph.") {
		t.Fatalf("missing paragraph: %q", out)
	}
	if strings.Contains(out, "var x=1") {
		t.Fatalf("script content leaked: %q", out)
	}
}
