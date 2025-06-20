package tools

import (
	"github.com/higress-group/wasm-go/pkg/mcp"
	"github.com/higress-group/wasm-go/pkg/mcp/server"
)

func LoadTools(mcpServer *mcp.MCPServer) server.Server {
	return mcpServer.
		AddMCPTool("github_manager", &GitHubManager{}).
		AddMCPTool("issue_classifier", &IssueClassifier{}).
		AddMCPTool("bug_analyzer", &BugAnalyzer{}).
		AddMCPTool("image_analyzer", &ImageAnalyzer{}).
		AddMCPTool("knowledge_base", &KnowledgeBase{}).
		AddMCPTool("community_stats", &CommunityStats{})
}
