package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/community-governance-mcp-higress/internal/model"
	"github.com/sirupsen/logrus"
)

// Manager MCP管理器
type Manager struct {
	clients    map[string]*Client
	config     *model.MCPConfig
	logger     *logrus.Logger
	mutex      sync.RWMutex
	cache      map[string]*CacheEntry
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// NewManager 创建新的MCP管理器
func NewManager(config *model.MCPConfig) *Manager {
	manager := &Manager{
		clients: make(map[string]*Client),
		config:  config,
		logger:  logrus.New(),
		cache:   make(map[string]*CacheEntry),
	}

	// 初始化已启用的MCP服务器客户端
	if config != nil {
		for serverLabel, serverConfig := range config.Servers {
			if serverConfig.Enabled {
				client := NewClient(30 * time.Second)
				manager.clients[serverLabel] = client
				manager.logger.WithField("server", serverLabel).Info("MCP服务器客户端已初始化")
			}
		}
	}

	return manager
}

// GetClient 获取MCP客户端
func (m *Manager) GetClient(serverLabel string) (*Client, error) {
	m.mutex.RLock()
	client, exists := m.clients[serverLabel]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("MCP服务器 %s 未找到或未启用", serverLabel)
	}

	return client, nil
}

// Query 执行MCP查询
func (m *Manager) Query(ctx context.Context, serverLabel, input string, repoName string) (*QueryResponse, error) {
	client, err := m.GetClient(serverLabel)
	if err != nil {
		return nil, err
	}

	// 构建查询请求
	req := &QueryRequest{
		ServerLabel: serverLabel,
		Input:       input,
		RepoName:    repoName,
	}

	// 执行查询
	return client.Query(ctx, req)
}

// ListTools 获取MCP服务器工具列表
func (m *Manager) ListTools(ctx context.Context, serverLabel string) (*ListToolsResponse, error) {
	client, err := m.GetClient(serverLabel)
	if err != nil {
		return nil, err
	}

	// 获取服务器配置
	serverConfig, exists := m.config.Servers[serverLabel]
	if !exists {
		return nil, fmt.Errorf("服务器配置未找到: %s", serverLabel)
	}

	// 构建请求
	req := &ListToolsRequest{
		ServerLabel: serverLabel,
		ServerURL:   serverConfig.ServerURL,
		Headers:     serverConfig.Headers,
	}

	// 执行请求
	return client.ListTools(ctx, req)
}

// CallTool 调用MCP工具
func (m *Manager) CallTool(ctx context.Context, serverLabel, toolName string, arguments map[string]interface{}) (*CallToolResponse, error) {
	client, err := m.GetClient(serverLabel)
	if err != nil {
		return nil, err
	}

	// 获取服务器配置
	serverConfig, exists := m.config.Servers[serverLabel]
	if !exists {
		return nil, fmt.Errorf("服务器配置未找到: %s", serverLabel)
	}

	// 构建请求
	req := &CallToolRequest{
		ServerLabel: serverLabel,
		ServerURL:   serverConfig.ServerURL,
		ToolName:    toolName,
		Arguments:   arguments,
		Headers:     serverConfig.Headers,
	}

	// 执行请求
	return client.CallTool(ctx, req)
}

// QueryWithFallback 执行带备用方案的查询
func (m *Manager) QueryWithFallback(ctx context.Context, serverLabel, input string, repoName string, fallbackFunc func() ([]model.KnowledgeItem, error)) ([]model.KnowledgeItem, error) {
	// 尝试MCP查询
	queryResp, err := m.Query(ctx, serverLabel, input, repoName)
	if err != nil {
		m.logger.WithError(err).Warn("MCP查询失败，使用备用方案")
		return fallbackFunc()
	}

	if queryResp.Error != "" {
		m.logger.WithField("error", queryResp.Error).Warn("MCP查询返回错误，使用备用方案")
		return fallbackFunc()
	}

	// 解析MCP响应为KnowledgeItem
	items := m.parseMCPResponseToKnowledgeItems(queryResp.Output, model.KnowledgeSourceDeepWiki)
	return items, nil
}

// parseMCPResponseToKnowledgeItems 解析MCP响应为KnowledgeItem
func (m *Manager) parseMCPResponseToKnowledgeItems(mcpOutput string, source model.KnowledgeSource) []model.KnowledgeItem {
	var items []model.KnowledgeItem

	// 尝试解析JSON响应
	var response struct {
		Results []struct {
			Title   string  `json:"title"`
			Content string  `json:"content"`
			URL     string  `json:"url"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}

	if err := json.Unmarshal([]byte(mcpOutput), &response); err == nil {
		// 成功解析JSON
		for _, result := range response.Results {
			item := model.KnowledgeItem{
				ID:        fmt.Sprintf("mcp_%s_%s", source, result.Title),
				Source:    source,
				Title:     result.Title,
				Content:   result.Content,
				URL:       result.URL,
				Relevance: result.Score,
				Tags:      []string{"mcp", string(source)},
				CreatedAt: time.Now(),
				Metadata: map[string]interface{}{
					"source": "mcp_response",
				},
			}
			items = append(items, item)
		}
	} else {
		// 如果JSON解析失败，将整个输出作为一个知识项
		item := model.KnowledgeItem{
			ID:        fmt.Sprintf("mcp_%s_%d", source, time.Now().Unix()),
			Source:    source,
			Title:     "MCP查询结果",
			Content:   mcpOutput,
			URL:       "",
			Relevance: 0.8, // 默认相关性分数
			Tags:      []string{"mcp", string(source)},
			CreatedAt: time.Now(),
			Metadata: map[string]interface{}{
				"source": "mcp_raw_response",
			},
		}
		items = append(items, item)
	}

	return items
}

// GetServerConfig 获取服务器配置
func (m *Manager) GetServerConfig(serverLabel string) (*model.MCPServer, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	config, exists := m.config.Servers[serverLabel]
	return &config, exists
}

// IsServerEnabled 检查服务器是否启用
func (m *Manager) IsServerEnabled(serverLabel string) bool {
	config, exists := m.GetServerConfig(serverLabel)
	return exists && config.Enabled
}

// GetEnabledServers 获取所有启用的服务器
func (m *Manager) GetEnabledServers() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var enabledServers []string
	for serverLabel, config := range m.config.Servers {
		if config.Enabled {
			enabledServers = append(enabledServers, serverLabel)
		}
	}

	return enabledServers
}

// HealthCheck 健康检查
func (m *Manager) HealthCheck(ctx context.Context) map[string]bool {
	results := make(map[string]bool)

	for serverLabel := range m.clients {
		// 尝试列出工具来检查连接
		_, err := m.ListTools(ctx, serverLabel)
		results[serverLabel] = err == nil
	}

	return results
} 