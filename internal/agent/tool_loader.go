package agent

import (
	"github.com/community-governance-mcp-higress/tools"
	"github.com/sirupsen/logrus"
)

// ToolLoader 工具加载器
// 用于统一加载和管理所有工具，避免循环依赖
type ToolLoader struct {
	processor *Processor
	logger    *logrus.Logger
	tools     map[string]interface{}
}

// NewToolLoader 创建新的工具加载器
func NewToolLoader(processor *Processor) *ToolLoader {
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
	bugAnalyzer := tools.NewBugAnalyzer(p.processor.Config.OpenAI.APIKey)
	p.tools["bug_analyzer"] = bugAnalyzer

	// 加载问题分类器
	issueClassifier := tools.NewIssueClassifier(p.processor.Config.OpenAI.APIKey)
	p.tools["issue_classifier"] = issueClassifier

	// 加载图片分析器
	imageAnalyzer := tools.NewImageAnalyzer(p.processor.Config.OpenAI.APIKey)
	p.tools["image_analyzer"] = imageAnalyzer

	// 加载GitHub管理器
	githubManager := tools.NewGitHubManager(p.processor.Config.GitHub.Token)
	p.tools["github_manager"] = githubManager

	// 加载社区统计工具
	communityStats := tools.NewCommunityStats(p.processor.Config.GitHub.Token)
	p.tools["community_stats"] = communityStats

	// 加载知识库管理器
	knowledgeBase := tools.NewKnowledgeBase(p.processor.Config.Knowledge.StoragePath)
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
	toolsList := make([]string, 0, len(p.tools))
	for name := range p.tools {
		toolsList = append(toolsList, name)
	}
	return toolsList
}

// RegisterTool 注册工具
func (p *ToolLoader) RegisterTool(name string, tool interface{}) {
	p.tools[name] = tool
	p.logger.WithField("tool_name", name).Info("工具注册成功")
}

// LoadTools 兼容性函数，用于向后兼容
func LoadTools(processor *Processor) error {
	loader := NewToolLoader(processor)
	return loader.LoadTools()
}
