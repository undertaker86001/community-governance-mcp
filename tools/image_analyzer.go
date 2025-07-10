package tools

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/community-governance-mcp-higress/internal/model"
	"github.com/community-governance-mcp-higress/internal/openai"
)

// ImageAnalyzer 图片分析器
type ImageAnalyzer struct {
	openaiClient *openai.Client
}

// NewImageAnalyzer 创建新的图片分析器
func NewImageAnalyzer(apiKey string) *ImageAnalyzer {
	return &ImageAnalyzer{
		openaiClient: openai.NewClient(apiKey, "gpt-4o"),
	}
}

// AnalyzeImage 分析图片
func (c *ImageAnalyzer) AnalyzeImage(imageURL string) (*model.ImageAnalysisResult, error) {
	// 验证图片URL
	if err := c.validateImageURL(imageURL); err != nil {
		return nil, fmt.Errorf("图片URL验证失败: %w", err)
	}

	// 使用AI分析图片
	analysis, err := c.aiAnalyzeImage(context.Background(), imageURL)
	if err != nil {
		return nil, fmt.Errorf("AI图片分析失败: %w", err)
	}

	return analysis, nil
}

// validateImageURL 验证图片URL
func (c *ImageAnalyzer) validateImageURL(imageURL string) error {
	if imageURL == "" {
		return fmt.Errorf("图片URL不能为空")
	}

	// 检查URL格式
	if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		return fmt.Errorf("图片URL必须是有效的HTTP/HTTPS链接")
	}

	// 检查图片格式
	validExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	hasValidExtension := false
	for _, ext := range validExtensions {
		if strings.HasSuffix(strings.ToLower(imageURL), ext) {
			hasValidExtension = true
			break
		}
	}

	if !hasValidExtension {
		return fmt.Errorf("不支持的图片格式，支持的格式: %v", validExtensions)
	}

	return nil
}

// aiAnalyzeImage AI分析图片
func (c *ImageAnalyzer) aiAnalyzeImage(ctx context.Context, imageURL string) (*model.ImageAnalysisResult, error) {
	prompt := fmt.Sprintf(`请分析以下图片，并提供详细的分析结果：

图片URL: %s

请提供以下格式的分析结果：
1. 图片描述：详细描述图片中的内容
2. 发现的问题：识别图片中的错误、异常或问题
3. 改进建议：提供具体的改进建议和解决方案

请重点关注：
- 界面布局和设计问题
- 错误信息和警告
- 用户体验问题
- 技术相关问题`, imageURL)

	response, err := c.openaiClient.GenerateText(ctx, prompt, 800, 0.3)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析AI响应
	analysis := c.parseAIResponse(response)

	return analysis, nil
}

// parseAIResponse 解析AI响应
func (c *ImageAnalyzer) parseAIResponse(response string) *model.ImageAnalysisResult {
	analysis := &model.ImageAnalysisResult{
		Description: "图片分析结果",
		Issues:      []string{},
		Suggestions: []string{},
		Confidence:  0.8,
	}

	// 简单的响应解析逻辑
	lines := strings.Split(response, "\n")
	var currentSection string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 识别章节
		if strings.Contains(line, "描述") || strings.Contains(line, "Description") {
			currentSection = "description"
			continue
		}
		if strings.Contains(line, "问题") || strings.Contains(line, "Issues") {
			currentSection = "issues"
			continue
		}
		if strings.Contains(line, "建议") || strings.Contains(line, "Suggestions") {
			currentSection = "suggestions"
			continue
		}

		// 根据章节处理内容
		switch currentSection {
		case "description":
			if analysis.Description == "图片分析结果" {
				analysis.Description = line
			} else {
				analysis.Description += " " + line
			}
		case "issues":
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") {
				issue := strings.TrimPrefix(strings.TrimPrefix(line, "-"), "•")
				issue = strings.TrimSpace(issue)
				if issue != "" {
					analysis.Issues = append(analysis.Issues, issue)
				}
			}
		case "suggestions":
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "•") {
				suggestion := strings.TrimPrefix(strings.TrimPrefix(line, "-"), "•")
				suggestion = strings.TrimSpace(suggestion)
				if suggestion != "" {
					analysis.Suggestions = append(analysis.Suggestions, suggestion)
				}
			}
		}
	}

	// 如果没有解析到具体内容，使用原始响应
	if analysis.Description == "图片分析结果" {
		analysis.Description = response
	}

	return analysis
}

