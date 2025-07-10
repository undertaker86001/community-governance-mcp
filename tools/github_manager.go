package tools

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/community-governance-mcp-higress/internal/model"
)

// GitHubManager GitHub管理器
type GitHubManager struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewGitHubManager 创建新的GitHub管理器
func NewGitHubManager(token string) *GitHubManager {
	return &GitHubManager{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// GetIssue 获取Issue详情
func (gm *GitHubManager) GetIssue(owner string, repo string, issueNumber int) (*model.GitHubIssue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", gm.baseURL, owner, repo, issueNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var issue map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return gm.parseIssue(issue), nil
}

// GetIssues 获取Issue列表
func (gm *GitHubManager) GetIssues(owner string, repo string, state string, labels []string) ([]*model.GitHubIssue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", gm.baseURL, owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 添加查询参数
	q := req.URL.Query()
	if state != "" {
		q.Add("state", state)
	}
	if len(labels) > 0 {
		q.Add("labels", strings.Join(labels, ","))
	}
	q.Add("per_page", "100")
	req.URL.RawQuery = q.Encode()

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gm.httpClient.Do(req)
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

	var result []*model.GitHubIssue
	for _, issue := range issues {
		result = append(result, gm.parseIssue(issue))
	}

	return result, nil
}

// CreateIssue 创建Issue
func (gm *GitHubManager) CreateIssue(owner string, repo string, title string, body string, labels []string) (*model.GitHubIssue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues", gm.baseURL, owner, repo)

	requestBody := map[string]interface{}{
		"title":  title,
		"body":   body,
		"labels": labels,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := gm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("创建Issue失败: %d", resp.StatusCode)
	}

	var issue map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return gm.parseIssue(issue), nil
}

// UpdateIssue 更新Issue
func (gm *GitHubManager) UpdateIssue(owner string, repo string, issueNumber int, updates map[string]interface{}) (*model.GitHubIssue, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d", gm.baseURL, owner, repo, issueNumber)

	bodyBytes, err := json.Marshal(updates)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := gm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("更新Issue失败: %d", resp.StatusCode)
	}

	var issue map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return gm.parseIssue(issue), nil
}

// AddComment 添加评论
func (gm *GitHubManager) AddComment(owner string, repo string, issueNumber int, body string) (*model.GitHubComment, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", gm.baseURL, owner, repo, issueNumber)

	requestBody := map[string]string{
		"body": body,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := gm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("添加评论失败: %d", resp.StatusCode)
	}

	var comment map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, err
	}

	return gm.parseComment(comment), nil
}

// GetComments 获取Issue评论
func (gm *GitHubManager) GetComments(owner string, repo string, issueNumber int) ([]*model.GitHubComment, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", gm.baseURL, owner, repo, issueNumber)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var comments []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
		return nil, err
	}

	var result []*model.GitHubComment
	for _, comment := range comments {
		result = append(result, gm.parseComment(comment))
	}

	return result, nil
}

// SearchIssues 搜索Issue
func (gm *GitHubManager) SearchIssues(query string, owner string, repo string) ([]*model.GitHubIssue, error) {
	searchQuery := fmt.Sprintf("%s repo:%s/%s", query, owner, repo)
	url := fmt.Sprintf("%s/search/issues?q=%s", gm.baseURL, searchQuery)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gm.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API请求失败: %d", resp.StatusCode)
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, err
	}

	var issues []*model.GitHubIssue
	if items, ok := searchResult["items"].([]interface{}); ok {
		for _, item := range items {
			if issueMap, ok := item.(map[string]interface{}); ok {
				issues = append(issues, gm.parseIssue(issueMap))
			}
		}
	}

	return issues, nil
}

// GetRepositoryStats 获取仓库统计
func (gm *GitHubManager) GetRepositoryStats(owner string, repo string) (*model.RepositoryStats, error) {
	// 获取仓库基本信息
	repoInfo, err := gm.getRepositoryInfo(owner, repo)
	if err != nil {
		return nil, err
	}

	// 获取Issue统计
	openIssues, err := gm.GetIssues(owner, repo, "open", nil)
	if err != nil {
		return nil, err
	}

	closedIssues, err := gm.GetIssues(owner, repo, "closed", nil)
	if err != nil {
		return nil, err
	}

	stats := &model.RepositoryStats{
		Repository:   repoInfo,
		OpenIssues:   len(openIssues),
		ClosedIssues: len(closedIssues),
		TotalIssues:  len(openIssues) + len(closedIssues),
		LastUpdated:  time.Now().Format("2006-01-02 15:04:05"),
	}

	return stats, nil
}

