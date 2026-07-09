package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"bay-area-opportunity-daily/internal/model"
)

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (a Analyzer) analyzeWithOpenAI(ctx context.Context, op model.Opportunity) (model.Analysis, error) {
	body := op.Body
	runes := []rune(body)
	if len(runes) > 5000 {
		body = string(runes[:5000])
	}
	prompt := fmt.Sprintf(`请从以下公告中提取结构化机会信息，只输出 JSON，不要输出解释。

JSON 字段：
{
  "type": "政策补贴/项目申报/采购意向/招投标/产业空间/政策通知",
  "region": "地区",
  "deadline": "截止时间，找不到填空字符串",
  "budget": "金额或预算，找不到填空字符串",
  "tags": ["标签"],
  "summary": "100-160字摘要",
  "suitable_for": "适合对象",
  "next_steps": "下一步动作建议",
  "score": 0
}

要求：
- 不要编造公告没有的信息。
- 截止时间、金额以原文为准。
- 适合对象优先面向广东/深圳 AI、软件、数字化、制造业中小企业。

标题：%s
来源：%s
地区：%s
正文：
%s`, op.Title, op.Source, op.Region, body)

	reqBody := chatRequest{
		Model: a.cfg.OpenAIModel,
		Messages: []chatMessage{
			{Role: "system", Content: "你是政策补贴、项目申报和招投标机会分析助手。"},
			{Role: "user", Content: prompt},
		},
	}
	payload, _ := json.Marshal(reqBody)
	endpoint := strings.TrimRight(a.cfg.OpenAIBaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return model.Analysis{}, err
	}
	req.Header.Set("Authorization", "Bearer "+a.cfg.OpenAIAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model.Analysis{}, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return model.Analysis{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.Analysis{}, fmt.Errorf("llm status %d: %s", resp.StatusCode, string(b))
	}
	var cr chatResponse
	if err := json.Unmarshal(b, &cr); err != nil {
		return model.Analysis{}, err
	}
	if len(cr.Choices) == 0 {
		return model.Analysis{}, errors.New("empty choices")
	}
	return parseAnalysisJSON(cr.Choices[0].Message.Content)
}

func parseAnalysisJSON(content string) (model.Analysis, error) {
	content = strings.TrimSpace(content)
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end < start {
		return model.Analysis{}, errors.New("no json object")
	}
	var analysis model.Analysis
	if err := json.Unmarshal([]byte(content[start:end+1]), &analysis); err != nil {
		return model.Analysis{}, err
	}
	return analysis, nil
}
