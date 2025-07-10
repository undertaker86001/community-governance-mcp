package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Client OpenAI客户端
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewClient 创建新的OpenAI客户端
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logrus.New(),
	}
}

// MCPTool MCP工具配置
type MCPTool struct {
	Type           string            `json:"type"`
	ServerLabel    string            `json:"server_label"`
	ServerURL      string            `json:"server_url"`
	RequireApproval string           `json:"require_approval,omitempty"`
	AllowedTools   []string          `json:"allowed_tools,omitempty"`
	Headers        map[string]string `json:"headers,omitempty"`
}

// ResponsesRequest OpenAI Responses API请求
type ResponsesRequest struct {
	Model            string      `json:"model"`
	Tools            []MCPTool   `json:"tools"`
	Input            interface{} `json:"input"`
	PreviousResponseID string    `json:"previous_response_id,omitempty"`
}

// ResponsesResponse OpenAI Responses API响应
type ResponsesResponse struct {
	ID        string    `json:"id"`
	Model     string    `json:"model"`
	CreatedAt int64     `json:"created_at"`
	Output    []Output  `json:"output"`
}

// Output 输出项
type Output struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	ServerLabel   string                 `json:"server_label,omitempty"`
	Tools         []Tool                 `json:"tools,omitempty"`
	Name          string                 `json:"name,omitempty"`
	Arguments     string                 `json:"arguments,omitempty"`
	Output        string                 `json:"output,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Approve       bool                   `json:"approve,omitempty"`
	ApprovalRequestID string             `json:"approval_request_id,omitempty"`
	Content       []Content              `json:"content,omitempty"`
}

// Tool 工具定义
type Tool struct {
	Name         string                 `json:"name"`
	InputSchema  map[string]interface{} `json:"input_schema"`
	Description  string                 `json:"description"`
}

// Content 内容项
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// DeepWikiQuestion DeepWiki问题
type DeepWikiQuestion struct {
	RepoName string `json:"repoName"`
	Question string `json:"question"`
}

// CreateResponse 创建Responses API请求
func (c *Client) CreateResponse(ctx context.Context, input string, tools []MCPTool) (*ResponsesResponse, error) {
	// 构建请求体
	request := ResponsesRequest{
		Model: "gpt-4o",
		Tools: tools,
		Input: input,
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/responses", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	// 解析响应
	var response ResponsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// AskDeepWiki 向DeepWiki提问
func (c *Client) AskDeepWiki(ctx context.Context, repoName, question string) (string, error) {
	// 配置DeepWiki MCP工具
	tools := []MCPTool{
		{
			Type:           "mcp",
			ServerLabel:    "deepwiki",
			ServerURL:      "https://mcp.deepwiki.com/mcp",
			RequireApproval: "never",
		},
	}

	// 构建问题
	input := fmt.Sprintf("关于仓库 %s 的问题: %s", repoName, question)

	// 创建响应
	response, err := c.CreateResponse(ctx, input, tools)
	if err != nil {
		return "", fmt.Errorf("failed to create response: %w", err)
	}

	// 提取回答内容
	var answer string
	for _, output := range response.Output {
		if output.Type == "text" && len(output.Content) > 0 {
			for _, content := range output.Content {
				if content.Type == "text" {
					answer += content.Text
				}
			}
		}
	}

	if answer == "" {
		return "", fmt.Errorf("no answer found in response")
	}

	return answer, nil
}

// ProcessWithApproval 处理需要审批的MCP调用
func (c *Client) ProcessWithApproval(ctx context.Context, previousResponseID, approvalRequestID string, approve bool) (*ResponsesResponse, error) {
	// 构建审批响应
	approvalResponse := map[string]interface{}{
		"type": "mcp_approval_response",
		"approve": approve,
		"approval_request_id": approvalRequestID,
	}

	// 配置DeepWiki MCP工具
	tools := []MCPTool{
		{
			Type:           "mcp",
			ServerLabel:    "deepwiki",
			ServerURL:      "https://mcp.deepwiki.com/mcp",
		},
	}

	// 构建请求
	request := ResponsesRequest{
		Model: "gpt-4o",
		Tools: tools,
		Input: []interface{}{approvalResponse},
		PreviousResponseID: previousResponseID,
	}

	// 序列化请求
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approval request: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/responses", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send approval request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("approval API request failed with status: %d", resp.StatusCode)
	}

	// 解析响应
	var response ResponsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode approval response: %w", err)
	}

	return &response, nil
}

// GetMCPTools 获取MCP服务器提供的工具列表
func (c *Client) GetMCPTools(ctx context.Context, tools []MCPTool) ([]Tool, error) {
	// 创建响应以获取工具列表
	response, err := c.CreateResponse(ctx, "List available tools", tools)
	if err != nil {
		return nil, fmt.Errorf("failed to get MCP tools: %w", err)
	}

	// 从输出中提取工具列表
	var availableTools []Tool
	for _, output := range response.Output {
		if output.Type == "mcp_list_tools" {
			availableTools = append(availableTools, output.Tools...)
		}
	}

	return availableTools, nil
}

// SetLogger 设置日志记录器
func (c *Client) SetLogger(logger *logrus.Logger) {
	c.logger = logger
} 