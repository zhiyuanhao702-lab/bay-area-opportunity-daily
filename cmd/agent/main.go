package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"bay-area-opportunity-daily/internal/app"
	"bay-area-opportunity-daily/internal/config"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		usage()
		return nil
	}
	cmd := os.Args[1]
	cfg, err := parseFlags(os.Args[2:])
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	application := app.New(cfg)

	switch cmd {
	case "crawl":
		return application.Crawl(ctx)
	case "analyze":
		return application.Analyze(ctx)
	case "digest":
		return application.Digest()
	case "run":
		return application.Run(ctx)
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func parseFlags(args []string) (config.Config, error) {
	cfg := config.Default()
	fs := flag.NewFlagSet("policy-agent", flag.ContinueOnError)
	fs.StringVar(&cfg.SourcesPath, "sources", cfg.SourcesPath, "path to sources JSON")
	fs.StringVar(&cfg.DataDir, "data", cfg.DataDir, "data directory")
	fs.StringVar(&cfg.OutputDir, "output", cfg.OutputDir, "output directory")
	fs.IntVar(&cfg.MaxItemsPerSource, "limit-per-source", cfg.MaxItemsPerSource, "max list items per source")
	fs.IntVar(&cfg.ReportDays, "days", cfg.ReportDays, "include opportunities published within N days")
	fs.IntVar(&cfg.MinScore, "min-score", cfg.MinScore, "minimum score for report")
	fs.IntVar(&cfg.ReportLimit, "report-limit", cfg.ReportLimit, "max opportunities in digest")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func usage() {
	fmt.Print(`湾区机会日报 CLI

Usage:
  go run ./cmd/agent crawl   [flags]   # 抓取公开信息源并入库
  go run ./cmd/agent analyze [flags]   # AI/规则分析未处理机会
  go run ./cmd/agent digest  [flags]   # 生成 Markdown 日报
  go run ./cmd/agent run     [flags]   # 依次执行 crawl + analyze + digest

Flags:
  --sources configs/sources.json
  --data data
  --output output
  --limit-per-source 20
  --days 30
  --min-score 50
  --report-limit 10

Optional env:
  OPENAI_API_KEY    不设置则使用本地规则分析
  OPENAI_BASE_URL   默认 https://api.openai.com/v1
  OPENAI_MODEL      默认 gpt-4.1-mini
`)
}
