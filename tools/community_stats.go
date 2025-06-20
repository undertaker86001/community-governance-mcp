package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"community-governance-mcp/config"
	mcp_tool "community-governance-mcp/utils"
	"github.com/higress-group/wasm-go/pkg/mcp/server"
	"github.com/higress-group/wasm-go/pkg/mcp/utils"
)

type CommunityStats struct {
	StatsType    string `json:"stats_type" jsonschema_description:"统计类型：issues, contributors, activity, all" jsonschema:"example=all"`
	TimeRange    string `json:"time_range,omitempty" jsonschema_description:"时间范围：week, month, quarter, year" jsonschema:"example=month"`
	OutputFormat string `json:"output_format,omitempty" jsonschema_description:"输出格式：text, json" jsonschema:"example=text"`
}

func (t CommunityStats) Description() string {
	return `社区统计工具，生成Issues统计、贡献者分析、活跃度报告等社区治理数据，帮助了解项目健康状况。`
}

func (t CommunityStats) InputSchema() map[string]any {
	return server.ToInputSchema(&CommunityStats{})
}

func (t CommunityStats) Create(params []byte) server.Tool {
	stats := &CommunityStats{
		StatsType:    "all",
		TimeRange:    "month",
		OutputFormat: "text",
	}
	json.Unmarshal(params, stats)
	return stats
}

func (t CommunityStats) Call(ctx server.HttpContext, s server.Server) error {
	serverConfig := &config.CommunityGovernanceConfig{}
	s.GetConfig(serverConfig)

	if serverConfig.GitHubToken == "" {
		return fmt.Errorf("GitHub token 未配置")
	}

	var report string
	var err error

	switch t.StatsType {
	case "issues":
		report, err = t.generateIssueStats(ctx, serverConfig)
	case "contributors":
		report, err = t.generateContributorStats(ctx, serverConfig)
	case "activity":
		report, err = t.generateActivityStats(ctx, serverConfig)
	default:
		report, err = t.generateFullReport(ctx, serverConfig)
	}

	if err != nil {
		return err
	}

	utils.SendMCPToolTextResult(ctx, report)
	return nil
}

func (t CommunityStats) generateIssueStats(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	// 获取Issues统计数据
	issuesURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all&per_page=100",
		config.RepoOwner, config.RepoName)

	headers := map[string]string{
		"Authorization": "Bearer " + config.GitHubToken,
		"Accept":        "application/vnd.github+json",
	}

	response, err := mcp_tool.SendHTTPRequest(ctx, "GET", issuesURL, headers, "")
	if err != nil {
		return "", err
	}

	var issues []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &issues); err != nil {
		return "", err
	}

	// 统计分析
	openCount := 0
	closedCount := 0
	bugCount := 0
	featureCount := 0

	for _, issue := range issues {
		state := issue["state"].(string)
		if state == "open" {
			openCount++
		} else {
			closedCount++
		}

		// 分析标签
		if labels, ok := issue["labels"].([]interface{}); ok {
			for _, label := range labels {
				labelMap := label.(map[string]interface{})
				labelName := labelMap["name"].(string)
				if labelName == "bug" {
					bugCount++
				} else if labelName == "enhancement" || labelName == "feature" {
					featureCount++
				}
			}
		}
	}

	report := fmt.Sprintf(`# Issues 统计报告  
  
## 总体概况  
- 总Issues数量: %d  
- 开放Issues: %d  
- 已关闭Issues: %d  
- 关闭率: %.1f%%  
  
## 类型分布  
- Bug报告: %d  
- 功能请求: %d  
- 其他: %d  
  
## 健康度评估  
%s`,
		len(issues), openCount, closedCount,
		float64(closedCount)/float64(len(issues))*100,
		bugCount, featureCount, len(issues)-bugCount-featureCount,
		t.assessProjectHealth(openCount, closedCount, bugCount))

	return report, nil
}

func (t CommunityStats) generateContributorStats(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	contributorsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors",
		config.RepoOwner, config.RepoName)

	headers := map[string]string{
		"Authorization": "Bearer " + config.GitHubToken,
		"Accept":        "application/vnd.github+json",
	}

	response, err := mcp_tool.SendHTTPRequest(ctx, "GET", contributorsURL, headers, "")
	if err != nil {
		return "", err
	}

	var contributors []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &contributors); err != nil {
		return "", err
	}

	report := fmt.Sprintf(`# 贡献者统计报告  
  
## 总体数据  
- 总贡献者数量: %d  
- 核心贡献者(>10次提交): %d  
  
## 前10名贡献者  
`, len(contributors), t.countCoreContributors(contributors))

	for i, contributor := range contributors {
		if i >= 10 {
			break
		}

		login := contributor["login"].(string)
		contributions := int(contributor["contributions"].(float64))

		report += fmt.Sprintf("%d. %s - %d次贡献\n", i+1, login, contributions)
	}

	return report, nil
}

func (t CommunityStats) generateActivityStats(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	// 获取最近的活动数据
	eventsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/events",
		config.RepoOwner, config.RepoName)

	headers := map[string]string{
		"Authorization": "Bearer " + config.GitHubToken,
		"Accept":        "application/vnd.github+json",
	}

	response, err := mcp_tool.SendHTTPRequest(ctx, "GET", eventsURL, headers, "")
	if err != nil {
		return "", err
	}

	var events []map[string]interface{}
	if err := json.Unmarshal([]byte(response), &events); err != nil {
		return "", err
	}

	// 统计活动类型
	activityStats := make(map[string]int)
	for _, event := range events {
		eventType := event["type"].(string)
		activityStats[eventType]++
	}

	report := fmt.Sprintf(`# 活跃度统计报告  
  
## 最近活动概况  
- 总活动数量: %d  
- 活动类型分布:  
`, len(events))

	for eventType, count := range activityStats {
		report += fmt.Sprintf("  - %s: %d次\n", eventType, count)
	}

	return report, nil
}

func (t CommunityStats) generateFullReport(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	issueStats, _ := t.generateIssueStats(ctx, config)
	contributorStats, _ := t.generateContributorStats(ctx, config)
	activityStats, _ := t.generateActivityStats(ctx, config)

	fullReport := fmt.Sprintf(`# 社区治理完整报告  
  
生成时间: %s  
  
%s  
  
---  
  
%s  
  
---  
  
%s  
  
## 总结建议  
- 保持活跃的社区互动  
- 及时处理开放的Issues  
- 鼓励新贡献者参与  
- 定期更新文档和最佳实践  
`, time.Now().Format("2006-01-02 15:04:05"), issueStats, contributorStats, activityStats)

	return fullReport, nil
}

func (t CommunityStats) assessProjectHealth(openCount, closedCount, bugCount int) string {
	total := openCount + closedCount
	if total == 0 {
		return "项目刚起步，需要更多社区成员参与"
	}

	closeRate := float64(closedCount) / float64(total)
	bugRate := float64(bugCount) / float64(total)

	var health string
	if closeRate > 0.8 && bugRate < 0.3 {
		health = "项目健康状况良好"
	} else if closeRate > 0.6 && bugRate < 0.5 {
		health = "项目健康状况一般，需要关注Bug处理"
	} else {
		health = "项目需要加强维护，建议优先处理积压的Issues"
	}

	return health
}

func (t CommunityStats) countCoreContributors(contributors []map[string]interface{}) int {
	count := 0
	for _, contributor := range contributors {
		contributions := int(contributor["contributions"].(float64))
		if contributions > 10 {
			count++
		}
	}
	return count
}
