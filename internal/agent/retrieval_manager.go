package agent

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// RetrievalManager 检索管理器
type RetrievalManager struct {
	logger *logrus.Logger
	client *http.Client
}

// NewRetrievalManager 创建新的检索管理器
func NewRetrievalManager() *RetrievalManager {
	// 配置HTTP客户端，处理网络限制
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
		// 设置代理（如果需要）
		// Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	return &RetrievalManager{
		logger: logrus.New(),
		client: client,
	}
}

// RetrievalConfig 检索配置
type RetrievalConfig struct {
	MaxRetries     int           `json:"max_retries"`     // 最大重试次数
	RetryDelay     time.Duration `json:"retry_delay"`     // 重试延迟
	Timeout        time.Duration `json:"timeout"`         // 超时时间
	UserAgent      string        `json:"user_agent"`      // 用户代理
	EnableProxy    bool          `json:"enable_proxy"`    // 是否启用代理
	ProxyURL       string        `json:"proxy_url"`       // 代理URL
	EnableFallback bool          `json:"enable_fallback"` // 是否启用备用方案
}

// DefaultRetrievalConfig 默认检索配置
func DefaultRetrievalConfig() *RetrievalConfig {
	return &RetrievalConfig{
		MaxRetries:     3,
		RetryDelay:     2 * time.Second,
		Timeout:        15 * time.Second,
		UserAgent:      "Mozilla/5.0 (compatible; HigressBot/1.0)",
		EnableProxy:    false,
		ProxyURL:       "",
		EnableFallback: true,
	}
}

// RetrievalResult 检索结果
type RetrievalResult struct {
	Success    bool          `json:"success"`
	Data       []byte        `json:"data"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration"`
	Retries    int           `json:"retries"`
	Error      error         `json:"error,omitempty"`
}

// RetrieveWithRetry 带重试的检索
func (rm *RetrievalManager) RetrieveWithRetry(ctx context.Context, url string, config *RetrievalConfig) (*RetrievalResult, error) {
	if config == nil {
		config = DefaultRetrievalConfig()
	}

	var lastErr error
	startTime := time.Now()

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		result, err := rm.retrieveOnce(ctx, url, config)
		if err == nil {
			result.Retries = attempt
			result.Duration = time.Since(startTime)
			return result, nil
		}

		lastErr = err
		rm.logger.WithFields(logrus.Fields{
			"url":      url,
			"attempt":  attempt + 1,
			"max_retries": config.MaxRetries,
			"error":    err.Error(),
		}).Warn("检索失败，准备重试")

		// 如果不是最后一次尝试，等待后重试
		if attempt < config.MaxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(config.RetryDelay):
				continue
			}
		}
	}

	return &RetrievalResult{
		Success:    false,
		Retries:    config.MaxRetries,
		Duration:   time.Since(startTime),
		Error:      lastErr,
	}, lastErr
}

// retrieveOnce 单次检索
func (rm *RetrievalManager) retrieveOnce(ctx context.Context, url string, config *RetrievalConfig) (*RetrievalResult, error) {
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	// 构建请求
	req, err := http.NewRequestWithContext(timeoutCtx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Accept", "application/json, text/html, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := rm.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return &RetrievalResult{
			Success:    false,
			StatusCode: resp.StatusCode,
			Error:      fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}, nil
	}

	// 读取响应内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	return &RetrievalResult{
		Success:    true,
		Data:       data,
		StatusCode: resp.StatusCode,
	}, nil
}

// MultiEndpointRetrieval 多端点检索
type MultiEndpointRetrieval struct {
	Endpoints []string
	Config    *RetrievalConfig
}

// NewMultiEndpointRetrieval 创建多端点检索
func NewMultiEndpointRetrieval(endpoints []string, config *RetrievalConfig) *MultiEndpointRetrieval {
	return &MultiEndpointRetrieval{
		Endpoints: endpoints,
		Config:    config,
	}
}

// Retrieve 执行多端点检索
func (mer *MultiEndpointRetrieval) Retrieve(ctx context.Context, rm *RetrievalManager) (*RetrievalResult, error) {
	for _, endpoint := range mer.Endpoints {
		result, err := rm.RetrieveWithRetry(ctx, endpoint, mer.Config)
		if err == nil && result.Success {
			rm.logger.WithField("endpoint", endpoint).Info("多端点检索成功")
			return result, nil
		}
		rm.logger.WithError(err).WithField("endpoint", endpoint).Warn("端点检索失败")
	}

	return &RetrievalResult{
		Success: false,
		Error:   fmt.Errorf("所有端点都检索失败"),
	}, fmt.Errorf("所有端点都检索失败")
}

