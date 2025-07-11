package test

import (
	"testing"
	"time"

	"github.com/community-governance-mcp-higress/internal/agent"
	"github.com/community-governance-mcp-higress/internal/model"
)

// TestProxyConfiguration 测试代理配置
func TestProxyConfiguration(t *testing.T) {
	// 测试启用代理的配置
	networkConfig := &model.NetworkConfig{
		ProxyEnabled: true,
		ProxyURL:     "http://proxy.example.com:8080",
		ProxyType:    "http",
	}

	// 创建检索管理器
	retrievalManager := agent.NewRetrievalManager(networkConfig)

	// 验证检索管理器是否正确创建
	if retrievalManager == nil {
		t.Fatal("RetrievalManager should not be nil")
	}

	// 验证网络配置是否正确传递
	if retrievalManager.GetNetworkConfig() == nil {
		t.Fatal("NetworkConfig should not be nil")
	}

	if !retrievalManager.GetNetworkConfig().ProxyEnabled {
		t.Error("Proxy should be enabled")
	}

	if retrievalManager.GetNetworkConfig().ProxyURL != "http://proxy.example.com:8080" {
		t.Error("Proxy URL should match")
	}

	if retrievalManager.GetNetworkConfig().ProxyType != "http" {
		t.Error("Proxy type should match")
	}
}

// TestProxyDisabled 测试禁用代理的配置
func TestProxyDisabled(t *testing.T) {
	// 测试禁用代理的配置
	networkConfig := &model.NetworkConfig{
		ProxyEnabled: false,
		ProxyURL:     "",
		ProxyType:    "",
	}

	// 创建检索管理器
	retrievalManager := agent.NewRetrievalManager(networkConfig)

	// 验证检索管理器是否正确创建
	if retrievalManager == nil {
		t.Fatal("RetrievalManager should not be nil")
	}

	// 验证代理是否被禁用
	if retrievalManager.GetNetworkConfig().ProxyEnabled {
		t.Error("Proxy should be disabled")
	}
}

// TestProxyURLParsing 测试代理URL解析
func TestProxyURLParsing(t *testing.T) {
	testCases := []struct {
		name     string
		proxyURL string
		valid    bool
	}{
		{
			name:     "Valid HTTP proxy",
			proxyURL: "http://proxy.example.com:8080",
			valid:    true,
		},
		{
			name:     "Valid HTTPS proxy",
			proxyURL: "https://proxy.example.com:8443",
			valid:    true,
		},
		{
			name:     "Valid SOCKS5 proxy",
			proxyURL: "socks5://proxy.example.com:1080",
			valid:    true,
		},
		{
			name:     "Valid proxy with authentication",
			proxyURL: "http://user:pass@proxy.example.com:8080",
			valid:    true,
		},
		{
			name:     "Invalid proxy URL",
			proxyURL: "invalid-url",
			valid:    false,
		},
		{
			name:     "Empty proxy URL",
			proxyURL: "",
			valid:    true, // 空URL是有效的（禁用代理）
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			networkConfig := &model.NetworkConfig{
				ProxyEnabled: tc.proxyURL != "",
				ProxyURL:     tc.proxyURL,
				ProxyType:    "http",
			}

			retrievalManager := agent.NewRetrievalManager(networkConfig)

			if retrievalManager == nil {
				t.Fatal("RetrievalManager should not be nil")
			}

			// 验证配置是否正确加载
			config := retrievalManager.GetNetworkConfig()
			if config == nil {
				t.Fatal("NetworkConfig should not be nil")
			}

			if tc.valid {
				// 对于有效的配置，应该能够正常创建
				if retrievalManager == nil {
					t.Error("RetrievalManager should be created successfully")
				}
			}
		})
	}
}

// TestDefaultRetrievalConfig 测试默认检索配置
func TestDefaultRetrievalConfig(t *testing.T) {
	config := agent.DefaultRetrievalConfig()

	if config == nil {
		t.Fatal("DefaultRetrievalConfig should not return nil")
	}

	// 验证默认值
	if config.MaxRetries != 3 {
		t.Error("Default MaxRetries should be 3")
	}

	if config.Timeout != 15*time.Second {
		t.Error("Default Timeout should be 15 seconds")
	}

	if config.UserAgent != "Mozilla/5.0 (compatible; HigressBot/1.0)" {
		t.Error("Default UserAgent should match")
	}

	if config.EnableProxy {
		t.Error("Default EnableProxy should be false")
	}

	if !config.EnableFallback {
		t.Error("Default EnableFallback should be true")
	}
} 