package test

import (
	"encoding/json"
	"fmt"
	"github.com/community-governance-mcp-higress/config"
	"github.com/community-governance-mcp-higress/intent"
	"github.com/higress-group/wasm-go/pkg/mcp"
	"log"
	"net/http"
	"strings"
)

type TestServer struct {
	McpServer        *mcp.MCPServer
	Config           *config.CommunityGovernanceConfig
	IntentRecognizer *intent.IntentRecognizer
}

type ChatRequest struct {
	Message  string `json:"message"`
	ImageURL string `json:"image_url,omitempty"`
	Context  string `json:"context,omitempty"`
}

type ChatResponse struct {
	Intent     string  `json:"intent"`
	ToolUsed   string  `json:"tool_used"`
	Response   string  `json:"response"`
	Confidence float64 `json:"confidence"`
	Reasoning  string  `json:"reasoning"`
}

func (ts *TestServer) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 使用LLM进行意图识别
	intentResult, err := ts.IntentRecognizer.RecognizeIntent(req.Message, req.ImageURL, req.Context)
	if err != nil {
		log.Printf("意图识别失败: %v", err)
		http.Error(w, "Intent recognition failed", http.StatusInternalServerError)
		return
	}

	// 执行相应的工具
	response := ts.executeTool(intentResult.ToolName, req, intentResult.Intent)

	chatResp := ChatResponse{
		Intent:     intentResult.Intent,
		ToolUsed:   intentResult.ToolName,
		Response:   response,
		Confidence: intentResult.Confidence,
		Reasoning:  intentResult.Reasoning,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chatResp)
}

func (ts *TestServer) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (ts *TestServer) executeTool(toolName string, req ChatRequest, intent string) string {
	switch toolName {
	case "image_analyzer":
		return ts.executeImageAnalyzer(req, intent)
	case "bug_analyzer":
		return ts.executeBugAnalyzer(req)
	case "issue_classifier":
		return ts.executeIssueClassifier(req)
	case "github_manager":
		return ts.executeGitHubManager(req)
	case "community_stats":
		return ts.executeCommunityStats(req)
	case "knowledge_base":
		return ts.executeKnowledgeBase(req)
	default:
		return "抱歉，我无法理解您的请求。请提供更多详细信息。"
	}
}

// 其他执行工具的方法保持不变...
func (ts *TestServer) executeImageAnalyzer(req ChatRequest, intent string) string {
	analysisType := "general"
	if strings.Contains(intent, "bug") || strings.Contains(intent, "error") {
		analysisType = "error_screenshot"
	}

	return fmt.Sprintf(`# 图片分析结果  
  
**分析类型**: %s  
**图片URL**: %s  
**上下文**: %s  
**识别意图**: %s  
  
基于LLM意图识别，这是一个%s相关的分析请求。  
建议您检查相关配置和日志信息。`, analysisType, req.ImageURL, req.Context, intent, analysisType)
}

func (ts *TestServer) executeBugAnalyzer(req ChatRequest) string {
	return fmt.Sprintf(`# Bug 分析报告  
  
## 问题描述  
%s  
  
## LLM智能分析  
基于大语言模型的分析，这可能是以下几种情况之一：  
1. 配置问题 - 检查相关配置文件  
2. 环境兼容性问题 - 验证环境依赖  
3. 代码逻辑错误 - 查看完整的错误日志  
  
## 建议解决方案  
1. 提供完整的错误堆栈信息  
2. 描述环境信息（OS、版本等）  
3. 提供重现步骤  
4. 说明期望行为和实际行为  
  
如需更详细的分析，请使用 bug_analyzer 工具并提供完整信息。`, req.Message)
}

func (ts *TestServer) executeIssueClassifier(req ChatRequest) string {
	return fmt.Sprintf(`# Issue 智能分类结果  
  
## 原始内容  
%s  
  
## LLM分类建议  
基于大语言模型分析，推荐以下标签：  
  
**主要类型**:  
- bug（如果描述了软件缺陷）  
- enhancement（如果是功能增强请求）  
- documentation（如果是文档相关）  
- question（如果是问题咨询）  
  
**优先级**:  
- priority/high（紧急问题）  
- priority/medium（一般问题）  
- priority/low（低优先级）  
  
**其他标签**:  
- good first issue（适合新手）  
- help wanted（需要社区帮助）  
  
## 分类依据  
基于LLM对内容语义的深度理解和上下文分析。`, req.Message)
}

func (ts *TestServer) executeGitHubManager(req ChatRequest) string {
	return fmt.Sprintf(`# GitHub 仓库管理  
  
## 请求内容分析  
%s  
  
## 可执行操作  
基于您的请求，我可以帮您执行以下GitHub操作：  
  
1. **Issue管理**  
   - 列出开放/关闭的Issues  
   - 创建新Issue  
   - 更新Issue状态和标签  
   - 添加评论  
  
2. **PR管理**  
   - 查看PR状态  
   - 添加审核评论  
   - 合并PR  
  
3. **仓库统计**  
   - 获取贡献者信息  
   - 查看项目活跃度  
  
## 下一步  
请明确指定您需要执行的具体操作，我将调用相应的GitHub API为您处理。`, req.Message)
}

func (ts *TestServer) executeCommunityStats(req ChatRequest) string {
	return fmt.Sprintf(`# 社区统计报告  
  
## 请求分析  
基于您的请求：%s  
  
## 当前统计数据（示例）  
  
### Issues 概况  
- 总Issues数量: 156  
- 开放Issues: 23  
- 已关闭Issues: 133  
- 关闭率: 85.3%%  
  
### 贡献者统计  
- 总贡献者: 45  
- 活跃贡献者（近30天）: 12  
- 新贡献者（近30天）: 3  
  
### 活跃度指标  
- 平均Issue响应时间: 2.5天  
- 平均PR合并时间: 4.2天  
- 社区参与度: 高  
  
### 项目健康度评估  
✅ 项目健康状况良好  
- Issue处理效率高  
- 社区活跃度稳定  
- 贡献者增长良好  
  
## 详细报告  
如需更详细的统计数据，请使用 community_stats 工具指定具体的统计类型。`, req.Message)
}

func (ts *TestServer) executeKnowledgeBase(req ChatRequest) string {
	return fmt.Sprintf(`# 知识库搜索结果  
  
## 搜索查询  
%s  
  
## 相关文档  
1. **Higress 快速开始指南**  
   - 安装和基本配置  
   - 常见使用场景  
   - [查看文档](https://higress.cn/docs/latest/overview/what-is-higress/)  
  
2. **AI Gateway 配置**  
   - LLM提供商配置  
   - 插件开发指南  
   - [查看文档](https://higress.cn/docs/latest/ai/ai-gateway/)  
  
3. **MCP 服务器开发**  
   - 工具开发框架  
   - 最佳实践  
   - [查看文档](https://higress.cn/docs/latest/ai/mcp-quick-start/)  
  
## 相关 Issues  
- [Issue #123] 类似问题的解决方案  
- [Issue #456] 相关配置问题讨论  
- [Issue #789] 社区最佳实践分享  
  
## 最佳实践建议  
基于您的查询，建议关注以下方面：  
- 配置文件的正确格式  
- 插件的生命周期管理  
- 错误处理和日志记录  
  
## 需要更多帮助？  
如果以上信息不能解决您的问题，请：  
1. 访问 [Higress官方文档](https://higress.cn/docs/)  
2. 在GitHub上提交Issue  
3. 加入社区讨论群`, req.Message)
}
