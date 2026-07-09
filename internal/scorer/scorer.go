package scorer

import (
	"strings"
	"time"

	"bay-area-opportunity-daily/internal/model"
)

type Scorer struct{}

func New() Scorer {
	return Scorer{}
}

func (s Scorer) Score(op model.Opportunity, analysis model.Analysis) int {
	score := analysis.Score
	text := strings.ToLower(op.Title + " " + op.Body + " " + analysis.Summary + " " + strings.Join(analysis.Tags, " "))

	add := func(points int, terms ...string) {
		for _, term := range terms {
			if strings.Contains(text, strings.ToLower(term)) {
				score += points
				return
			}
		}
	}

	add(20, "深圳", "广东")
	add(25, "人工智能", "AI", "软件", "数字化", "工业互联网", "信息化", "系统开发", "平台建设")
	add(20, "申报", "补贴", "资助", "奖励", "招标", "采购", "征集")
	add(10, "万元", "预算", "金额", "最高")

	if analysis.Deadline != "" {
		score += 10
	}
	if op.PublishedAt != nil && time.Since(*op.PublishedAt) <= 7*24*time.Hour {
		score += 10
	}
	if strings.Contains(text, "中标结果") || strings.Contains(text, "废标") || strings.Contains(text, "终止公告") {
		score -= 35
	}
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
