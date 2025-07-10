package agent

import (
	"fmt"
	"sync"

	"github.com/community-governance-mcp-higress/internal/model"
	"github.com/community-governance-mcp-higress/tools"
)

// ToolLoader 工具加载器
type ToolLoader struct {
	tools map[string]interface{}
	mutex sync.RWMutex
}

// NewToolLoader 创建新的工具加载器
func NewToolLoader() *ToolLoader {
	return &ToolLoader{
		tools: make(map[string]interface{}),
	}
}

// LoadTools 加载所有工具
func (tl *ToolLoader) LoadTools(config *model.Config) error {
	tl.mutex.Lock()
	defer tl.mutex.Unlock()

	// 加载Bug分析器
	if config.OpenAIKey != "" {
		bugAnalyzer := tools.NewBugAnalyzer(config.OpenAIKey)
		tl.tools["bug_analyzer"] = bugAnalyzer
	}

	// 加载图片分析器
	if config.OpenAIKey != "" {
		imageAnalyzer := tools.NewImageAnalyzer(config.OpenAIKey)
		tl.tools["image_analyzer"] = imageAnalyzer
	}

	// 加载社区统计工具
	if config.GitHubToken != "" {
		communityStats := tools.NewCommunityStats(config.GitHubToken)
		tl.tools["community_stats"] = communityStats
	}

	// 加载Issue分类器
	if config.OpenAIKey != "" {
		issueClassifier := tools.NewIssueClassifier(config.OpenAIKey)
		tl.tools["issue_classifier"] = issueClassifier
	}

	// 加载知识库
	if config.OpenAIKey != "" {
		knowledgeBase := tools.NewKnowledgeBase(config.OpenAIKey)
		tl.tools["knowledge_base"] = knowledgeBase
	}

	// 加载GitHub管理器
	if config.GitHubToken != "" {
		githubManager := tools.NewGitHubManager(config.GitHubToken)
		tl.tools["github_manager"] = githubManager
	}

	return nil
}

// GetTool 获取工具
func (tl *ToolLoader) GetTool(name string) (interface{}, error) {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	tool, exists := tl.tools[name]
	if !exists {
		return nil, fmt.Errorf("工具未找到: %s", name)
	}

	return tool, nil
}

// GetToolNames 获取所有工具名称
func (tl *ToolLoader) GetToolNames() []string {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	var names []string
	for name := range tl.tools {
		names = append(names, name)
	}

	return names
}

// HasTool 检查是否有指定工具
func (tl *ToolLoader) HasTool(name string) bool {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	_, exists := tl.tools[name]
	return exists
}

// GetToolCount 获取工具数量
func (tl *ToolLoader) GetToolCount() int {
	tl.mutex.RLock()
	defer tl.mutex.RUnlock()

	return len(tl.tools)
}
