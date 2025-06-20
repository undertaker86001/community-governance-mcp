package tools

import (
	"community-governance-mcp/config"
	mcp_tool "community-governance-mcp/utils"
	"encoding/json"
	"fmt"
	"github.com/higress-group/wasm-go/pkg/mcp/server"
	"github.com/higress-group/wasm-go/pkg/mcp/utils"
	"strings"
)

type BugAnalyzer struct {
	ErrorMessage     string `json:"error_message" jsonschema_description:"错误信息或堆栈跟踪"`
	Environment      string `json:"environment" jsonschema_description:"环境信息（OS、版本等）"`
	ReproduceSteps   string `json:"reproduce_steps" jsonschema_description:"重现步骤"`
	ExpectedBehavior string `json:"expected_behavior" jsonschema_description:"期望行为"`
	ActualBehavior   string `json:"actual_behavior" jsonschema_description:"实际行为"`
}

func (t BugAnalyzer) Description() string {
	return `Bug 分析工具，分析错误信息、环境配置和重现步骤，提供问题诊断和解决建议。`
}

func (t BugAnalyzer) InputSchema() map[string]any {
	return server.ToInputSchema(&BugAnalyzer{})
}

func (t BugAnalyzer) Create(params []byte) server.Tool {
	analyzer := &BugAnalyzer{}
	json.Unmarshal(params, analyzer)
	return analyzer
}

func (t BugAnalyzer) Call(ctx server.HttpContext, s server.Server) error {
	serverConfig := &config.CommunityGovernanceConfig{}
	s.GetConfig(serverConfig)

	// 检测信息完整性
	missingInfo := t.detectMissingInformation()
	if len(missingInfo) > 0 {
		// 如果有缺失信息，返回补全提示
		prompt := t.generateCompletionPrompt(missingInfo)
		utils.SendMCPToolTextResult(ctx, prompt)
		return nil
	}

	// 分析错误信息
	analysis := t.analyzeBug()

	// 如果配置了 AI，使用 AI 进行深度分析
	if serverConfig.OpenAIKey != "" {
		aiAnalysis, err := t.aiAnalyzeBug(ctx, serverConfig)
		if err == nil {
			analysis += "\n\n## AI 深度分析\n" + aiAnalysis
		}
	}

	utils.SendMCPToolTextResult(ctx, analysis)
	return nil
}

func (t BugAnalyzer) analyzeBug() string {
	var analysis strings.Builder

	analysis.WriteString("# Bug 分析报告\n\n")

	// 错误类型分析
	errorType := t.classifyError()
	analysis.WriteString(fmt.Sprintf("## 错误类型\n%s\n\n", errorType))

	// 环境分析
	if t.Environment != "" {
		analysis.WriteString(fmt.Sprintf("## 环境信息\n%s\n\n", t.Environment))
	}

	// 重现步骤分析
	if t.ReproduceSteps != "" {
		analysis.WriteString(fmt.Sprintf("## 重现步骤\n%s\n\n", t.ReproduceSteps))
	}

	// 行为对比
	if t.ExpectedBehavior != "" && t.ActualBehavior != "" {
		analysis.WriteString("## 行为对比\n")
		analysis.WriteString(fmt.Sprintf("**期望行为**: %s\n", t.ExpectedBehavior))
		analysis.WriteString(fmt.Sprintf("**实际行为**: %s\n\n", t.ActualBehavior))
	}

	// 建议解决方案
	suggestions := t.generateSuggestions()
	analysis.WriteString(fmt.Sprintf("## 建议解决方案\n%s\n", suggestions))

	return analysis.String()
}

func (t BugAnalyzer) classifyError() string {
	errorMsg := strings.ToLower(t.ErrorMessage)

	if strings.Contains(errorMsg, "null pointer") || strings.Contains(errorMsg, "nil pointer") {
		return "空指针异常 - 可能是未初始化的对象或变量"
	}
	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
		return "网络连接问题 - 检查网络配置和服务可用性"
	}
	if strings.Contains(errorMsg, "permission") || strings.Contains(errorMsg, "access denied") {
		return "权限问题 - 检查文件权限或API访问权限"
	}
	if strings.Contains(errorMsg, "out of memory") || strings.Contains(errorMsg, "oom") {
		return "内存不足 - 检查内存使用和配置"
	}
	if strings.Contains(errorMsg, "parse") || strings.Contains(errorMsg, "syntax") {
		return "解析错误 - 检查配置文件格式或数据格式"
	}

	return "未知错误类型 - 需要进一步分析"
}

