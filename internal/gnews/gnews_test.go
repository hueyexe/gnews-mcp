package gnews

import (
	"strings"
	"testing"
)

const sampleRSS = `<?xml version="1.0"?><rss version="2.0"><channel>
<title>"MU stock" - Google News</title>
<item><title>Micron Stock Falls - Barron's</title></item>
<item><title>Memory Demand Surges - Reuters</title></item>
</channel></rss>`

func TestRenderIncludesHeaderAndHeadlines(t *testing.T) {
	t.Parallel()

	out, err := render([]byte(sampleRSS), "MU", 10)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "### MU") {
		t.Fatalf("missing header: %q", out)
	}
	if !strings.Contains(out, "Micron Stock Falls - Barron's") {
		t.Fatalf("missing headline: %q", out)
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

func TestRenderClampsLimit(t *testing.T) {
	t.Parallel()

	out, err := render([]byte(sampleRSS), "MU", 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Micron Stock Falls") {
		t.Fatalf("expected a headline with default limit: %q", out)
	}
}
