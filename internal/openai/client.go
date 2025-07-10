package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"community-governance-mcp-higress/internal/agent"
)

// Client OpenAI客户端
type Client struct {
	config *agent.OpenAIConfig
	client *http.Client
}

// NewClient 创建新的OpenAI客户端
func NewClient(config *agent.OpenAIConfig) *Client {
	return &Client{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code,omitempty"`
	} `json:"error"`
}

// GenerateAnswer 生成回答
func (c *Client) GenerateAnswer(ctx context.Context, question string, context string) (string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "你是一个专业的Higress社区治理助手，专门帮助用户解决Higress相关的问题。请基于提供的上下文信息，给出准确、有用的回答。",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("上下文信息：%s\n\n问题：%s", context, question),
		},
	}

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: c.config.Temperature,
	}

	response, err := c.chat(ctx, request)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("没有生成回答")
	}

	return response.Choices[0].Message.Content, nil
}

// GenerateSummary 生成摘要
func (c *Client) GenerateSummary(ctx context.Context, content string) (string, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "你是一个专业的文本摘要助手。请为提供的内容生成简洁、准确的摘要。",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("请为以下内容生成摘要：\n\n%s", content),
		},
	}

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   200,
		Temperature: 0.3,
	}

	response, err := c.chat(ctx, request)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("没有生成摘要")
	}

	return response.Choices[0].Message.Content, nil
}

// AnalyzeBug 分析Bug
func (c *Client) AnalyzeBug(ctx context.Context, stackTrace string, environment string) (*agent.BugAnalysis, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "你是一个专业的Bug分析助手。请分析提供的错误堆栈信息，识别错误类型、严重程度、根本原因，并提供解决方案和预防措施。",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("环境信息：%s\n\n错误堆栈：\n%s\n\n请分析这个错误。", environment, stackTrace),
		},
	}

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: 0.2,
	}

	response, err := c.chat(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("没有生成分析结果")
	}

	// 解析分析结果
	analysis := &agent.BugAnalysis{
		ErrorType:  "unknown",
		Language:   "unknown",
		Severity:   "medium",
		RootCause:  "需要进一步分析",
		Solutions:  []string{"请提供更多详细信息"},
		Prevention: []string{"定期检查系统状态"},
		Confidence: 0.5,
	}

	// 这里可以添加更复杂的解析逻辑
	// 或者使用结构化的提示来获得JSON格式的响应

	return analysis, nil
}

// AnalyzeImage 分析图片
func (c *Client) AnalyzeImage(ctx context.Context, imageURL string) (*agent.ImageAnalysis, error) {
	// 注意：这里需要支持图片分析的模型，如GPT-4V
	messages := []Message{
		{
			Role: "system",
			Content: "你是一个专业的图片分析助手。请分析提供的图片，识别界面元素、错误信息、UI问题等，并提供改进建议。",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("请分析这张图片：%s", imageURL),
		},
	}

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: 0.2,
	}

	response, err := c.chat(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("没有生成分析结果")
	}

	// 解析分析结果
	analysis := &agent.ImageAnalysis{
		DetectedElements: []string{},
		ErrorMessages:    []string{},
		UIElements:       []string{},
		Suggestions:      []string{"请提供更清晰的截图"},
		Confidence:       0.5,
	}

	return analysis, nil
}

// ClassifyIssue 分类Issue
func (c *Client) ClassifyIssue(ctx context.Context, issueContent string) (*agent.IssueClassification, error) {
	messages := []Message{
		{
			Role: "system",
			Content: "你是一个专业的Issue分类助手。请分析提供的Issue内容，确定其类别、优先级、标签、建议的负责人等。",
		},
		{
			Role: "user",
			Content: fmt.Sprintf("请分类这个Issue：\n\n%s", issueContent),
		},
	}

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   c.config.MaxTokens,
		Temperature: 0.2,
	}

	response, err := c.chat(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("没有生成分类结果")
	}

	// 解析分类结果
	classification := &agent.IssueClassification{
		Category:   "general",
		Priority:   "normal",
		Labels:     []string{},
		Assignees:  []string{},
		Milestone:  "",
		Confidence: 0.5,
	}

	return classification, nil
}

// chat 发送聊天请求
func (c *Client) chat(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	// 构建请求体
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("API请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("API请求失败: %s", errorResp.Error.Message)
	}

	// 解析响应
	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}

// TestConnection 测试连接
func (c *Client) TestConnection(ctx context.Context) error {
	messages := []Message{
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	request := ChatRequest{
		Model:       c.config.Model,
		Messages:    messages,
		MaxTokens:   10,
		Temperature: 0.0,
	}

	_, err := c.chat(ctx, request)
	return err
} 