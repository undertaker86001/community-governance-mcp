package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/community-governance-mcp-higress/internal/model"
	"github.com/community-governance-mcp-higress/internal/openai"
)

// IssueClassifier Issue分类器
type IssueClassifier struct {
	openaiClient *openai.Client
}

// NewIssueClassifier 创建新的Issue分类器
func NewIssueClassifier(apiKey string) *IssueClassifier {
	return &IssueClassifier{
		openaiClient: openai.NewClient(apiKey, "gpt-4o"),
	}
}

// ClassifyIssue 分类Issue
func (c *IssueClassifier) ClassifyIssue(title string, body string, labels []string) (*model.IssueClassification, error) {
	// 构建分类提示
	prompt := c.buildClassificationPrompt(title, body, labels)

	// 使用AI进行分类
	response, err := c.openaiClient.GenerateText(context.Background(), prompt, 600, 0.3)
	if err != nil {
		return nil, fmt.Errorf("AI分类失败: %w", err)
	}

	// 解析分类结果
	classification := c.parseClassificationResponse(response)

	return classification, nil
}

// buildClassificationPrompt 构建分类提示
func (c *IssueClassifier) buildClassificationPrompt(title string, body string, labels []string) string {
	labelsStr := strings.Join(labels, ", ")
	if labelsStr == "" {
		labelsStr = "无标签"
	}

	return fmt.Sprintf(`请分析以下GitHub Issue并进行分类：

标题: %s
内容: %s
现有标签: %s

请提供以下格式的分类结果：
{
  "category": "bug|feature|documentation|enhancement|question|other",
  "priority": "high|medium|low",
  "severity": "critical|major|minor|trivial",
  "type": "bug|feature|improvement|task|epic",
  "labels": ["建议的标签1", "建议的标签2"],
  "confidence": 0.95,
  "reasoning": "分类理由"
}

分类规则：
- bug: 错误报告、崩溃、异常行为
- feature: 新功能请求、功能增强
- documentation: 文档相关、说明文档
- enhancement: 改进现有功能
- question: 问题咨询、使用疑问
- other: 其他类型

优先级规则：
- high: 影响核心功能、安全漏洞、严重错误
- medium: 一般功能问题、性能问题
- low: 小问题、优化建议、文档完善

严重程度规则：
- critical: 系统崩溃、数据丢失、安全漏洞
- major: 主要功能失效、性能严重下降
- minor: 小功能问题、UI问题
- trivial: 拼写错误、格式问题`, title, body, labelsStr)
}

// parseClassificationResponse 解析分类响应
func (c *IssueClassifier) parseClassificationResponse(response string) *model.IssueClassification {
	classification := &model.IssueClassification{
		Category:   "other",
		Priority:   "medium",
		Severity:   "minor",
		Type:       "task",
		Labels:     []string{},
		Confidence: 0.8,
		Reasoning:  "AI分析结果",
	}

	// 尝试解析JSON响应
	if strings.Contains(response, "{") && strings.Contains(response, "}") {
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}") + 1
		if start >= 0 && end > start {
			jsonStr := response[start:end]

			var result map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
				if category, ok := result["category"].(string); ok {
					classification.Category = category
				}
				if priority, ok := result["priority"].(string); ok {
					classification.Priority = priority
				}
				if severity, ok := result["severity"].(string); ok {
					classification.Severity = severity
				}
				if issueType, ok := result["type"].(string); ok {
					classification.Type = issueType
				}
				if confidence, ok := result["confidence"].(float64); ok {
					classification.Confidence = confidence
				}
				if reasoning, ok := result["reasoning"].(string); ok {
					classification.Reasoning = reasoning
				}
				if labels, ok := result["labels"].([]interface{}); ok {
					for _, label := range labels {
						if labelStr, ok := label.(string); ok {
							classification.Labels = append(classification.Labels, labelStr)
						}
					}
				}
			}
		}
	}

	// 如果没有解析到JSON，使用文本分析
	if classification.Category == "other" {
		classification = c.fallbackTextAnalysis(response)
	}

	return classification
}

// fallbackTextAnalysis 备用文本分析
func (c *IssueClassifier) fallbackTextAnalysis(response string) *model.IssueClassification {
	classification := &model.IssueClassification{
		Category:   "other",
		Priority:   "medium",
		Severity:   "minor",
		Type:       "task",
		Labels:     []string{},
		Confidence: 0.6,
		Reasoning:  response,
	}

	response = strings.ToLower(response)

	// 简单的关键词匹配
	if strings.Contains(response, "bug") || strings.Contains(response, "error") || strings.Contains(response, "crash") {
		classification.Category = "bug"
		classification.Type = "bug"
	}
	if strings.Contains(response, "feature") || strings.Contains(response, "new") || strings.Contains(response, "add") {
		classification.Category = "feature"
		classification.Type = "feature"
	}
	if strings.Contains(response, "doc") || strings.Contains(response, "documentation") {
		classification.Category = "documentation"
		classification.Type = "task"
	}
	if strings.Contains(response, "question") || strings.Contains(response, "how") || strings.Contains(response, "what") {
		classification.Category = "question"
		classification.Type = "task"
	}

	// 优先级分析
	if strings.Contains(response, "high") || strings.Contains(response, "critical") || strings.Contains(response, "urgent") {
		classification.Priority = "high"
	}
	if strings.Contains(response, "low") || strings.Contains(response, "minor") {
		classification.Priority = "low"
	}

	// 严重程度分析
	if strings.Contains(response, "critical") || strings.Contains(response, "crash") || strings.Contains(response, "security") {
		classification.Severity = "critical"
	}
	if strings.Contains(response, "major") || strings.Contains(response, "broken") {
		classification.Severity = "major"
	}
	if strings.Contains(response, "minor") || strings.Contains(response, "small") {
		classification.Severity = "minor"
	}

	return classification
}

