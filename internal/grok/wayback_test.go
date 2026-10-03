package grok

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/benoute/grokipedia-mcp/pkg/grokipedia"
	"github.com/caleb-parrot/quizez/internal/quiz"
)

func TestPageURL(t *testing.T) {
	if got := pageURL("Black_hole"); got != "https://grokipedia.com/page/Black_hole" {
		t.Fatal(got)
	}
	if got := pageURL("Slug_(publishing)"); got != "https://grokipedia.com/page/Slug_(publishing)" {
		t.Fatal(got)
	}
	for _, slug := range []string{"", "  ", "a/b", "a?b", "a b", "a#b"} {
		if got := pageURL(slug); got != "" {
			t.Fatalf("slug %q -> %q", slug, got)
		}
	}
}

func TestSaveRequestTargetsWayback(t *testing.T) {
	page := "https://grokipedia.com/page/Slug_(publishing)"
	req, err := saveRequest(context.Background(), page)
	if err != nil {
		t.Fatal(err)
	}
	if req.Method != "GET" {
		t.Fatal(req.Method)
	}
	if req.URL.Host != "web.archive.org" {
		t.Fatal(req.URL.Host)
	}
	if req.URL.RequestURI() != "/save/"+page {
		t.Fatal(req.URL.RequestURI())
	}
	if !strings.Contains(req.Header.Get("User-Agent"), "QuizEZ") {
		t.Fatal(req.Header.Get("User-Agent"))
	}
}

func TestDrawSavesUsedPagesOnce(t *testing.T) {
	var mu sync.Mutex
	var got []string
	ready := make(chan struct{}, 8)
	c := &Client{
		search: func(context.Context, string, int, int) ([]grokipedia.SearchResult, error) {
			return batch(), nil
		},
		save: func(_ context.Context, page string) {
			mu.Lock()
			got = append(got, page)
			mu.Unlock()
			ready <- struct{}{}
		},
	}
	cat := quiz.Category{Name: "Space", Queries: []string{"astronomy"}}
	if _, err := c.Draw(context.Background(), cat, nil); err != nil {
		t.Fatal(err)
	}
	waitSaves(t, ready, 4)
	if _, err := c.Draw(context.Background(), cat, nil); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ready:
		t.Fatal("saved a page twice")
	case <-time.After(50 * time.Millisecond):
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 4 {
		t.Fatalf("saves %v", got)
	}
	want := map[string]bool{
		"https://grokipedia.com/page/Europa":   true,
		"https://grokipedia.com/page/Titan":    true,
		"https://grokipedia.com/page/Ganymede": true,
		"https://grokipedia.com/page/Callisto": true,
	}
	for _, page := range got {
		if !want[page] {
			t.Fatalf("unexpected %s in %v", page, got)
		}
	}
}

func TestDrawDoesNotSaveWhenSearchFails(t *testing.T) {
	saved := false
	c := &Client{
		search: func(context.Context, string, int, int) ([]grokipedia.SearchResult, error) {
			return nil, errNoMatch
		},
		save: func(context.Context, string) { saved = true },
	}
	_, err := c.Draw(context.Background(), quiz.Category{Name: "Space", Queries: []string{"astronomy"}}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if saved {
		t.Fatal("wayback was called")
	}
}

func TestDrawSkipsEmptySlug(t *testing.T) {
	var mu sync.Mutex
	var got []string
	ready := make(chan struct{}, 8)
	c := &Client{
		search: func(context.Context, string, int, int) ([]grokipedia.SearchResult, error) {
			hits := batch()
			hits[0].Slug = ""
			return hits, nil
		},
		save: func(_ context.Context, page string) {
			mu.Lock()
			got = append(got, page)
			mu.Unlock()
			ready <- struct{}{}
		},
	}
	if _, err := c.Draw(context.Background(), quiz.Category{Name: "Space", Queries: []string{"astronomy"}}, nil); err != nil {
		t.Fatal(err)
	}
	waitSaves(t, ready, 3)
	mu.Lock()
	defer mu.Unlock()
	for _, page := range got {
		if strings.HasSuffix(page, "/Europa") || page == "" {
			t.Fatalf("saved %q", page)
		}
	}
}

func waitSaves(t *testing.T, ready <-chan struct{}, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		select {
		case <-ready:
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out waiting for save %d", i+1)
		}
	}
}
