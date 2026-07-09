package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"bay-area-opportunity-daily/internal/config"
	"bay-area-opportunity-daily/internal/model"
)

type Crawler struct {
	cfg    config.Config
	client *http.Client
}

func New(cfg config.Config) *Crawler {
	return &Crawler{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.HTTPTimeout,
		},
	}
}

func LoadSources(path string) ([]model.SourceConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var sources []model.SourceConfig
	if err := json.Unmarshal(b, &sources); err != nil {
		return nil, err
	}
	return sources, nil
}

func (c *Crawler) Crawl(ctx context.Context, sources []model.SourceConfig) ([]model.Document, error) {
	var docs []model.Document
	for _, source := range sources {
		items, err := c.list(ctx, source)
		if err != nil {
			return docs, fmt.Errorf("%s list: %w", source.Name, err)
		}
		for i, item := range items {
			if c.cfg.MaxItemsPerSource > 0 && i >= c.cfg.MaxItemsPerSource {
				break
			}
			doc, err := c.fetchDetail(ctx, item)
			if err != nil {
				continue
			}
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

func (c *Crawler) list(ctx context.Context, source model.SourceConfig) ([]model.RawItem, error) {
	body, err := c.fetch(ctx, source.ListURL)
	if err != nil {
		return nil, err
	}
	links := extractLinks(source.ListURL, body)
	items := make([]model.RawItem, 0, len(links))
	seen := map[string]bool{}
	for _, l := range links {
		if seen[l.URL] {
			continue
		}
		seen[l.URL] = true
		if !containsAny(l.Title, source.IncludeKeywords) {
			continue
		}
		items = append(items, model.RawItem{
			Title:      l.Title,
			URL:        l.URL,
			SourceName: source.Name,
			Region:     source.Region,
			SourceType: source.SourceType,
		})
	}
	return items, nil
}

func (c *Crawler) fetchDetail(ctx context.Context, item model.RawItem) (model.Document, error) {
	body, err := c.fetch(ctx, item.URL)
	if err != nil {
		return model.Document{}, err
	}
	text := stripHTML(body)
	publishedAt := firstDate(text)
	return model.Document{
		Title:       item.Title,
		URL:         item.URL,
		Source:      item.SourceName,
		SourceType:  item.SourceType,
		Region:      item.Region,
		PublishedAt: publishedAt,
		Body:        text,
		ContentHash: contentHash(item.Title, item.URL, text),
	}, nil
}

func (c *Crawler) fetch(ctx context.Context, target string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "bay-area-opportunity-daily/0.1 (+https://afdian.com/a/gdpolicy)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.6")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("http status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 3<<20))
	if err != nil {
		return "", err
	}
	time.Sleep(300 * time.Millisecond)
	return string(b), nil
}
