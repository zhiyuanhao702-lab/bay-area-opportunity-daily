package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	anchorRe = regexp.MustCompile(`(?is)<a\s+[^>]*href=["']([^"']+)["'][^>]*>(.*?)</a>`)
	scriptRe = regexp.MustCompile(`(?is)<script.*?</script>`)
	styleRe  = regexp.MustCompile(`(?is)<style.*?</style>`)
	tagRe    = regexp.MustCompile(`(?is)<[^>]+>`)
	spaceRe  = regexp.MustCompile(`\s+`)
	dateRe   = regexp.MustCompile(`(20\d{2})[-年./](\d{1,2})[-月./](\d{1,2})`)
	noiseRe  = regexp.MustCompile(`(?i)javascript:|mailto:|#`)
)

type link struct {
	Title string
	URL   string
}

func extractLinks(base, body string) []link {
	matches := anchorRe.FindAllStringSubmatch(body, -1)
	links := make([]link, 0, len(matches))
	for _, m := range matches {
		title := cleanText(m[2])
		if utf8.RuneCountInString(title) < 4 {
			continue
		}
		href := strings.TrimSpace(html.UnescapeString(m[1]))
		if href == "" || noiseRe.MatchString(href) {
			continue
		}
		u, err := resolveURL(base, href)
		if err != nil {
			continue
		}
		links = append(links, link{Title: title, URL: u})
	}
	return links
}

func cleanText(s string) string {
	s = scriptRe.ReplaceAllString(s, " ")
	s = styleRe.ReplaceAllString(s, " ")
	s = tagRe.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "　", " ")
	return strings.TrimSpace(spaceRe.ReplaceAllString(s, " "))
}

func stripHTML(body string) string {
	text := cleanText(body)
	runes := []rune(text)
	if len(runes) > 8000 {
		return string(runes[:8000])
	}
	return text
}

func resolveURL(base, href string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	return baseURL.ResolveReference(ref).String(), nil
}

func firstDate(text string) *time.Time {
	m := dateRe.FindStringSubmatch(text)
	if len(m) != 4 {
		return nil
	}
	value := m[1] + "-" + two(m[2]) + "-" + two(m[3])
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &t
}

func two(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

func contentHash(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

func containsAny(text string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
