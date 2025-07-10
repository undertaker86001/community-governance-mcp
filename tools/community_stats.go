package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"community-governance-mcp-higress/internal/agent"
)

// CommunityStats 社区统计工具
type CommunityStats struct {
	config *agent.AgentConfig
}

// NewCommunityStats 创建新的社区统计工具
func NewCommunityStats() *CommunityStats {
	return &CommunityStats{}
}

// SetConfig 设置配置
func (c *CommunityStats) SetConfig(config *agent.AgentConfig) {
	c.config = config
}

// GetStats 获取社区统计
func (c *CommunityStats) GetStats(ctx context.Context) (*agent.CommunityStats, error) {
	if c.config == nil || c.config.GitHub.Token == "" {
		return nil, fmt.Errorf("GitHub token 未配置")
	}

	// 获取Issues统计
	issuesStats, err := c.getIssuesStats(ctx)
	if err != nil {
		return nil, err
	}

	// 获取PR统计
	prStats, err := c.getPRStats(ctx)
	if err != nil {
		return nil, err
	}

	// 获取贡献者统计
	contributorStats, err := c.getContributorStats(ctx)
	if err != nil {
		return nil, err
	}

	// 构建完整统计
	stats := &agent.CommunityStats{
		TotalIssues:     issuesStats.Total,
		OpenIssues:      issuesStats.Open,
		ClosedIssues:    issuesStats.Closed,
		TotalPRs:        prStats.Total,
		OpenPRs:         prStats.Open,
		MergedPRs:       prStats.Merged,
		Contributors:    contributorStats.Total,
		ActiveUsers:     contributorStats.Active,
		TopContributors: contributorStats.Top,
		IssueTrends:     issuesStats.Trends,
		PRTrends:        prStats.Trends,
		GeneratedAt:     time.Now(),
	}

	return stats, nil
}

// IssuesStats Issues统计
type IssuesStats struct {
	Total   int                    `json:"total"`
	Open    int                    `json:"open"`
	Closed  int                    `json:"closed"`
	Trends  []agent.IssueTrend    `json:"trends"`
}

// PRStats PR统计
type PRStats struct {
	Total   int                `json:"total"`
	Open    int                `json:"open"`
	Merged  int                `json:"merged"`
	Trends  []agent.PRTrend    `json:"trends"`
}

// ContributorStats 贡献者统计
type ContributorStats struct {
	Total int                    `json:"total"`
	Active int                   `json:"active"`
	Top    []agent.Contributor   `json:"top"`
}

// getIssuesStats 获取Issues统计
func (c *CommunityStats) getIssuesStats(ctx context.Context) (*IssuesStats, error) {
	issuesURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all&per_page=100",
		c.config.GitHub.Owner, c.config.GitHub.Repo)

	req, err := http.NewRequestWithContext(ctx, "GET", issuesURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issues []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	stats := &IssuesStats{}
	for _, issue := range issues {
		stats.Total++
		state := issue["state"].(string)
		if state == "open" {
			stats.Open++
		} else {
			stats.Closed++
		}
	}

	// 生成趋势数据（简化版本）
	stats.Trends = c.generateIssueTrends()

	return stats, nil
}

// getPRStats 获取PR统计
func (c *CommunityStats) getPRStats(ctx context.Context) (*PRStats, error) {
	prsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all&per_page=100",
		c.config.GitHub.Owner, c.config.GitHub.Repo)

	req, err := http.NewRequestWithContext(ctx, "GET", prsURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var prs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	stats := &PRStats{}
	for _, pr := range prs {
		stats.Total++
		state := pr["state"].(string)
		if state == "open" {
			stats.Open++
		} else if state == "closed" {
			stats.Merged++
		}
	}

	// 生成趋势数据（简化版本）
	stats.Trends = c.generatePRTrends()

	return stats, nil
}

// getContributorStats 获取贡献者统计
func (c *CommunityStats) getContributorStats(ctx context.Context) (*ContributorStats, error) {
	contributorsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors",
		c.config.GitHub.Owner, c.config.GitHub.Repo)

	req, err := http.NewRequestWithContext(ctx, "GET", contributorsURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.config.GitHub.Token)
	req.Header.Set("Accept", "application/vnd.github+json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var contributors []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&contributors); err != nil {
		return nil, err
	}

	stats := &ContributorStats{
		Total: len(contributors),
	}

	// 计算活跃用户（贡献次数大于5的用户）
	activeCount := 0
	for _, contributor := range contributors {
		contributions := int(contributor["contributions"].(float64))
		if contributions > 5 {
			activeCount++
		}
	}
	stats.Active = activeCount

	// 获取前10名贡献者
	stats.Top = c.getTopContributors(contributors)

	return stats, nil
}

// generateIssueTrends 生成Issue趋势
func (c *CommunityStats) generateIssueTrends() []agent.IssueTrend {
	// 简化版本，实际应该从GitHub API获取历史数据
	trends := make([]agent.IssueTrend, 7)
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		trends[6-i] = agent.IssueTrend{
			Date:   date.Format("2006-01-02"),
			Opened: 5 + i,  // 模拟数据
			Closed: 3 + i,
		}
	}
	return trends
}

// generatePRTrends 生成PR趋势
func (c *CommunityStats) generatePRTrends() []agent.PRTrend {
	// 简化版本，实际应该从GitHub API获取历史数据
	trends := make([]agent.PRTrend, 7)
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		trends[6-i] = agent.PRTrend{
			Date:   date.Format("2006-01-02"),
			Opened: 2 + i,  // 模拟数据
			Merged: 1 + i,
		}
	}
	return trends
}

// getTopContributors 获取前10名贡献者
func (c *CommunityStats) getTopContributors(contributors []map[string]interface{}) []agent.Contributor {
	topContributors := make([]agent.Contributor, 0, 10)
	
	for i, contributor := range contributors {
		if i >= 10 {
			break
		}

		login := contributor["login"].(string)
		contributions := int(contributor["contributions"].(float64))
		avatarURL := contributor["avatar_url"].(string)

		topContributors = append(topContributors, agent.Contributor{
			Username:     login,
			AvatarURL:    avatarURL,
			Contributions: contributions,
			Issues:       0, // 需要额外API调用获取
			PRs:          0, // 需要额外API调用获取
		})
	}

	return topContributors
}
