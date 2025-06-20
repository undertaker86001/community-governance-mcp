package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"community-governance-mcp/config"
	mcp_tool "community-governance-mcp/utils"
	"github.com/higress-group/wasm-go/pkg/mcp/server"
	"github.com/higress-group/wasm-go/pkg/mcp/utils"
)

type KnowledgeBase struct {
	Query      string `json:"query" jsonschema_description:"搜索查询内容" jsonschema:"example=如何配置Higress网关"`
	SearchType string `json:"search_type" jsonschema_description:"搜索类型：docs, issues, best_practices, all" jsonschema:"example=docs"`
	Language   string `json:"language,omitempty" jsonschema_description:"返回结果的语言" jsonschema:"example=zh"`
}

func (t KnowledgeBase) Description() string {
	return `知识库搜索工具，支持搜索Higress官方文档、历史Issues、最佳实践等内容，为用户提供准确的技术支持。`
}

func (t KnowledgeBase) InputSchema() map[string]any {
	return server.ToInputSchema(&KnowledgeBase{})
}

func (t KnowledgeBase) Create(params []byte) server.Tool {
	kb := &KnowledgeBase{
		SearchType: "all",
		Language:   "zh",
	}
	json.Unmarshal(params, kb)
	return kb
}

func (t KnowledgeBase) Call(ctx server.HttpContext, s server.Server) error {
	serverConfig := &config.CommunityGovernanceConfig{}
	s.GetConfig(serverConfig)

	var results []string

	// 根据搜索类型执行不同的搜索策略
	switch t.SearchType {
	case "docs":
		docResults := t.searchDocumentation()
		results = append(results, docResults...)
	case "issues":
		issueResults, err := t.searchIssues(ctx, serverConfig)
		if err == nil {
			results = append(results, issueResults...)
		}
	case "best_practices":
		practiceResults := t.searchBestPractices()
		results = append(results, practiceResults...)
	default:
		// 搜索所有类型
		docResults := t.searchDocumentation()
		results = append(results, docResults...)

		issueResults, err := t.searchIssues(ctx, serverConfig)
		if err == nil {
			results = append(results, issueResults...)
		}

		practiceResults := t.searchBestPractices()
		results = append(results, practiceResults...)
	}

	// 格式化搜索结果
	formattedResults := t.formatSearchResults(results)

	utils.SendMCPToolTextResult(ctx, formattedResults)
	return nil
}

func (t KnowledgeBase) searchDocumentation() []string {
	var results []string

	query := strings.ToLower(t.Query)

	// 基于关键词匹配的文档搜索
	docSections := map[string]string{
		"安装部署":   "Higress支持Docker和Kubernetes两种部署方式，详见官方文档安装指南。",
		"配置管理":   "Higress使用YAML格式进行配置，支持动态配置更新。",
		"插件开发":   "Higress支持WASM插件开发，提供Go、Rust、JavaScript等语言SDK。",
		"AI网关":   "Higress AI网关支持多种LLM提供商，包括OpenAI、Claude、Qwen等。",
		"MCP服务器": "Higress支持托管MCP服务器，为AI Agent提供工具调用能力。",
		"故障排查":   "常见问题包括网络连接、配置错误、权限问题等。",
	}

	for section, content := range docSections {
		if strings.Contains(query, strings.ToLower(section)) {
			results = append(results, fmt.Sprintf("**%s**: %s", section, content))
		}
	}

	return results
}

func (t KnowledgeBase) searchIssues(ctx server.HttpContext, config *config.CommunityGovernanceConfig) ([]string, error) {
	if config.GitHubToken == "" {
		return []string{"GitHub token未配置，无法搜索Issues"}, nil
	}

	// 构建GitHub搜索API请求
	searchURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s+repo:%s/%s",
		t.Query, config.RepoOwner, config.RepoName)

	headers := map[string]string{
		"Authorization": "Bearer " + config.GitHubToken,
		"Accept":        "application/vnd.github+json",
	}

	response, err := mcp_tool.SendHTTPRequest(ctx, "GET", searchURL, headers, "")
	if err != nil {
		return nil, err
	}

	// 解析搜索结果
	var searchResult map[string]interface{}
	if err := json.Unmarshal([]byte(response), &searchResult); err != nil {
		return nil, err
	}

	var results []string
	items, ok := searchResult["items"].([]interface{})
	if !ok {
		return results, nil
	}

	for i, item := range items {
		if i >= 5 { // 限制返回前5个结果
			break
		}

		issue := item.(map[string]interface{})
		title := issue["title"].(string)
		url := issue["html_url"].(string)
		state := issue["state"].(string)

		results = append(results, fmt.Sprintf("- [%s](%s) - 状态: %s", title, url, state))
	}

	return results, nil
}

func (t KnowledgeBase) searchBestPractices() []string {
	var results []string

	query := strings.ToLower(t.Query)

	// 最佳实践知识库
	practices := map[string]string{
		"性能优化": "建议启用缓存、配置合适的连接池大小、使用流式处理。",
		"安全配置": "启用HTTPS、配置WAF防护、使用JWT认证、设置访问控制。",
		"监控告警": "配置Prometheus指标、设置健康检查、启用访问日志。",
		"高可用":  "部署多实例、配置负载均衡、设置故障转移。",
		"插件开发": "遵循WASM插件开发规范、进行充分测试、注意内存管理。",
	}

	for practice, content := range practices {
		if strings.Contains(query, strings.ToLower(practice)) {
			results = append(results, fmt.Sprintf("**%s最佳实践**: %s", practice, content))
		}
	}

	return results
}

func (t KnowledgeBase) formatSearchResults(results []string) string {
	if len(results) == 0 {
		return fmt.Sprintf("未找到与 \"%s\" 相关的内容。建议：\n1. 检查关键词拼写\n2. 尝试使用更通用的关键词\n3. 查看官方文档", t.Query)
	}

	var formatted strings.Builder
	formatted.WriteString(fmt.Sprintf("# 知识库搜索结果\n\n搜索关键词：%s\n\n", t.Query))

	for i, result := range results {
		formatted.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, result))
	}

	formatted.WriteString("---\n\n如需更多帮助，请访问 [Higress官方文档](https://higress.cn/docs/)")

	return formatted.String()
}