func (t BugAnalyzer) generateSuggestions() string {
	var suggestions []string

	errorMsg := strings.ToLower(t.ErrorMessage)

	if strings.Contains(errorMsg, "connection") {
		suggestions = append(suggestions, "1. 检查网络连接状态")
		suggestions = append(suggestions, "2. 验证服务端点是否可访问")
		suggestions = append(suggestions, "3. 检查防火墙设置")
	}

	if strings.Contains(errorMsg, "config") || strings.Contains(errorMsg, "configuration") {
		suggestions = append(suggestions, "1. 验证配置文件格式")
		suggestions = append(suggestions, "2. 检查必需的配置项")
		suggestions = append(suggestions, "3. 对比正确的配置示例")
	}

	if strings.Contains(errorMsg, "version") || strings.Contains(errorMsg, "compatibility") {
		suggestions = append(suggestions, "1. 检查版本兼容性")
		suggestions = append(suggestions, "2. 升级到推荐版本")
		suggestions = append(suggestions, "3. 查看版本变更日志")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "1. 查看完整的错误日志")
		suggestions = append(suggestions, "2. 检查相关文档")
		suggestions = append(suggestions, "3. 搜索类似问题的解决方案")
	}

	return strings.Join(suggestions, "\n")
}

func (t BugAnalyzer) aiAnalyzeBug(ctx server.HttpContext, config *config.CommunityGovernanceConfig) (string, error) {
	prompt := fmt.Sprintf(`作为一个技术专家，请分析以下 Bug 信息并提供详细的诊断和解决建议：  
  
错误信息：%s  
环境信息：%s  
重现步骤：%s  
期望行为：%s  
实际行为：%s  
  
请提供：  
1. 根本原因分析  
2. 详细的解决步骤  
3. 预防措施  
4. 相关的最佳实践`,
		t.ErrorMessage, t.Environment, t.ReproduceSteps, t.ExpectedBehavior, t.ActualBehavior)

	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 500,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	headers := map[string]string{
		"Authorization": "Bearer " + config.OpenAIKey,
		"Content-Type":  "application/json",
	}

	response, err := mcp_tool.SendHTTPRequest(ctx, "POST", "https://api.openai.com/v1/chat/completions", headers, string(bodyBytes))
	if err != nil {
		return "", err
	}

	var aiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(response), &aiResponse); err != nil {
		return "", err
	}

	choices, ok := aiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("AI 响应格式错误")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	return content, nil
}

func (t BugAnalyzer) detectMissingInformation() []string {
	var missing []string

	// 检测各个字段是否为空或只包含堆栈信息
	if t.ErrorMessage == "" {
		missing = append(missing, "错误信息")
	}

	if t.Environment == "" {
		missing = append(missing, "环境信息")
	}

	if t.ReproduceSteps == "" {
		missing = append(missing, "重现步骤")
	}

	if t.ExpectedBehavior == "" {
		missing = append(missing, "期望行为")
	}

	if t.ActualBehavior == "" {
		missing = append(missing, "实际行为")
	}

	// 特殊检测：如果只有错误信息且看起来像堆栈跟踪
	if len(missing) == 4 && t.ErrorMessage != "" && t.isStackTrace(t.ErrorMessage) {
		// 用户只提供了堆栈信息，需要其他信息
		return []string{"环境信息", "重现步骤", "期望行为", "实际行为"}
	}

	return missing
}

