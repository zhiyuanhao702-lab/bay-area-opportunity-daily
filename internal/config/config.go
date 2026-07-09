package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	SourcesPath       string
	DataDir           string
	OutputDir         string
	HTTPTimeout       time.Duration
	MaxItemsPerSource int
	ReportDays        int
	MinScore          int
	ReportLimit       int

	OpenAIAPIKey  string
	OpenAIBaseURL string
	OpenAIModel   string
}

func Default() Config {
	return Config{
		SourcesPath:       getenv("BAY_DAILY_SOURCES", "configs/sources.json"),
		DataDir:           getenv("BAY_DAILY_DATA_DIR", "data"),
		OutputDir:         getenv("BAY_DAILY_OUTPUT_DIR", "output"),
		HTTPTimeout:       15 * time.Second,
		MaxItemsPerSource: envInt("BAY_DAILY_MAX_ITEMS", 20),
		ReportDays:        envInt("BAY_DAILY_REPORT_DAYS", 30),
		MinScore:          envInt("BAY_DAILY_MIN_SCORE", 50),
		ReportLimit:       envInt("BAY_DAILY_REPORT_LIMIT", 10),
		OpenAIAPIKey:      os.Getenv("OPENAI_API_KEY"),
		OpenAIBaseURL:     getenv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
		OpenAIModel:       getenv("OPENAI_MODEL", "gpt-4.1-mini"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