// getRepositoryInfo 获取仓库信息
func (gm *GitHubManager) getRepositoryInfo(owner string, repo string) (*model.Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", gm.baseURL, owner, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if gm.token != "" {
		req.Header.Set("Authorization", "Bearer "+gm.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gm.httpClient.Do(req)
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

	return gm.parseRepository(repoInfo), nil
}

// parseIssue 解析Issue数据
func (gm *GitHubManager) parseIssue(data map[string]interface{}) *model.GitHubIssue {
	issue := &model.GitHubIssue{
		ID:         getInt(data, "id"),
		Number:     getInt(data, "number"),
		Title:      getString(data, "title"),
		Body:       getString(data, "body"),
		State:      getString(data, "state"),
		CreatedAt:  getString(data, "created_at"),
		UpdatedAt:  getString(data, "updated_at"),
		ClosedAt:   getString(data, "closed_at"),
		User:       gm.parseUser(getMap(data, "user")),
		Labels:     gm.parseLabels(getArray(data, "labels")),
		Assignees:  gm.parseUsers(getArray(data, "assignees")),
		Comments:   getInt(data, "comments"),
		HTMLURL:    getString(data, "html_url"),
		Repository: getString(data, "repository"),
	}

	return issue
}

// parseComment 解析评论数据
func (gm *GitHubManager) parseComment(data map[string]interface{}) *model.GitHubComment {
	comment := &model.GitHubComment{
		ID:        getInt(data, "id"),
		Body:      getString(data, "body"),
		User:      gm.parseUser(getMap(data, "user")),
		CreatedAt: getString(data, "created_at"),
		UpdatedAt: getString(data, "updated_at"),
		HTMLURL:   getString(data, "html_url"),
	}

	return comment
}

// parseRepository 解析仓库数据
func (gm *GitHubManager) parseRepository(data map[string]interface{}) *model.Repository {
	repo := &model.Repository{
		ID:          getInt(data, "id"),
		Name:        getString(data, "name"),
		FullName:    getString(data, "full_name"),
		Description: getString(data, "description"),
		Private:     getBool(data, "private"),
		Fork:        getBool(data, "fork"),
		Stars:       getInt(data, "stargazers_count"),
		Forks:       getInt(data, "forks_count"),
		Watchers:    getInt(data, "watchers_count"),
		OpenIssues:  getInt(data, "open_issues_count"),
		Language:    getString(data, "language"),
		CreatedAt:   getString(data, "created_at"),
		UpdatedAt:   getString(data, "updated_at"),
		HTMLURL:     getString(data, "html_url"),
	}

	return repo
}

// parseUser 解析用户数据
func (gm *GitHubManager) parseUser(data map[string]interface{}) *model.GitHubUser {
	if data == nil {
		return nil
	}

	user := &model.GitHubUser{
		ID:        getInt(data, "id"),
		Login:     getString(data, "login"),
		AvatarURL: getString(data, "avatar_url"),
		HTMLURL:   getString(data, "html_url"),
		Type:      getString(data, "type"),
	}

	return user
}

// parseUsers 解析用户列表
func (gm *GitHubManager) parseUsers(data []interface{}) []*model.GitHubUser {
	var users []*model.GitHubUser
	for _, item := range data {
		if userMap, ok := item.(map[string]interface{}); ok {
			users = append(users, gm.parseUser(userMap))
		}
	}
	return users
}

// parseLabels 解析标签列表
func (gm *GitHubManager) parseLabels(data []interface{}) []string {
	var labels []string
	for _, item := range data {
		if labelMap, ok := item.(map[string]interface{}); ok {
			if name, ok := labelMap["name"].(string); ok {
				labels = append(labels, name)
			}
		}
	}
	return labels
}

// 辅助函数
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return 0
}

func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}

func getMap(data map[string]interface{}, key string) map[string]interface{} {
	if val, ok := data[key].(map[string]interface{}); ok {
		return val
	}
	return nil
}

func getArray(data map[string]interface{}, key string) []interface{} {
	if val, ok := data[key].([]interface{}); ok {
		return val
	}
	return []interface{}{}
}
