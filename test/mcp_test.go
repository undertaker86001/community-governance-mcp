package test

import (
	"context"
	"testing"
	"time"

	"github.com/community-governance-mcp-higress/internal/mcp"
)

// TestMCPClient 测试MCP客户端
func TestMCPClient(t *testing.T) {
	client := mcp.NewClient(30 * time.Second)

	// 测试列出工具
	t.Run("ListTools", func(t *testing.T) {
		req := &mcp.ListToolsRequest{
			ServerLabel: "deepwiki",
			ServerURL:   "https://mcp.deepwiki.com/mcp",
		}

		resp, err := client.ListTools(context.Background(), req)
		if err != nil {
			t.Skipf("跳过MCP测试，无法连接到服务器: %v", err)
			return
		}

		if len(resp.Tools) == 0 {
			t.Error("工具列表为空")
		}

		t.Logf("获取到 %d 个工具", len(resp.Tools))
		for _, tool := range resp.Tools {
			t.Logf("工具: %s - %s", tool.Name, tool.Description)
		}
	})

	// 测试查询功能
	t.Run("Query", func(t *testing.T) {
		req := &mcp.QueryRequest{
			ServerLabel: "deepwiki",
			Input:       "What is MCP?",
			RepoName:    "modelcontextprotocol/modelcontextprotocol",
		}

		resp, err := client.Query(context.Background(), req)
		if err != nil {
			t.Skipf("跳过MCP查询测试，无法连接到服务器: %v", err)
			return
		}

		if resp.Error != "" {
			t.Errorf("查询失败: %s", resp.Error)
		}

		if resp.Output == "" {
			t.Error("查询结果为空")
		}

		t.Logf("查询结果: %s", resp.Output)
	})

	// 测试工具调用
	t.Run("CallTool", func(t *testing.T) {
		req := &mcp.CallToolRequest{
			ServerLabel: "deepwiki",
			ServerURL:   "https://mcp.deepwiki.com/mcp",
			ToolName:    "ask_question",
			Arguments: map[string]interface{}{
				"repoName": "modelcontextprotocol/modelcontextprotocol",
				"question": "What is MCP?",
			},
		}

		resp, err := client.CallTool(context.Background(), req)
		if err != nil {
			t.Skipf("跳过MCP工具调用测试，无法连接到服务器: %v", err)
			return
		}

		if resp.Error != "" {
			t.Errorf("工具调用失败: %s", resp.Error)
		}

		if resp.Output == "" {
			t.Error("工具调用结果为空")
		}

		t.Logf("工具调用结果: %s", resp.Output)
	})
}

// TestMCPServerURL 测试服务器URL获取
func TestMCPServerURL(t *testing.T) {
	testCases := []struct {
		serverLabel string
		expectedURL string
	}{
		{"deepwiki", "https://mcp.deepwiki.com/mcp"},
		{"stripe", "https://mcp.stripe.com"},
		{"shopify", "https://mcp.shopify.com"},
		{"unknown", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.serverLabel, func(t *testing.T) {
			url := mcp.GetServerURL(tc.serverLabel)
			if url != tc.expectedURL {
				t.Errorf("期望URL: %s, 实际URL: %s", tc.expectedURL, url)
			}
		})
	}
}

// TestMCPConfig 测试MCP配置
func TestMCPConfig(t *testing.T) {
	config := &mcp.Config{
		Enabled: "true",
		Timeout: "30s",
		Servers: map[string]mcp.ServerConfig{
			"deepwiki": {
				Enabled:         true,
				ServerURL:       "https://mcp.deepwiki.com/mcp",
				ServerLabel:     "deepwiki",
				RequireApproval: "never",
				AllowedTools:    []string{"ask_question"},
			},
		},
	}

	if !config.IsEnabled() {
		t.Error("MCP应该被启用")
	}

	server, exists := config.GetServer("deepwiki")
	if !exists {
		t.Error("应该找到deepwiki服务器配置")
	}

	if !server.Enabled {
		t.Error("deepwiki服务器应该被启用")
	}

	if server.ServerURL != "https://mcp.deepwiki.com/mcp" {
		t.Error("服务器URL不正确")
	}
} 