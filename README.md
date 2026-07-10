# 湾区机会日报

我正在开发一个面向广东/深圳科技型中小企业的机会日报项目，主要整理政策补贴、项目申报、政府采购、采购意向与招投标等公开信息。

日报内容会从官方公开渠道中筛选，并用 AI 辅助生成摘要、截止时间、适合对象、推荐理由和下一步动作，帮助 AI、软件、数字化、制造业相关企业减少信息筛选成本，避免错过有价值的申报和投标机会。

爱发电主页：https://afdian.com/a/agentlab

## MVP 功能

- 抓取公开信息源的公告列表和详情页。
- 按关键词筛选广东/深圳 AI、软件、数字化、制造业相关机会。
- 使用 OpenAI-compatible API 分析公告；未配置 API Key 时自动使用本地规则兜底。
- 为机会生成类型、地区、金额/预算、截止时间、适合对象、下一步动作和评分。
- 生成可复制到爱发电/公众号/小报童的 Markdown 日报。

## 快速开始

```bash
go run ./cmd/agent help
go run ./cmd/agent run
```

如果暂时没有 AI API Key，也可以运行。程序会使用本地规则生成摘要和评分：

```bash
go run ./cmd/agent crawl
go run ./cmd/agent analyze
go run ./cmd/agent digest
```

生成结果：

```text
data/opportunities.json          # 本地机会库
data/reports/YYYY-MM-DD.json     # 日报结构化备份
output/YYYY-MM-DD.md             # 可发布的 Markdown 日报
```

## 可选 AI 配置

复制 `.env.example` 并配置环境变量：

```bash
export OPENAI_API_KEY="你的 key"
export OPENAI_BASE_URL="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-4.1-mini"
```

然后运行：

```bash
go run ./cmd/agent run
```

## 命令

```bash
go run ./cmd/agent crawl
go run ./cmd/agent analyze
go run ./cmd/agent digest
go run ./cmd/agent run
```

常用参数：

```bash
go run ./cmd/agent run \
  --sources configs/sources.json \
  --data data \
  --output output \
  --limit-per-source 20 \
  --days 30 \
  --min-score 50 \
  --report-limit 10
```

## 信息源配置

信息源在 `configs/sources.json` 中维护。第一版内置：

- 深圳市科技创新局
- 深圳市工业和信息化局
- 广东省科学技术厅
- 广东省工业和信息化厅
- 中国政府采购网地方公告入口

## 发布流程

第一版建议半自动：

```text
go run ./cmd/agent run
  ↓
检查 output/YYYY-MM-DD.md
  ↓
复制到爱发电公开样刊或会员内容
```

## 免责声明

本项目只整理公开渠道信息，AI/规则摘要仅用于辅助筛选，不构成申报、投标、法律或财务建议。具体条件、截止时间、金额和材料要求，请以官方原文为准。
