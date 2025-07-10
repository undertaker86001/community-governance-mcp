package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client OpenAI客户端
type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string
}

// Message 聊天消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
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

// NewClient 创建新的OpenAI客户端
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, messages []Message, maxTokens int, temperature float64) (*ChatResponse, error) {
	request := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp ErrorResponse
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("解析错误响应失败: %w", err)
		}
		return nil, fmt.Errorf("API错误: %s", errorResp.Error.Message)
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}

// GenerateText 生成文本
func (c *Client) GenerateText(ctx context.Context, prompt string, maxTokens int, temperature float64) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	response, err := c.Chat(ctx, messages, maxTokens, temperature)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("没有生成任何内容")
	}

	return response.Choices[0].Message.Content, nil
}

// AnalyzeText 分析文本
func (c *Client) AnalyzeText(ctx context.Context, text string, analysisType string) (string, error) {
	prompt := fmt.Sprintf("请分析以下文本，分析类型：%s\n\n文本内容：%s\n\n请提供详细的分析结果：", analysisType, text)
	
	return c.GenerateText(ctx, prompt, 1000, 0.3)
}

// SummarizeText 总结文本
func (c *Client) SummarizeText(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf("请总结以下文本的主要内容：\n\n%s\n\n总结：", text)
	
	return c.GenerateText(ctx, prompt, 500, 0.3)
}

// ExtractKeywords 提取关键词
func (c *Client) ExtractKeywords(ctx context.Context, text string) (string, error) {
	prompt := fmt.Sprintf("请从以下文本中提取关键词，以逗号分隔：\n\n%s\n\n关键词：", text)
	
	return c.GenerateText(ctx, prompt, 200, 0.1)
}

// ClassifyText 分类文本
func (c *Client) ClassifyText(ctx context.Context, text string, categories []string) (string, error) {
	categoriesStr := ""
	for i, category := range categories {
		if i > 0 {
			categoriesStr += ", "
		}
		categoriesStr += category
	}
	
	prompt := fmt.Sprintf("请将以下文本分类到以下类别之一：%s\n\n文本：%s\n\n分类结果：", categoriesStr, text)
	
	return c.GenerateText(ctx, prompt, 100, 0.1)
}

// GenerateRecommendations 生成建议
func (c *Client) GenerateRecommendations(ctx context.Context, context string, problem string) (string, error) {
	prompt := fmt.Sprintf("基于以下上下文和问题，请提供具体的建议和解决方案：\n\n上下文：%s\n\n问题：%s\n\n建议：", context, problem)
	
	return c.GenerateText(ctx, prompt, 800, 0.7)
}

// SetModel 设置模型
func (c *Client) SetModel(model string) {
	c.model = model
}

// SetTimeout 设置超时时间
func (c *Client) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
} 