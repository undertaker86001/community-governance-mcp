package tools

import (
	"community-governance-mcp-higress/internal/agent"
	"github.com/sirupsen/logrus"
)

// ToolLoader 工具加载器
type ToolLoader struct {
	processor *agent.Processor
	logger    *logrus.Logger
	tools     map[string]interface{}
}

// NewToolLoader 创建新的工具加载器
func NewToolLoader(processor *agent.Processor) *ToolLoader {
	return &ToolLoader{
		processor: processor,
		logger:    logrus.New(),
		tools:     make(map[string]interface{}),
	}
}

// LoadTools 加载所有工具
func (p *ToolLoader) LoadTools() error {
	p.logger.Info("开始加载工具...")

	// 加载Bug分析器
	bugAnalyzer := NewBugAnalyzer(p.processor.GetConfig().OpenAI.APIKey)
	p.tools["bug_analyzer"] = bugAnalyzer

	// 加载问题分类器
	issueClassifier := NewIssueClassifier(p.processor.GetConfig().OpenAI.APIKey)
	p.tools["issue_classifier"] = issueClassifier

	// 加载图片分析器
	imageAnalyzer := NewImageAnalyzer(p.processor.GetConfig().OpenAI.APIKey)
	p.tools["image_analyzer"] = imageAnalyzer

	// 加载GitHub管理器
	githubManager := NewGitHubManager(p.processor.GetConfig().GitHub.Token)
	p.tools["github_manager"] = githubManager

	// 加载社区统计工具
	communityStats := NewCommunityStats(p.processor.GetConfig().GitHub.Token)
	p.tools["community_stats"] = communityStats

	// 加载知识库管理器
	knowledgeBase := NewKnowledgeBase(p.processor.GetConfig().Knowledge.StoragePath)
	p.tools["knowledge_base"] = knowledgeBase

	p.logger.WithField("tools_count", len(p.tools)).Info("工具加载完成")
	return nil
}

// GetTool 获取工具
func (p *ToolLoader) GetTool(name string) (interface{}, bool) {
	tool, exists := p.tools[name]
	return tool, exists
}

// ListTools 列出所有工具
func (p *ToolLoader) ListTools() []string {
	tools := make([]string, 0, len(p.tools))
	for name := range p.tools {
		tools = append(tools, name)
	}
	return tools
}

// RegisterTool 注册工具
func (p *ToolLoader) RegisterTool(name string, tool interface{}) {
	p.tools[name] = tool
	p.logger.WithField("tool_name", name).Info("工具注册成功")
}

// LoadTools 兼容性函数，用于向后兼容
func LoadTools(processor *agent.Processor) error {
	loader := NewToolLoader(processor)
	return loader.LoadTools()
}
