package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// Client MCP客户端
type Client struct {
	httpClient *http.Client
	logger     *logrus.Logger
}

// NewClient 创建新的MCP客户端
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger: logrus.New(),
	}
}

// ListToolsRequest 列出工具请求
type ListToolsRequest struct {
	ServerLabel string            `json:"server_label"`
	ServerURL   string            `json:"server_url"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// ListToolsResponse 列出工具响应
type ListToolsResponse struct {
	Tools []Tool `json:"tools"`
}

// Tool 工具定义
type Tool struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	InputSchema  map[string]interface{} `json:"input_schema"`
	OutputSchema map[string]interface{} `json:"output_schema,omitempty"`
}

// CallToolRequest 调用工具请求
type CallToolRequest struct {
	ServerLabel string            `json:"server_label"`
	ServerURL   string            `json:"server_url"`
	ToolName    string            `json:"tool_name"`
	Arguments   map[string]interface{} `json:"arguments"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// CallToolResponse 调用工具响应
type CallToolResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	ServerLabel string            `json:"server_label"`
	Input       string            `json:"input"`
	Headers     map[string]string `json:"headers,omitempty"`
	RepoName    string            `json:"repo_name,omitempty"`
}

// QueryResponse 查询响应
type QueryResponse struct {
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

// ListTools 列出MCP服务器提供的工具
func (c *Client) ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error) {
	// 构建MCP协议请求
	mcpReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	reqBody, err := json.Marshal(mcpReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", req.ServerURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析MCP响应
	var mcpResp map[string]interface{}
	if err := json.Unmarshal(respBody, &mcpResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if errorObj, exists := mcpResp["error"]; exists && errorObj != nil {
		return nil, fmt.Errorf("MCP服务器错误: %v", errorObj)
	}

	// 提取工具列表
	result, exists := mcpResp["result"]
	if !exists {
		return nil, fmt.Errorf("响应中缺少result字段")
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	var tools []Tool
	if err := json.Unmarshal(resultBytes, &tools); err != nil {
		return nil, fmt.Errorf("解析工具列表失败: %w", err)
	}

	return &ListToolsResponse{Tools: tools}, nil
}

// CallTool 调用MCP工具
func (c *Client) CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error) {
	// 构建MCP协议请求
	mcpReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      req.ToolName,
			"arguments": req.Arguments,
		},
	}

	reqBody, err := json.Marshal(mcpReq)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 创建HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", req.ServerURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	// 设置请求头
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 解析MCP响应
	var mcpResp map[string]interface{}
	if err := json.Unmarshal(respBody, &mcpResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查错误
	if errorObj, exists := mcpResp["error"]; exists && errorObj != nil {
		errorBytes, _ := json.Marshal(errorObj)
		return &CallToolResponse{Error: string(errorBytes)}, nil
	}

	// 提取结果
	result, exists := mcpResp["result"]
	if !exists {
		return nil, fmt.Errorf("响应中缺少result字段")
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("序列化结果失败: %w", err)
	}

	return &CallToolResponse{Output: string(resultBytes)}, nil
}

// Query 执行查询（针对DeepWiki等特定服务器）
func (c *Client) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	// 构建工具调用参数
	arguments := map[string]interface{}{
		"question": req.Input,
	}

	// 如果是DeepWiki且有仓库名，添加仓库参数
	if req.ServerLabel == "deepwiki" && req.RepoName != "" {
		arguments["repoName"] = req.RepoName
	}

	// 调用工具
	callReq := &CallToolRequest{
		ServerLabel: req.ServerLabel,
		ServerURL:   getServerURL(req.ServerLabel),
		ToolName:    "ask_question",
		Arguments:   arguments,
		Headers:     req.Headers,
	}

	callResp, err := c.CallTool(ctx, callReq)
	if err != nil {
		return nil, err
	}

	if callResp.Error != "" {
		return &QueryResponse{Error: callResp.Error}, nil
	}

	return &QueryResponse{Output: callResp.Output}, nil
}

// getServerURL 根据服务器标签获取URL
func getServerURL(serverLabel string) string {
	servers := map[string]string{
		"deepwiki": "https://mcp.deepwiki.com/mcp",
		"stripe":   "https://mcp.stripe.com",
		"shopify":  "https://mcp.shopify.com",
		"twilio":   "https://mcp.twilio.com",
	}

	if url, exists := servers[serverLabel]; exists {
		return url
	}

	return ""
} 