// ClassifyMultipleIssues 批量分类Issue
func (c *IssueClassifier) ClassifyMultipleIssues(issues []model.IssueInfo) ([]*model.IssueClassification, error) {
	var classifications []*model.IssueClassification

	for _, issue := range issues {
		classification, err := c.ClassifyIssue(issue.Title, issue.Body, issue.Labels)
		if err != nil {
			// 记录错误但继续处理其他Issue
			fmt.Printf("分类Issue失败: %v\n", err)
			continue
		}
		classifications = append(classifications, classification)
	}

	return classifications, nil
}

// GetClassificationStats 获取分类统计
func (c *IssueClassifier) GetClassificationStats(classifications []*model.IssueClassification) *model.ClassificationStats {
	stats := &model.ClassificationStats{
		CategoryCounts: make(map[string]int),
		PriorityCounts: make(map[string]int),
		SeverityCounts: make(map[string]int),
		TypeCounts:     make(map[string]int),
		TotalIssues:    len(classifications),
	}

	for _, classification := range classifications {
		// 统计分类
		stats.CategoryCounts[classification.Category]++

		// 统计优先级
		stats.PriorityCounts[classification.Priority]++

		// 统计严重程度
		stats.SeverityCounts[classification.Severity]++

		// 统计类型
		stats.TypeCounts[classification.Type]++

		// 计算平均置信度
		stats.AverageConfidence += classification.Confidence
	}

	if stats.TotalIssues > 0 {
		stats.AverageConfidence /= float64(stats.TotalIssues)
	}

	return stats
}

// SuggestLabels 建议标签
func (c *IssueClassifier) SuggestLabels(title string, body string) ([]string, error) {
	prompt := fmt.Sprintf(`请为以下GitHub Issue建议合适的标签：

标题: %s
内容: %s

请提供5-10个最合适的标签，用逗号分隔。标签应该简洁明了，能够准确描述Issue的类型和内容。`, title, body)

	response, err := c.openaiClient.GenerateText(context.Background(), prompt, 300, 0.3)
	if err != nil {
		return nil, fmt.Errorf("AI标签建议失败: %w", err)
	}

	// 解析标签
	labels := c.parseLabels(response)
	return labels, nil
}

// parseLabels 解析标签
func (c *IssueClassifier) parseLabels(response string) []string {
	var labels []string

	// 移除多余的标点符号和空格
	response = strings.TrimSpace(response)
	response = strings.Trim(response, ".,;:!?")

	// 按逗号分割
	parts := strings.Split(response, ",")

	for _, part := range parts {
		label := strings.TrimSpace(part)
		label = strings.Trim(label, ".,;:!?")

		// 移除引号
		label = strings.Trim(label, `"'`)

		if label != "" {
			labels = append(labels, label)
		}
	}

	return labels
}

// ValidateClassification 验证分类结果
func (c *IssueClassifier) ValidateClassification(classification *model.IssueClassification) error {
	// 验证分类
	validCategories := []string{"bug", "feature", "documentation", "enhancement", "question", "other"}
	if !contains(validCategories, classification.Category) {
		return fmt.Errorf("无效的分类: %s", classification.Category)
	}

	// 验证优先级
	validPriorities := []string{"high", "medium", "low"}
	if !contains(validPriorities, classification.Priority) {
		return fmt.Errorf("无效的优先级: %s", classification.Priority)
	}

	// 验证严重程度
	validSeverities := []string{"critical", "major", "minor", "trivial"}
	if !contains(validSeverities, classification.Severity) {
		return fmt.Errorf("无效的严重程度: %s", classification.Severity)
	}

	// 验证类型
	validTypes := []string{"bug", "feature", "improvement", "task", "epic"}
	if !contains(validTypes, classification.Type) {
		return fmt.Errorf("无效的类型: %s", classification.Type)
	}

	// 验证置信度
	if classification.Confidence < 0.0 || classification.Confidence > 1.0 {
		return fmt.Errorf("置信度必须在0-1之间: %f", classification.Confidence)
	}

	return nil
}

// contains 检查切片是否包含元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