// AnalyzeScreenshot 分析截图
func (c *ImageAnalyzer) AnalyzeScreenshot(imageURL string, contextInfo string) (*model.ImageAnalysisResult, error) {
	// 验证图片URL
	if err := c.validateImageURL(imageURL); err != nil {
		return nil, fmt.Errorf("截图URL验证失败: %w", err)
	}

	// 使用AI分析截图
	analysis, err := c.aiAnalyzeScreenshot(context.Background(), imageURL, contextInfo)
	if err != nil {
		return nil, fmt.Errorf("AI截图分析失败: %w", err)
	}

	return analysis, nil
}

// aiAnalyzeScreenshot AI分析截图
func (c *ImageAnalyzer) aiAnalyzeScreenshot(ctx context.Context, imageURL string, context string) (*model.ImageAnalysisResult, error) {
	prompt := fmt.Sprintf(`请分析以下截图，并提供详细的分析结果：

截图URL: %s
上下文信息: %s

请提供以下格式的分析结果：
1. 截图描述：详细描述截图中的内容
2. 发现的问题：识别截图中的错误、异常或问题
3. 改进建议：提供具体的改进建议和解决方案

请重点关注：
- 界面布局和设计问题
- 错误信息和警告
- 用户体验问题
- 技术相关问题
- 与上下文信息的关联`, imageURL, context)

	response, err := c.openaiClient.GenerateText(ctx, prompt, 800, 0.3)
	if err != nil {
		return nil, fmt.Errorf("AI截图分析失败: %w", err)
	}

	// 解析AI响应
	analysis := c.parseAIResponse(response)

	return analysis, nil
}

// AnalyzeErrorScreenshot 分析错误截图
func (c *ImageAnalyzer) AnalyzeErrorScreenshot(imageURL string, errorContext string) (*model.ImageAnalysisResult, error) {
	// 验证图片URL
	if err := c.validateImageURL(imageURL); err != nil {
		return nil, fmt.Errorf("错误截图URL验证失败: %w", err)
	}

	// 使用AI分析错误截图
	analysis, err := c.enhanceWithErrorContext(context.Background(), imageURL, errorContext)
	if err != nil {
		return nil, fmt.Errorf("AI错误截图分析失败: %w", err)
	}

	return analysis, nil
}

// enhanceWithErrorContext 结合错误上下文增强分析
func (c *ImageAnalyzer) enhanceWithErrorContext(ctx context.Context, imageURL string, errorContext string) (*model.ImageAnalysisResult, error) {
	prompt := fmt.Sprintf(`请结合错误上下文分析以下图片：

图片URL: %s
错误上下文: %s

请提供增强的分析结果，重点关注：
1. 图片中的错误信息与上下文的关联
2. 可能的错误原因和解决方案
3. 预防类似错误的建议`, imageURL, errorContext)

	response, err := c.openaiClient.GenerateText(ctx, prompt, 600, 0.3)
	if err != nil {
		return nil, fmt.Errorf("增强分析失败: %w", err)
	}

	return c.parseAIResponse(response), nil
}

// GetImageInfo 获取图片基本信息
func (c *ImageAnalyzer) GetImageInfo(imageURL string) (map[string]interface{}, error) {
	// 获取图片基本信息
	resp, err := http.Head(imageURL)
	if err != nil {
		return nil, fmt.Errorf("获取图片信息失败: %w", err)
	}
	defer resp.Body.Close()

	info := map[string]interface{}{
		"url":            imageURL,
		"content_type":   resp.Header.Get("Content-Type"),
		"content_length": resp.Header.Get("Content-Length"),
		"last_modified":  resp.Header.Get("Last-Modified"),
	}

	return info, nil
}

// ValidateImageAccessibility 验证图片可访问性
func (c *ImageAnalyzer) ValidateImageAccessibility(imageURL string) (bool, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return false, fmt.Errorf("图片访问失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("图片不可访问，状态码: %d", resp.StatusCode)
	}

	return true, nil
}
