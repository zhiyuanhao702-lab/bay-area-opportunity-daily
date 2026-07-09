package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bay-area-opportunity-daily/internal/model"
)

type JSONStore struct {
	dir string
}

func NewJSONStore(dir string) JSONStore {
	return JSONStore{dir: dir}
}

func (s JSONStore) Ensure() error {
	return os.MkdirAll(s.dir, 0o755)
}

func (s JSONStore) UpsertDocuments(docs []model.Document) (int, error) {
	opps, err := s.LoadOpportunities()
	if err != nil {
		return 0, err
	}
	byURL := map[string]int{}
	byHash := map[string]int{}
	for i, op := range opps {
		byURL[op.URL] = i
		if op.ContentHash != "" {
			byHash[op.ContentHash] = i
		}
	}
	now := time.Now()
	added := 0
	for _, doc := range docs {
		if _, ok := byURL[doc.URL]; ok {
			continue
		}
		if doc.ContentHash != "" {
			if _, ok := byHash[doc.ContentHash]; ok {
				continue
			}
		}
		opps = append(opps, model.Opportunity{
			ID:          shortID(doc.ContentHash),
			Title:       doc.Title,
			URL:         doc.URL,
			Source:      doc.Source,
			SourceType:  doc.SourceType,
			Region:      doc.Region,
			PublishedAt: doc.PublishedAt,
			Body:        doc.Body,
			ContentHash: doc.ContentHash,
			CreatedAt:   now,
		})
		added++
	}
	if err := s.SaveOpportunities(opps); err != nil {
		return 0, err
	}
	return added, nil
}

func (s JSONStore) LoadOpportunities() ([]model.Opportunity, error) {
	path := filepath.Join(s.dir, "opportunities.json")
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var opps []model.Opportunity
	if err := json.Unmarshal(b, &opps); err != nil {
		return nil, err
	}
	return opps, nil
}

func (s JSONStore) SaveOpportunities(opps []model.Opportunity) error {
	if err := s.Ensure(); err != nil {
		return err
	}
	sort.SliceStable(opps, func(i, j int) bool {
		return opps[i].CreatedAt.After(opps[j].CreatedAt)
	})
	b, err := json.MarshalIndent(opps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "opportunities.json"), b, 0o644)
}

func (s JSONStore) SaveReport(report model.Report) (string, error) {
	if err := os.MkdirAll(filepath.Join(s.dir, "reports"), 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(s.dir, "reports", report.Date.Format("2006-01-02")+".json")
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, b, 0o644)
}

func shortID(hash string) string {
	hash = strings.TrimSpace(hash)
	if len(hash) >= 12 {
		return hash[:12]
	}
	return hash
}
