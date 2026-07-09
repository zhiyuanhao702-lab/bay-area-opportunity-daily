package analyzer

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"bay-area-opportunity-daily/internal/config"
	"bay-area-opportunity-daily/internal/model"
)

type Analyzer struct {
	cfg config.Config
}

func New(cfg config.Config) Analyzer {
	return Analyzer{cfg: cfg}
}

func (a Analyzer) Analyze(ctx context.Context, op model.Opportunity) model.Analysis {
	if a.cfg.OpenAIAPIKey != "" {
		if analysis, err := a.analyzeWithOpenAI(ctx, op); err == nil && analysis.Summary != "" {
			return analysis
		}
	}
	return ruleBased(op)
}

func ruleBased(op model.Opportunity) model.Analysis {
	text := op.Title + "\n" + op.Body
	tags := detectTags(text)
	return model.Analysis{
		Type:        detectType(text, op.SourceType),
		Region:      op.Region,
		Deadline:    firstMatch(text, deadlinePatterns),
		Budget:      firstMatch(text, budgetPatterns),
		Tags:        tags,
		Summary:     firstRunes(cleanSummary(op.Body), 160),
		SuitableFor: suitableFor(tags),
		NextSteps:   nextSteps(text),
		Score:       0,
	}
}

func detectType(text, fallback string) string {
	switch {
	case contains(text, "招标", "投标", "采购公告"):
		return "招投标"
	case contains(text, "采购意向"):
		return "采购意向"
	case contains(text, "补贴", "补助", "资助", "奖励", "扶持"):
		return "政策补贴"
	case contains(text, "申报", "征集", "指南"):
		return "项目申报"
	case contains(text, "产业用房", "租赁"):
		return "产业空间"
	case fallback == "tender":
		return "招投标"
	default:
		return "政策/通知"
	}
}

func detectTags(text string) []string {
	candidates := []string{"人工智能", "软件", "数字化", "工业互联网", "信息化", "系统开发", "平台建设", "中小企业", "专精特新", "高新技术", "制造业", "设备更新", "技术改造", "项目申报", "政府采购", "招投标"}
	var tags []string
	for _, tag := range candidates {
		if strings.Contains(text, tag) {
			tags = append(tags, tag)
		}
	}
	if len(tags) == 0 {
		tags = append(tags, "科技企业")
	}
	return tags
}

func suitableFor(tags []string) string {
	joined := strings.Join(tags, "、")
	if joined == "" {
		return "广东/深圳科技型中小企业"
	}
	return joined + "相关企业或项目负责人"
}

func nextSteps(text string) string {
	switch {
	case contains(text, "申报", "征集", "指南"):
		return "查看官方指南，确认申报条件、截止时间和材料清单，尽快准备线上/线下材料。"
	case contains(text, "招标", "采购"):
		return "查看采购需求、预算金额、投标资格、评分标准和开标时间，判断是否值得参与。"
	case contains(text, "产业用房", "租赁"):
		return "核对企业资质、可申请面积、租金和材料要求，评估办公/研发空间成本。"
	default:
		return "打开官方原文，核对适用条件、截止时间和办理材料。"
	}
}

var deadlinePatterns = []*regexp.Regexp{
	regexp.MustCompile(`截止(?:时间|日期)?[：:\s]*([0-9]{4}年[0-9]{1,2}月[0-9]{1,2}日(?:\s*[0-9]{1,2}[:：][0-9]{2})?)`),
	regexp.MustCompile(`截止(?:时间|日期)?[：:\s]*([0-9]{4}[-/.][0-9]{1,2}[-/.][0-9]{1,2}(?:\s*[0-9]{1,2}[:：][0-9]{2})?)`),
	regexp.MustCompile(`([0-9]{4}年[0-9]{1,2}月[0-9]{1,2}日(?:前|止))`),
}

var budgetPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?:最高|不超过|预算(?:金额)?|资助(?:额度)?|补助(?:金额)?)[^。；;\n]{0,40}?([0-9,.]+ ?万?元)`),
	regexp.MustCompile(`([0-9,.]+ ?万?元)`),
}

func firstMatch(text string, patterns []*regexp.Regexp) string {
	for _, pattern := range patterns {
		if m := pattern.FindStringSubmatch(text); len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

func cleanSummary(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(text, " "))
	return text
}

func firstRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n]) + "..."
}

func contains(text string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}
