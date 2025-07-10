package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/community-governance-mcp-higress/config"
	"github.com/community-governance-mcp-higress/internal/agent"
	"github.com/community-governance-mcp-higress/utils"
	"github.com/higress-group/wasm-go/pkg/mcp/server"
)

type IssueClassifier struct {
	IssueTitle     string   `json:"issue_title" jsonschema_description:"Issue标题" jsonschema:"example=Bug: 网关启动失败"`
	IssueBody      string   `json:"issue_body" jsonschema_description:"Issue内容描述"`
	ExistingLabels []string `json:"existing_labels,omitempty" jsonschema_description:"现有标签列表"`
	AutoClassify   bool     `json:"auto_classify" jsonschema_description:"是否自动分类" jsonschema:"default=true"`
}

func (t IssueClassifier) Description() string {
	return `Issue 智能分类工具，基于标题和内容自动识别问题类型（bug、feature、documentation等），并推荐合适的标签。`
}

func (t IssueClassifier) InputSchema() map[string]any {
	return server.ToInputSchema(&IssueClassifier{})
}

func (t IssueClassifier) Create(params []byte) server.Tool {
	classifier := &IssueClassifier{
		AutoClassify: true,
	}
	json.Unmarshal(params, classifier)
	return classifier
}

func (t IssueClassifier) Call(ctx server.HttpContext, s server.Server) error {
	serverConfig := &config.CommunityGovernanceConfig{}
	s.GetConfig(serverConfig)

	// 基于规则的分类逻辑
	suggestedLabels := t.classifyIssue()

	if t.AutoClassify && serverConfig.OpenAIKey != "" {
		// 使用 AI 进行更精确的分类
		aiLabels, err := t.aiClassifyIssue(ctx, serverConfig)
		if err == nil {
			suggestedLabels = append(suggestedLabels, aiLabels...)
		}
	}

	// 去重
	uniqueLabels := make(map[string]bool)
	var finalLabels []string
	for _, label := range suggestedLabels {
		if !uniqueLabels[label] {
			uniqueLabels[label] = true
			finalLabels = append(finalLabels, label)
		}
	}

	result := fmt.Sprintf(`Issue 分类结果：  
标题：%s  
推荐标签：%s  
分类依据：基于标题和内容的关键词分析`,
		t.IssueTitle,
		strings.Join(finalLabels, ", "))

	utils.SendMCPToolTextResult(ctx, "IssueClassifier", result, true)
	return nil
}

func (t IssueClassifier) classifyIssue() []string {
	var labels []string

	title := strings.ToLower(t.IssueTitle)
	body := strings.ToLower(t.IssueBody)
	content := title + " " + body

	// Bug 相关关键词
	bugKeywords := []string{"bug", "error", "fail", "crash", "exception", "问题", "错误", "失败", "崩溃", "异常"}
	for _, keyword := range bugKeywords {
		if strings.Contains(content, keyword) {
			labels = append(labels, "bug")
			break
		}
	}

	// Feature 相关关键词
	featureKeywords := []string{"feature", "enhancement", "improve", "add", "新功能", "增强", "改进", "添加"}
	for _, keyword := range featureKeywords {
		if strings.Contains(content, keyword) {
			labels = append(labels, "enhancement")
			break
		}
	}

	// Documentation 相关关键词
	docKeywords := []string{"doc", "documentation", "readme", "guide", "文档", "说明", "指南"}
	for _, keyword := range docKeywords {
		if strings.Contains(content, keyword) {
			labels = append(labels, "documentation")
			break
		}
	}

	// Question 相关关键词
	questionKeywords := []string{"question", "how to", "help", "问题", "如何", "帮助"}
	for _, keyword := range questionKeywords {
		if strings.Contains(content, keyword) {
			labels = append(labels, "question")
			break
		}
	}

	return labels
}

func (t IssueClassifier) aiClassifyIssue(ctx server.HttpContext, config *config.CommunityGovernanceConfig) ([]string, error) {
	// 构建 AI 分类请求
	prompt := fmt.Sprintf(`请分析以下 GitHub Issue 并推荐合适的标签：  
  
标题：%s  
内容：%s  
  
请从以下标签中选择最合适的（可多选）：  
- bug: 软件缺陷  
- enhancement: 功能增强  
- documentation: 文档相关  
- question: 问题咨询  
- good first issue: 适合新手  
- help wanted: 需要帮助  
- priority/high: 高优先级  
- priority/medium: 中优先级  
- priority/low: 低优先级  
  
请只返回标签名称，用逗号分隔。`, t.IssueTitle, t.IssueBody)

	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 100,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	headers := map[string]string{
		"Authorization": "Bearer " + config.OpenAIKey,
		"Content-Type":  "application/json",
	}

	response, err := utils.SendHTTPRequest(ctx, "POST", "https://api.openai.com/v1/chat/completions", headers, string(bodyBytes))
	if err != nil {
		return nil, err
	}

	// 解析 AI 响应
	var aiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
		return nil, err
	}

	choices, ok := aiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, errors.New("AI 响应格式错误")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	// 解析标签
	labels := strings.Split(strings.TrimSpace(content), ",")
	for i, label := range labels {
		labels[i] = strings.TrimSpace(label)
	}

	return labels, nil
}
