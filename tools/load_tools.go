package tools

import (
	"community-governance-mcp-higress/internal/agent"
)

// LoadTools 加载所有工具到处理器中
func LoadTools(processor *agent.Processor) {
	// 加载Bug分析器
	bugAnalyzer := NewBugAnalyzer()
	processor.RegisterTool("bug_analyzer", bugAnalyzer)

	// 加载图片分析器
	imageAnalyzer := NewImageAnalyzer()
	processor.RegisterTool("image_analyzer", imageAnalyzer)

	// 加载Issue分类器
	issueClassifier := NewIssueClassifier()
	processor.RegisterTool("issue_classifier", issueClassifier)

	// 加载GitHub管理器
	githubManager := NewGitHubManager()
	processor.RegisterTool("github_manager", githubManager)

	// 加载社区统计工具
	communityStats := NewCommunityStats()
	processor.RegisterTool("community_stats", communityStats)

	// 加载知识库工具
	knowledgeBase := NewKnowledgeBase()
	processor.RegisterTool("knowledge_base", knowledgeBase)
}
