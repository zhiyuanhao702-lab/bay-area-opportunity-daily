package crawler

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"bay-area-opportunity-daily/internal/config"
	"bay-area-opportunity-daily/internal/model"
)

func TestCrawlContinuesAfterSourceFailure(t *testing.T) {
	c := New(config.Config{
		HTTPTimeout:       time.Second,
		MaxItemsPerSource: 10,
	})
	c.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/forbidden":
			return response(r, http.StatusForbidden, ""), nil
		case "/list":
			return response(r, http.StatusOK, `<a href="/detail">人工智能项目申报</a>`), nil
		case "/detail":
			return response(r, http.StatusOK, `<html><body>人工智能项目申报，截止时间 2026年7月31日。</body></html>`), nil
		default:
			return response(r, http.StatusNotFound, ""), nil
		}
	})
	sources := []model.SourceConfig{
		{Name: "blocked", ListURL: "https://example.test/forbidden"},
		{
			Name:            "working",
			ListURL:         "https://example.test/list",
			Region:          "深圳",
			SourceType:      "policy",
			IncludeKeywords: []string{"人工智能"},
		},
	}

	docs, err := c.Crawl(context.Background(), sources)
	if err != nil {
		t.Fatalf("Crawl() error = %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("Crawl() documents = %d, want 1", len(docs))
	}
	if docs[0].Title != "人工智能项目申报" {
		t.Fatalf("Crawl() title = %q", docs[0].Title)
	}
}

func TestCrawlFailsWhenEverySourceFails(t *testing.T) {
	c := New(config.Config{HTTPTimeout: time.Second})
	c.client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return response(r, http.StatusForbidden, ""), nil
	})
	_, err := c.Crawl(context.Background(), []model.SourceConfig{
		{Name: "blocked", ListURL: "https://example.test/forbidden"},
	})
	if err == nil {
		t.Fatal("Crawl() error = nil, want all-sources-failed error")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func response(req *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}
