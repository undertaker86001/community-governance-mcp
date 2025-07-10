package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"community-governance-mcp-higress/internal/agent"
)

// ImageAnalyzer 图片分析器
type ImageAnalyzer struct {
	config *agent.AgentConfig
}

// NewImageAnalyzer 创建新的图片分析器
func NewImageAnalyzer() *ImageAnalyzer {
	return &ImageAnalyzer{}
}

// SetConfig 设置配置
func (i *ImageAnalyzer) SetConfig(config *agent.AgentConfig) {
	i.config = config
}

// Analyze 分析图片
func (i *ImageAnalyzer) Analyze(ctx context.Context, imageURL string) (*agent.ImageAnalysis, error) {
	if i.config == nil || i.config.OpenAI.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API密钥未配置")
	}

	// 根据图片URL分析类型
	analysisType := i.determineAnalysisType(imageURL)
	
	// 调用AI分析
	analysis, err := i.analyzeImage(ctx, imageURL, analysisType)
	if err != nil {
		return nil, err
	}

	return analysis, nil
}

// determineAnalysisType 确定分析类型
func (i *ImageAnalyzer) determineAnalysisType(imageURL string) string {
	// 基于URL或文件名判断分析类型
	if i.containsKeywords(imageURL, []string{"error", "screenshot", "bug"}) {
		return "error_screenshot"
	}
	if i.containsKeywords(imageURL, []string{"arch", "diagram", "architecture"}) {
		return "architecture_diagram"
	}
	if i.containsKeywords(imageURL, []string{"log", "console", "terminal"}) {
		return "log_image"
	}
	return "general"
}

// containsKeywords 检查是否包含关键词
func (i *ImageAnalyzer) containsKeywords(text string, keywords []string) bool {
	text = fmt.Sprintf("%s", text)
	for _, keyword := range keywords {
		if i.contains(text, keyword) {
			return true
		}
	}
	return false
}

// contains 简单的字符串包含检查
func (i *ImageAnalyzer) contains(text, keyword string) bool {
	return len(text) >= len(keyword) && 
		(text == keyword || 
		 (len(text) > len(keyword) && 
		  (text[:len(keyword)] == keyword || 
		   text[len(text)-len(keyword):] == keyword ||
		   i.containsSubstring(text, keyword))))
}

// containsSubstring 检查子字符串
func (i *ImageAnalyzer) containsSubstring(text, keyword string) bool {
	for i := 0; i <= len(text)-len(keyword); i++ {
		if text[i:i+len(keyword)] == keyword {
			return true
		}
	}
	return false
}

// analyzeImage 分析图片
func (i *ImageAnalyzer) analyzeImage(ctx context.Context, imageURL string, analysisType string) (*agent.ImageAnalysis, error) {
	prompt := i.generatePrompt(analysisType)

	requestBody := map[string]interface{}{
		"model": i.config.OpenAI.Model,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
				},
			},
		},
		"max_tokens": 500,
		"temperature": 0.2,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	headers := map[string]string{
		"Authorization": "Bearer " + i.config.OpenAI.APIKey,
		"Content-Type":  "application/json",
	}

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var aiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		return nil, err
	}

	choices, ok := aiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("AI响应格式错误")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	// 解析分析结果
	analysis := &agent.ImageAnalysis{
		DetectedElements: i.extractElements(content),
		ErrorMessages:    i.extractErrors(content),
		UIElements:       i.extractUIElements(content),
		Suggestions:      i.extractSuggestions(content),
		Confidence:       0.8,
	}

	return analysis, nil
}

// generatePrompt 生成分析提示
func (i *ImageAnalyzer) generatePrompt(analysisType string) string {
	switch analysisType {
	case "error_screenshot":
		return `这是一个错误截图。请分析图片中的错误信息，包括：
1. 错误类型和错误代码
2. 可能的原因分析
3. 建议的解决步骤
4. 相关的调试方法

请提供详细的分析结果。`
	case "architecture_diagram":
		return `这是一个系统架构图。请分析图片中的架构设计，包括：
1. 系统组件和模块
2. 数据流和调用关系
3. 可能的性能瓶颈
4. 架构优化建议
5. 潜在的问题点

请提供详细的架构分析。`
	case "log_image":
		return `这是一个日志截图。请分析图片中的日志信息，包括：
1. 提取关键的日志内容
2. 识别错误和警告信息
3. 分析日志模式和趋势
4. 提供问题诊断建议

请提供详细的日志分析结果。`
	default:
		return `请分析这张图片的内容，特别关注技术相关的信息，包括：
1. 图片中的文字内容
2. 技术组件或界面元素
3. 可能的问题或异常
4. 相关的技术建议

请提供详细的分析结果。`
	}
}

// extractElements 提取检测到的元素
func (i *ImageAnalyzer) extractElements(content string) []string {
	// 简化版本，实际应该使用更复杂的解析逻辑
	elements := []string{}
	if i.contains(content, "error") {
		elements = append(elements, "error_message")
	}
	if i.contains(content, "button") {
		elements = append(elements, "button")
	}
	if i.contains(content, "form") {
		elements = append(elements, "form")
	}
	if i.contains(content, "table") {
		elements = append(elements, "table")
	}
	return elements
}

// extractErrors 提取错误信息
func (i *ImageAnalyzer) extractErrors(content string) []string {
	// 简化版本，实际应该使用更复杂的解析逻辑
	errors := []string{}
	if i.contains(content, "error") {
		errors = append(errors, "检测到错误信息")
	}
	if i.contains(content, "warning") {
		errors = append(errors, "检测到警告信息")
	}
	if i.contains(content, "exception") {
		errors = append(errors, "检测到异常信息")
	}
	return errors
}

// extractUIElements 提取UI元素
func (i *ImageAnalyzer) extractUIElements(content string) []string {
	// 简化版本，实际应该使用更复杂的解析逻辑
	elements := []string{}
	if i.contains(content, "button") {
		elements = append(elements, "按钮")
	}
	if i.contains(content, "input") {
		elements = append(elements, "输入框")
	}
	if i.contains(content, "menu") {
		elements = append(elements, "菜单")
	}
	if i.contains(content, "dialog") {
		elements = append(elements, "对话框")
	}
	return elements
}

// extractSuggestions 提取建议
func (i *ImageAnalyzer) extractSuggestions(content string) []string {
	// 简化版本，实际应该使用更复杂的解析逻辑
	suggestions := []string{}
	if i.contains(content, "error") {
		suggestions = append(suggestions, "建议检查错误日志")
	}
	if i.contains(content, "performance") {
		suggestions = append(suggestions, "建议优化性能")
	}
	if i.contains(content, "security") {
		suggestions = append(suggestions, "建议加强安全措施")
	}
	if len(suggestions) == 0 {
		suggestions = append(suggestions, "建议进一步分析图片内容")
	}
	return suggestions
}
