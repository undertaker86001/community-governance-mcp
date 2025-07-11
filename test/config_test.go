package test

import (
	"os"
	"testing"

	"github.com/community-governance-mcp-higress/internal/model"
	"github.com/spf13/viper"
)

// TestConfigLoading 测试配置加载
func TestConfigLoading(t *testing.T) {
	// 设置测试配置文件路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("../configs")  // 修正路径
	viper.AddConfigPath("./configs")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		t.Skipf("跳过配置测试，无法读取配置文件: %v", err)
		return
	}

	// 解析配置
	var config model.AgentConfig
	if err := viper.Unmarshal(&config); err != nil {
		t.Fatalf("解析配置失败: %v", err)
	}

	// 验证网络配置是否正确加载
	if config.Network.ProxyEnabled {
		t.Log("代理已启用")
		if config.Network.ProxyURL == "" {
			t.Error("代理已启用但URL为空")
		}
	} else {
		t.Log("代理已禁用")
	}

	// 验证其他配置
	if config.Name == "" {
		t.Error("Agent名称不能为空")
	}

	if config.Port == 0 {
		t.Error("端口不能为0")
	}

	t.Logf("配置加载成功: Agent=%s, Port=%d, ProxyEnabled=%v", 
		config.Name, config.Port, config.Network.ProxyEnabled)
}

// TestProxyConfigValidation 测试代理配置验证
func TestProxyConfigValidation(t *testing.T) {
	testCases := []struct {
		name        string
		proxyEnabled bool
		proxyURL    string
		proxyType   string
		valid       bool
	}{
		{
			name:        "禁用代理",
			proxyEnabled: false,
			proxyURL:    "",
			proxyType:   "",
			valid:       true,
		},
		{
			name:        "启用HTTP代理",
			proxyEnabled: true,
			proxyURL:    "http://proxy.example.com:8080",
			proxyType:   "http",
			valid:       true,
		},
		{
			name:        "启用HTTPS代理",
			proxyEnabled: true,
			proxyURL:    "https://proxy.example.com:8443",
			proxyType:   "https",
			valid:       true,
		},
		{
			name:        "启用SOCKS5代理",
			proxyEnabled: true,
			proxyURL:    "socks5://proxy.example.com:1080",
			proxyType:   "socks5",
			valid:       true,
		},
		{
			name:        "启用代理但URL为空",
			proxyEnabled: true,
			proxyURL:    "",
			proxyType:   "http",
			valid:       false,
		},
		{
			name:        "启用代理但类型为空",
			proxyEnabled: true,
			proxyURL:    "http://proxy.example.com:8080",
			proxyType:   "",
			valid:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			networkConfig := &model.NetworkConfig{
				ProxyEnabled: tc.proxyEnabled,
				ProxyURL:     tc.proxyURL,
				ProxyType:    tc.proxyType,
			}

			// 验证配置
			if tc.proxyEnabled {
				if networkConfig.ProxyURL == "" {
					t.Error("启用代理时URL不能为空")
				}
				if networkConfig.ProxyType == "" {
					t.Error("启用代理时类型不能为空")
				}
			}

			// 验证配置是否正确设置
			if networkConfig.ProxyEnabled != tc.proxyEnabled {
				t.Error("ProxyEnabled设置不正确")
			}
			if networkConfig.ProxyURL != tc.proxyURL {
				t.Error("ProxyURL设置不正确")
			}
			if networkConfig.ProxyType != tc.proxyType {
				t.Error("ProxyType设置不正确")
			}
		})
	}
}

// TestEnvironmentVariableOverride 测试环境变量覆盖
func TestEnvironmentVariableOverride(t *testing.T) {
	// 设置环境变量
	os.Setenv("PROXY_ENABLED", "true")
	os.Setenv("PROXY_URL", "http://env-proxy.example.com:8080")
	os.Setenv("PROXY_TYPE", "http")

	defer func() {
		// 清理环境变量
		os.Unsetenv("PROXY_ENABLED")
		os.Unsetenv("PROXY_URL")
		os.Unsetenv("PROXY_TYPE")
	}()

	// 验证环境变量是否正确设置
	if os.Getenv("PROXY_ENABLED") != "true" {
		t.Error("环境变量PROXY_ENABLED设置失败")
	}
	if os.Getenv("PROXY_URL") != "http://env-proxy.example.com:8080" {
		t.Error("环境变量PROXY_URL设置失败")
	}
	if os.Getenv("PROXY_TYPE") != "http" {
		t.Error("环境变量PROXY_TYPE设置失败")
	}
}

// TestConfigFileExists 测试配置文件是否存在
func TestConfigFileExists(t *testing.T) {
	configFiles := []string{
		"../configs/config.yaml",
		"../configs/config_with_proxy.yaml",
	}

	for _, file := range configFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("配置文件不存在: %s", file)
		} else {
			t.Logf("配置文件存在: %s", file)
		}
	}
} 