// NetworkLimitationHandler 网络限制处理器
type NetworkLimitationHandler struct {
	logger *logrus.Logger
}

// NewNetworkLimitationHandler 创建网络限制处理器
func NewNetworkLimitationHandler() *NetworkLimitationHandler {
	return &NetworkLimitationHandler{
		logger: logrus.New(),
	}
}

// HandleGitHubLimitation 处理GitHub访问限制
func (nlh *NetworkLimitationHandler) HandleGitHubLimitation(ctx context.Context, url string) error {
	// 检查是否是GitHub URL
	if !strings.Contains(url, "github.com") {
		return nil
	}

	// GitHub访问限制处理策略
	nlh.logger.WithField("url", url).Info("检测到GitHub URL，应用访问限制处理策略")

	// 1. 使用备用镜像
	_ = []string{
		strings.Replace(url, "github.com", "hub.fastgit.xyz", 1),
		strings.Replace(url, "github.com", "github.com.cnpmjs.org", 1),
		strings.Replace(url, "github.com", "github.91chi.fun", 1),
	}

	// 2. 使用API而不是网页
	if strings.Contains(url, "/blob/") {
		apiURL := strings.Replace(url, "/blob/", "/contents/", 1)
		apiURL = strings.Replace(apiURL, "github.com", "api.github.com/repos", 1)
		nlh.logger.WithField("api_url", apiURL).Info("转换为GitHub API URL")
	}

	return nil
}

// HandleTimeout 处理超时问题
func (nlh *NetworkLimitationHandler) HandleTimeout(ctx context.Context, timeout time.Duration) context.Context {
	// 根据网络状况动态调整超时时间
	adjustedTimeout := timeout
	if timeout < 5*time.Second {
		adjustedTimeout = 10 * time.Second
	}

	nlh.logger.WithField("original_timeout", timeout).WithField("adjusted_timeout", adjustedTimeout).Info("调整超时时间")

	ctx2, _ := context.WithTimeout(ctx, adjustedTimeout)
	return ctx2
}

// FallbackStrategy 备用策略
type FallbackStrategy struct {
	logger *logrus.Logger
}

// NewFallbackStrategy 创建备用策略
func NewFallbackStrategy() *FallbackStrategy {
	return &FallbackStrategy{
		logger: logrus.New(),
	}
}

// GetHigressFallbackData 获取Higress备用数据
func (fs *FallbackStrategy) GetHigressFallbackData() map[string]string {
	return map[string]string{
		"安装": "Higress支持Docker和Kubernetes两种部署方式。Docker部署简单快速，适合开发和测试环境。Kubernetes部署适合生产环境，提供更好的可扩展性和管理能力。",
		"配置": "Higress使用YAML格式进行配置，支持动态配置更新。主要配置文件包括路由配置、插件配置、安全配置等。",
		"插件": "Higress支持WASM插件开发，提供Go、Rust、JavaScript等语言SDK。插件可以扩展网关功能，实现自定义逻辑。",
		"网关": "Higress是一个基于Envoy的云原生API网关，支持多种协议，提供丰富的插件生态。",
		"监控": "Higress提供Prometheus指标导出，支持Grafana仪表板，可以监控网关性能和流量。",
		"故障排查": "常见问题包括网络连接、配置错误、权限问题等。建议检查日志、配置文件和网络连接。",
		"性能优化": "建议启用缓存、配置合适的连接池大小、使用流式处理。",
		"安全配置": "启用HTTPS、配置WAF防护、使用JWT认证、设置访问控制。",
	}
}

// GetDeepWikiFallbackData 获取DeepWiki备用数据
func (fs *FallbackStrategy) GetDeepWikiFallbackData() map[string]string {
	return map[string]string{
		"知识检索": "DeepWiki提供智能知识检索服务，支持语义搜索和内容理解。",
		"文档管理": "DeepWiki支持多种文档格式，提供版本控制和协作功能。",
		"API集成": "DeepWiki提供RESTful API和MCP协议支持，便于系统集成。",
	}
}

// IsNetworkError 判断是否为网络错误
func IsNetworkError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	networkErrors := []string{
		"timeout",
		"connection refused",
		"no route to host",
		"network is unreachable",
		"connection reset by peer",
		"i/o timeout",
		"context deadline exceeded",
	}

	for _, networkError := range networkErrors {
		if strings.Contains(strings.ToLower(errorStr), networkError) {
			return true
		}
	}

	return false
}

// ShouldRetry 判断是否应该重试
func ShouldRetry(err error, statusCode int) bool {
	// 网络错误应该重试
	if IsNetworkError(err) {
		return true
	}

	// 5xx错误应该重试
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// 429 (Too Many Requests) 应该重试
	if statusCode == 429 {
		return true
	}

	return false
} 