func (t BugAnalyzer) isStackTrace(content string) bool {
	stackTraceIndicators := []string{
		"at ", "stack trace", "stacktrace", "exception",
		"error:", "panic:", "fatal:", "caused by:",
		".java:", ".go:", ".py:", ".js:", ".ts:",
		"line ", "function ", "method ",
	}

	content = strings.ToLower(content)
	indicatorCount := 0

	for _, indicator := range stackTraceIndicators {
		if strings.Contains(content, indicator) {
			indicatorCount++
		}
	}

	// 如果包含多个堆栈跟踪指示符，认为是堆栈信息
	return indicatorCount >= 2
}

func (t BugAnalyzer) generateCompletionPrompt(missingInfo []string) string {
	var prompt strings.Builder

	prompt.WriteString("# 信息不完整提示\n\n")
	prompt.WriteString("为了更好地分析您的问题，请补充以下信息：\n\n")

	for i, info := range missingInfo {
		switch info {
		case "错误信息":
			prompt.WriteString(fmt.Sprintf("%d. **错误信息**: 请提供完整的错误消息或堆栈跟踪\n", i+1))
		case "环境信息":
			prompt.WriteString(fmt.Sprintf("%d. **环境信息**: 请提供以下信息：\n", i+1))
			prompt.WriteString("   - 操作系统版本\n")
			prompt.WriteString("   - Higress 版本\n")
			prompt.WriteString("   - 相关组件版本（如 Kubernetes、Docker 等）\n")
			prompt.WriteString("   - 硬件配置（如内存、CPU）\n")
		case "重现步骤":
			prompt.WriteString(fmt.Sprintf("%d. **重现步骤**: 请详细描述如何重现这个问题：\n", i+1))
			prompt.WriteString("   - 具体的操作步骤\n")
			prompt.WriteString("   - 使用的配置文件\n")
			prompt.WriteString("   - 相关的命令或请求\n")
		case "期望行为":
			prompt.WriteString(fmt.Sprintf("%d. **期望行为**: 请描述您期望系统应该如何工作\n", i+1))
		case "实际行为":
			prompt.WriteString(fmt.Sprintf("%d. **实际行为**: 请描述实际发生了什么，与期望有什么不同\n", i+1))
		}
		prompt.WriteString("\n")
	}

	// 如果用户只提供了堆栈信息，给出特殊提示
	if len(missingInfo) == 4 && t.ErrorMessage != "" {
		prompt.WriteString("## 检测到堆栈信息\n")
		prompt.WriteString("我发现您提供了错误堆栈信息，这很有帮助！为了更准确地诊断问题，请补充上述其他信息。\n\n")

		// 尝试从堆栈中提取一些有用信息
		extractedInfo := t.extractInfoFromStack()
		if extractedInfo != "" {
			prompt.WriteString("## 从堆栈中提取的信息\n")
			prompt.WriteString(extractedInfo)
			prompt.WriteString("\n")
		}
	}

	prompt.WriteString("---\n")
	prompt.WriteString("**提示**: 信息越详细，我就能提供越准确的问题诊断和解决方案。")

	return prompt.String()
}

func (t BugAnalyzer) extractInfoFromStack() string {
	if t.ErrorMessage == "" {
		return ""
	}

	var extracted strings.Builder
	content := strings.ToLower(t.ErrorMessage)

	// 尝试识别错误类型
	if strings.Contains(content, "null pointer") || strings.Contains(content, "nil pointer") {
		extracted.WriteString("- **错误类型**: 空指针异常\n")
	} else if strings.Contains(content, "connection") || strings.Contains(content, "timeout") {
		extracted.WriteString("- **错误类型**: 网络连接问题\n")
	} else if strings.Contains(content, "out of memory") || strings.Contains(content, "oom") {
		extracted.WriteString("- **错误类型**: 内存不足\n")
	} else if strings.Contains(content, "permission") || strings.Contains(content, "access denied") {
		extracted.WriteString("- **错误类型**: 权限问题\n")
	}

	// 尝试识别编程语言
	if strings.Contains(content, ".java:") {
		extracted.WriteString("- **可能的语言**: Java\n")
	} else if strings.Contains(content, ".go:") {
		extracted.WriteString("- **可能的语言**: Go\n")
	} else if strings.Contains(content, ".py:") {
		extracted.WriteString("- **可能的语言**: Python\n")
	}

	return extracted.String()
}
