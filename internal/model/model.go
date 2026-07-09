package model

import "time"

type SourceConfig struct {
	Name            string   `json:"name"`
	BaseURL         string   `json:"base_url"`
	ListURL         string   `json:"list_url"`
	Region          string   `json:"region"`
	SourceType      string   `json:"source_type"`
	IncludeKeywords []string `json:"include_keywords"`
}

type RawItem struct {
	Title      string
	URL        string
	SourceName string
	Region     string
	SourceType string
}

type Document struct {
	Title       string
	URL         string
	Source      string
	SourceType  string
	Region      string
	PublishedAt *time.Time
	Body        string
	ContentHash string
}

type Analysis struct {
	Type        string   `json:"type"`
	Region      string   `json:"region"`
	Deadline    string   `json:"deadline"`
	Budget      string   `json:"budget"`
	Tags        []string `json:"tags"`
	Summary     string   `json:"summary"`
	SuitableFor string   `json:"suitable_for"`
	NextSteps   string   `json:"next_steps"`
	Score       int      `json:"score"`
}

type Opportunity struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	Source      string     `json:"source"`
	SourceType  string     `json:"source_type"`
	Region      string     `json:"region"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	Body        string     `json:"body"`
	ContentHash string     `json:"content_hash"`

	Type        string   `json:"type"`
	Deadline    string   `json:"deadline"`
	Budget      string   `json:"budget"`
	Tags        []string `json:"tags"`
	Summary     string   `json:"summary"`
	SuitableFor string   `json:"suitable_for"`
	NextSteps   string   `json:"next_steps"`
	Score       int      `json:"score"`

	CreatedAt  time.Time  `json:"created_at"`
	AnalyzedAt *time.Time `json:"analyzed_at,omitempty"`
}

type Report struct {
	Date      time.Time `json:"date"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ItemCount int       `json:"item_count"`
	CreatedAt time.Time `json:"created_at"`
}
