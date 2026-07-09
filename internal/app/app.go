package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"bay-area-opportunity-daily/internal/analyzer"
	"bay-area-opportunity-daily/internal/config"
	"bay-area-opportunity-daily/internal/crawler"
	"bay-area-opportunity-daily/internal/digest"
	"bay-area-opportunity-daily/internal/scorer"
	"bay-area-opportunity-daily/internal/store"
)

type App struct {
	cfg   config.Config
	store store.JSONStore
}

func New(cfg config.Config) App {
	return App{
		cfg:   cfg,
		store: store.NewJSONStore(cfg.DataDir),
	}
}

func (a App) Crawl(ctx context.Context) error {
	sources, err := crawler.LoadSources(a.cfg.SourcesPath)
	if err != nil {
		return err
	}
	docs, err := crawler.New(a.cfg).Crawl(ctx, sources)
	if err != nil {
		return err
	}
	added, err := a.store.UpsertDocuments(docs)
	if err != nil {
		return err
	}
	fmt.Printf("crawl: fetched=%d added=%d\n", len(docs), added)
	return nil
}

func (a App) Analyze(ctx context.Context) error {
	opps, err := a.store.LoadOpportunities()
	if err != nil {
		return err
	}
	an := analyzer.New(a.cfg)
	sc := scorer.New()
	now := time.Now()
	count := 0
	for i := range opps {
		if opps[i].AnalyzedAt != nil {
			continue
		}
		analysis := an.Analyze(ctx, opps[i])
		opps[i].Type = analysis.Type
		if analysis.Region != "" {
			opps[i].Region = analysis.Region
		}
		opps[i].Deadline = analysis.Deadline
		opps[i].Budget = analysis.Budget
		opps[i].Tags = analysis.Tags
		opps[i].Summary = analysis.Summary
		opps[i].SuitableFor = analysis.SuitableFor
		opps[i].NextSteps = analysis.NextSteps
		opps[i].Score = sc.Score(opps[i], analysis)
		opps[i].AnalyzedAt = &now
		count++
	}
	if err := a.store.SaveOpportunities(opps); err != nil {
		return err
	}
	fmt.Printf("analyze: analyzed=%d total=%d\n", count, len(opps))
	return nil
}

func (a App) Digest() error {
	opps, err := a.store.LoadOpportunities()
	if err != nil {
		return err
	}
	gen := digest.Generator{
		MinScore: a.cfg.MinScore,
		Limit:    a.cfg.ReportLimit,
		Days:     a.cfg.ReportDays,
	}
	report := gen.Generate(time.Now(), opps)
	if _, err := a.store.SaveReport(report); err != nil {
		return err
	}
	if err := os.MkdirAll(a.cfg.OutputDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(a.cfg.OutputDir, time.Now().Format("2006-01-02")+".md")
	if err := os.WriteFile(path, []byte(report.Content), 0o644); err != nil {
		return err
	}
	fmt.Printf("digest: items=%d output=%s\n", report.ItemCount, path)
	return nil
}

func (a App) Run(ctx context.Context) error {
	if err := a.Crawl(ctx); err != nil {
		return err
	}
	if err := a.Analyze(ctx); err != nil {
		return err
	}
	return a.Digest()
}
