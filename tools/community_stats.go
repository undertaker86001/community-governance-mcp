package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/community-governance-mcp-higress/internal/model"
	"math/rand"
)

// CommunityStats 社区统计工具
type CommunityStats struct {
	githubToken string
	httpClient  *http.Client
}

// NewCommunityStats 创建新的社区统计工具
func NewCommunityStats(githubToken string) *CommunityStats {
	return &CommunityStats{
		githubToken: githubToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetCommunityStats 获取社区统计信息
func (c *CommunityStats) GetCommunityStats(owner string, repo string, period string) (*model.CommunityStats, error) {
	stats := &model.CommunityStats{
		Period:          period,
		TopContributors: []model.Contributor{},
		ActivityTrend:   []model.ActivityData{},
		Metadata:        make(map[string]interface{}),
	}

	// 获取Issue统计
	issueStats, err := c.getIssueStats(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("获取Issue统计失败: %w", err)
	}
	stats.TotalIssues = issueStats.Total
	stats.OpenIssues = issueStats.Open
	stats.ClosedIssues = issueStats.Closed

	// 获取PR统计
	prStats, err := c.getPRStats(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("获取PR统计失败: %w", err)
	}
	stats.TotalPRs = prStats.Total
	stats.OpenPRs = prStats.Open
	stats.MergedPRs = prStats.Merged

	// 获取贡献者统计
	contributors, err := c.getContributors(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("获取贡献者统计失败: %w", err)
	}
	stats.Contributors = len(contributors)
	stats.TopContributors = contributors

	// 获取活跃度趋势
	activityTrend, err := c.getActivityTrend(owner, repo, period)
	if err != nil {
		return nil, fmt.Errorf("获取活跃度趋势失败: %w", err)
	}
	stats.ActivityTrend = activityTrend

	// 计算社区健康度
	stats.HealthScore = c.calculateHealthScore(stats)

	return stats, nil
}

// IssueStats Issue统计
type IssueStats struct {
	Total  int `json:"total"`
	Open   int `json:"open"`
	Closed int `json:"closed"`
}

// PRStats PR统计
type PRStats struct {
	Total  int `json:"total"`
	Open   int `json:"open"`
	Merged int `json:"merged"`
}

// getIssueStats 获取Issue统计
func (c *CommunityStats) getIssueStats(owner string, repo string) (*IssueStats, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=all&per_page=100", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.githubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var issues []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	stats := &IssueStats{}
	for _, issue := range issues {
		stats.Total++
		if state, ok := issue["state"].(string); ok {
			if state == "open" {
				stats.Open++
			} else {
				stats.Closed++
			}
		}
	}

	return stats, nil
}

// getPRStats 获取PR统计
func (c *CommunityStats) getPRStats(owner string, repo string) (*PRStats, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=all&per_page=100", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.githubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var prs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	stats := &PRStats{}
	for _, pr := range prs {
		stats.Total++
		if state, ok := pr["state"].(string); ok {
			if state == "open" {
				stats.Open++
			} else if merged, ok := pr["merged_at"].(string); ok && merged != "" {
				stats.Merged++
			}
		}
	}

	return stats, nil
}

// getContributors 获取贡献者信息
func (c *CommunityStats) getContributors(owner string, repo string) ([]model.Contributor, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors?per_page=10", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.githubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var contributors []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&contributors); err != nil {
		return nil, err
	}

	var result []model.Contributor
	for _, contributor := range contributors {
		username, _ := contributor["login"].(string)
		avatarURL, _ := contributor["avatar_url"].(string)
		contributions, _ := contributor["contributions"].(float64)

		result = append(result, model.Contributor{
			Username:      username,
			AvatarURL:     avatarURL,
			Contributions: int(contributions),
			LastActive:    time.Now().Format("2006-01-02"),
		})
	}

	return result, nil
}

// getActivityTrend 获取活跃度趋势
func (c *CommunityStats) getActivityTrend(owner string, repo string, period string) ([]model.ActivityData, error) {
	// 这里可以添加更复杂的活跃度趋势计算
	// 目前返回模拟数据
	var trend []model.ActivityData

	// 生成最近7天的数据
	for i := 6; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		trend = append(trend, model.ActivityData{
			Date:     date.Format("2006-01-02"),
			Issues:   rand.Intn(10) + 1,
			PRs:      rand.Intn(5) + 1,
			Comments: rand.Intn(20) + 5,
		})
	}

	return trend, nil
}

// calculateHealthScore 计算社区健康度
func (c *CommunityStats) calculateHealthScore(stats *model.CommunityStats) float64 {
	// 简单的健康度计算算法
	score := 0.0

	// Issue响应率
	if stats.TotalIssues > 0 {
		responseRate := float64(stats.ClosedIssues) / float64(stats.TotalIssues)
		score += responseRate * 0.3
	}

	// PR合并率
	if stats.TotalPRs > 0 {
		mergeRate := float64(stats.MergedPRs) / float64(stats.TotalPRs)
		score += mergeRate * 0.3
	}

	// 贡献者活跃度
	if stats.Contributors > 0 {
		contributorScore := float64(stats.Contributors) / 100.0 // 假设100个贡献者为满分
		if contributorScore > 1.0 {
			contributorScore = 1.0
		}
		score += contributorScore * 0.2
	}

	// 活跃度趋势
	if len(stats.ActivityTrend) > 0 {
		recentActivity := stats.ActivityTrend[len(stats.ActivityTrend)-1]
		activityScore := float64(recentActivity.Issues+recentActivity.PRs+recentActivity.Comments) / 50.0
		if activityScore > 1.0 {
			activityScore = 1.0
		}
		score += activityScore * 0.2
	}

	return score
}

// GetRepositoryInfo 获取仓库信息
func (c *CommunityStats) GetRepositoryInfo(owner string, repo string) (map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.githubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var repoInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, err
	}

	return repoInfo, nil
}

// GetRecentActivity 获取最近活动
func (c *CommunityStats) GetRecentActivity(owner string, repo string, days int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/events?per_page=100", owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.githubToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.githubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	// 过滤最近的活动
	var recentEvents []map[string]interface{}
	cutoffTime := time.Now().AddDate(0, 0, -days)

	for _, event := range events {
		createdAt, ok := event["created_at"].(string)
		if !ok {
			continue
		}

		eventTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			continue
		}

		if eventTime.After(cutoffTime) {
			recentEvents = append(recentEvents, event)
		}
	}

	return recentEvents, nil
}
