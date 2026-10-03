package grok

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/caleb-parrot/quizez/internal/quiz"
)

const waybackTimeout = 60 * time.Second

// SaveFunc is called with a public page URL after that page was used
// in a question. Tests substitute it. The real one asks the Wayback Machine
// to capture the page and ignores the result.
type SaveFunc func(ctx context.Context, page string)

var waybackClient = &http.Client{Timeout: waybackTimeout}

// pageURL is the public Grokipedia article for a search slug.
// The slug is one path segment, kept as the site publishes it.
func pageURL(slug string) string {
	slug = strings.TrimSpace(slug)
	if slug == "" || strings.ContainsAny(slug, " \t\r\n/?#\\") {
		return ""
	}
	return "https://grokipedia.com/page/" + slug
}

// pageURLs returns the public URLs of the hits that became choices.
func pageURLs(hits []quiz.Hit, choices []string) []string {
	byTitle := map[string]quiz.Hit{}
	for _, h := range hits {
		k := quiz.Fold(h.Title)
		if k == "" {
			continue
		}
		if _, ok := byTitle[k]; !ok {
			byTitle[k] = h
		}
	}
	var urls []string
	seen := map[string]bool{}
	for _, choice := range choices {
		h, ok := byTitle[quiz.Fold(choice)]
		if !ok {
			continue
		}
		u := pageURL(h.Slug)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		urls = append(urls, u)
	}
	return urls
}

// backup asks the Wayback Machine to save each page once per process.
// It returns immediately. A failed save does not affect the question.
func (c *Client) backup(hits []quiz.Hit, choices []string) {
	if c == nil || c.save == nil {
		return
	}
	for _, page := range pageURLs(hits, choices) {
		if !c.markOnce(page) {
			continue
		}
		go func(page string) {
			ctx, cancel := context.WithTimeout(context.Background(), waybackTimeout)
			defer cancel()
			c.save(ctx, page)
		}(page)
	}
}

func (c *Client) markOnce(page string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.saved[page] {
		return false
	}
	if c.saved == nil {
		c.saved = map[string]bool{}
	}
	c.saved[page] = true
	return true
}

// waybackSave sends the public Save Page Now request for page.
// https://web.archive.org/save/<page>
func waybackSave(ctx context.Context, page string) {
	req, err := saveRequest(ctx, page)
	if err != nil {
		return
	}
	resp, err := waybackClient.Do(req)
	if err != nil {
		return
	}
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	resp.Body.Close()
}

func saveRequest(ctx context.Context, page string) (*http.Request, error) {
	path := "/save/" + page
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://web.archive.org/save/", nil)
	if err != nil {
		return nil, err
	}
	req.URL.Path = path
	req.URL.RawPath = path
	if req.URL.RequestURI() != path {
		// Parentheses and other slug characters must survive. If RawPath is
		// rejected, fall back to the encoded path Go would send.
		req.URL.RawPath = ""
		u, err := url.Parse("https://web.archive.org" + path)
		if err != nil {
			return nil, err
		}
		req.URL = u
	}
	req.Header.Set("User-Agent", "QuizEZ/0.1 (+https://github.com/Caleb-parrot/quizez)")
	return req, nil
}
