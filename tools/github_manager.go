package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/community-governance-mcp-higress/config"
	"github.com/community-governance-mcp-higress/internal/agent"
	"github.com/community-governance-mcp-higress/utils"
	"github.com/higress-group/wasm-go/pkg/mcp/server"
)

type GitHubManager struct {
	Action      string   `json:"action" jsonschema_description:"操作类型：list_issues, create_issue, update_issue, add_comment" jsonschema:"example=list_issues"`
	IssueNumber int      `json:"issue_number,omitempty" jsonschema_description:"Issue编号" jsonschema:"example=123"`
	Title       string   `json:"title,omitempty" jsonschema_description:"Issue标题"`
	Body        string   `json:"body,omitempty" jsonschema_description:"Issue内容或评论内容"`
	Labels      []string `json:"labels,omitempty" jsonschema_description:"标签列表"`
	State       string   `json:"state,omitempty" jsonschema_description:"Issue状态：open, closed" jsonschema:"example=open"`
}

func (t GitHubManager) Description() string {
	return `GitHub 仓库管理工具，支持 Issue 的创建、更新、列表查询和评论添加等操作。用于社区治理中的问题跟踪和管理。`
}

func (t GitHubManager) InputSchema() map[string]any {
	return server.ToInputSchema(&GitHubManager{})
}

func (t GitHubManager) Create(params []byte) server.Tool {
	githubManager := &GitHubManager{}
	json.Unmarshal(params, githubManager)
	return githubManager
}

func (t GitHubManager) Call(ctx server.HttpContext, s server.Server) error {
	serverConfig := &config.CommunityGovernanceConfig{}
	s.GetConfig(serverConfig)

	if serverConfig.GitHubToken == "" {
		return errors.New("GitHub token 未配置")
	}

	var url string
	var method string
	var requestBody string

	baseURL := fmt.Sprintf("https://api.github.com/repos/%s/%s",
		serverConfig.RepoOwner, serverConfig.RepoName)

	switch t.Action {
	case "list_issues":
		url = fmt.Sprintf("%s/issues?state=%s", baseURL, t.State)
		method = "GET"
	case "create_issue":
		url = fmt.Sprintf("%s/issues", baseURL)
		method = "POST"
		bodyData := map[string]interface{}{
			"title": t.Title,
			"body":  t.Body,
		}
		if len(t.Labels) > 0 {
			bodyData["labels"] = t.Labels
		}
		bodyBytes, _ := json.Marshal(bodyData)
		requestBody = string(bodyBytes)
	case "update_issue":
		url = fmt.Sprintf("%s/issues/%d", baseURL, t.IssueNumber)
		method = "PATCH"
		bodyData := map[string]interface{}{}
		if t.Title != "" {
			bodyData["title"] = t.Title
		}
		if t.Body != "" {
			bodyData["body"] = t.Body
		}
		if t.State != "" {
			bodyData["state"] = t.State
		}
		if len(t.Labels) > 0 {
			bodyData["labels"] = t.Labels
		}
		bodyBytes, _ := json.Marshal(bodyData)
		requestBody = string(bodyBytes)
	case "add_comment":
		url = fmt.Sprintf("%s/issues/%d/comments", baseURL, t.IssueNumber)
		method = "POST"
		bodyData := map[string]interface{}{
			"body": t.Body,
		}
		bodyBytes, _ := json.Marshal(bodyData)
		requestBody = string(bodyBytes)
	default:
		return errors.New("不支持的操作类型")
	}

	headers := map[string]string{
		"Authorization": "Bearer " + serverConfig.GitHubToken,
		"Accept":        "application/vnd.github+json",
		"Content-Type":  "application/json",
	}

	_, err := utils.SendHTTPRequest(ctx, method, url, headers, requestBody)
	if err != nil {
		return err
	}

	utils.SendMCPToolTextResult(ctx, "GitHub 操作完成", "", true)
	return nil
}
