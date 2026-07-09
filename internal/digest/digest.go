package digest

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"bay-area-opportunity-daily/internal/model"
)

type Generator struct {
	MinScore int
	Limit    int
	Days     int
}

func (g Generator) Generate(now time.Time, all []model.Opportunity) model.Report {
	items := g.filter(now, all)
	title := fmt.Sprintf("湾区机会日报｜%s", now.Format("2006-01-02"))
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString("面向广东/深圳 AI、软件、数字化、制造业相关企业，筛选政策补贴、项目申报、政府采购与招投标机会。\n\n")
	if len(items) == 0 {
		b.WriteString("今天暂未筛选到达到阈值的机会。可以降低 `--min-score` 或扩大信息源后重新生成。\n\n")
		b.WriteString(disclaimer())
		return model.Report{Date: now, Title: title, Content: b.String(), ItemCount: 0, CreatedAt: time.Now()}
	}
	b.WriteString(fmt.Sprintf("今日筛选到 %d 条值得关注的机会：\n\n", len(items)))
	for i, op := range items {
		b.WriteString(renderItem(i+1, op))
	}
	b.WriteString("## 本期观察\n\n")
	b.WriteString(observation(items))
	b.WriteString("\n\n")
	b.WriteString(disclaimer())
	return model.Report{Date: now, Title: title, Content: b.String(), ItemCount: len(items), CreatedAt: time.Now()}
}

func (g Generator) filter(now time.Time, all []model.Opportunity) []model.Opportunity {
	cutoff := now.AddDate(0, 0, -g.Days)
	items := make([]model.Opportunity, 0, len(all))
	for _, op := range all {
		if op.AnalyzedAt == nil || op.Score < g.MinScore {
			continue
		}
		if op.PublishedAt != nil && op.PublishedAt.Before(cutoff) {
			continue
		}
		items = append(items, op)
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return newest(items[i]).After(newest(items[j]))
		}
		return items[i].Score > items[j].Score
	})
	if g.Limit > 0 && len(items) > g.Limit {
		return items[:g.Limit]
	}
	return items
}

func newest(op model.Opportunity) time.Time {
	if op.PublishedAt != nil {
		return *op.PublishedAt
	}
	return op.CreatedAt
}

func renderItem(index int, op model.Opportunity) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("## %d. %s\n\n", index, op.Title))
	writeLine(&b, "类型", op.Type)
	writeLine(&b, "地区", op.Region)
	writeLine(&b, "金额/预算", op.Budget)
	writeLine(&b, "截止时间", op.Deadline)
	writeLine(&b, "适合对象", op.SuitableFor)
	writeLine(&b, "推荐理由", op.Summary)
	writeLine(&b, "下一步动作", op.NextSteps)
	if len(op.Tags) > 0 {
		writeLine(&b, "标签", strings.Join(op.Tags, "、"))
	}
	writeLine(&b, "机会评分", fmt.Sprintf("%d/100", op.Score))
	writeLine(&b, "官方原文", op.URL)
	b.WriteString("\n")
	return b.String()
}

func writeLine(b *strings.Builder, label, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "未明确"
	}
	b.WriteString(fmt.Sprintf("**%s：** %s\n\n", label, value))
}

func observation(items []model.Opportunity) string {
	if len(items) == 0 {
		return ""
	}
	top := items[0]
	return fmt.Sprintf("今天优先关注「%s」。它的评分最高，命中地区、行业或申报/采购等关键词。建议先打开官方原文核对资格条件、截止时间和材料要求。", top.Title)
}

func disclaimer() string {
	return "## 说明\n\n本内容来自公开官方渠道，经过程序筛选和 AI/规则辅助整理，仅作信息筛选参考，不构成申报、投标、法律或财务建议。具体条件、截止时间、金额和材料要求，请以官方原文为准。\n"
}
