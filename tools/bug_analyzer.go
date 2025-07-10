package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"community-governance-mcp-higress/internal/agent"
)

// BugAnalyzer Bug分析器
type BugAnalyzer struct {
	config *agent.AgentConfig
}

// NewBugAnalyzer 创建新的Bug分析器
func NewBugAnalyzer() *BugAnalyzer {
	return &BugAnalyzer{}
}

// SetConfig 设置配置
func (b *BugAnalyzer) SetConfig(config *agent.AgentConfig) {
	b.config = config
}

// Analyze 分析Bug
func (b *BugAnalyzer) Analyze(ctx context.Context, stackTrace string, environment string) (*agent.BugAnalysis, error) {
	// 检测信息完整性
	missingInfo := b.detectMissingInformation(stackTrace, environment)
	if len(missingInfo) > 0 {
		// 如果有缺失信息，返回基础分析
		return b.generateBasicAnalysis(stackTrace, environment, missingInfo), nil
	}

	// 分析错误信息
	analysis := b.analyzeBug(stackTrace, environment)

	// 如果配置了AI，使用AI进行深度分析
	if b.config != nil && b.config.OpenAI.APIKey != "" {
		aiAnalysis, err := b.aiAnalyzeBug(ctx, stackTrace, environment)
		if err == nil {
			// 合并AI分析结果
			analysis.RootCause = aiAnalysis.RootCause
			analysis.Solutions = append(analysis.Solutions, aiAnalysis.Solutions...)
			analysis.Prevention = append(analysis.Prevention, aiAnalysis.Prevention...)
		}
	}

	return analysis, nil
}

// analyzeBug 分析Bug
func (b *BugAnalyzer) analyzeBug(stackTrace string, environment string) *agent.BugAnalysis {
	analysis := &agent.BugAnalysis{
		ErrorType:  b.classifyError(stackTrace),
		Language:   b.detectLanguage(stackTrace),
		Severity:   b.determineSeverity(stackTrace),
		RootCause:  b.analyzeRootCause(stackTrace),
		Solutions:  b.generateSolutions(stackTrace),
		Prevention: b.generatePrevention(stackTrace),
		Confidence: b.calculateConfidence(stackTrace, environment),
	}

	return analysis
}

// classifyError 分类错误类型
func (b *BugAnalyzer) classifyError(stackTrace string) string {
	errorMsg := strings.ToLower(stackTrace)

	if strings.Contains(errorMsg, "null pointer") || strings.Contains(errorMsg, "nil pointer") {
		return "空指针异常"
	}
	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
		return "网络连接问题"
	}
	if strings.Contains(errorMsg, "permission") || strings.Contains(errorMsg, "access denied") {
		return "权限问题"
	}
	if strings.Contains(errorMsg, "out of memory") || strings.Contains(errorMsg, "oom") {
		return "内存不足"
	}
	if strings.Contains(errorMsg, "parse") || strings.Contains(errorMsg, "syntax") {
		return "解析错误"
	}
	if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "404") {
		return "资源未找到"
	}
	if strings.Contains(errorMsg, "invalid") || strings.Contains(errorMsg, "bad request") {
		return "无效请求"
	}

	return "未知错误类型"
}

// detectLanguage 检测编程语言
func (b *BugAnalyzer) detectLanguage(stackTrace string) string {
	stackTrace = strings.ToLower(stackTrace)

	if strings.Contains(stackTrace, "java.lang") || strings.Contains(stackTrace, "exception") {
		return "java"
	}
	if strings.Contains(stackTrace, "panic:") || strings.Contains(stackTrace, "runtime error") {
		return "go"
	}
	if strings.Contains(stackTrace, "traceback") || strings.Contains(stackTrace, "python") {
		return "python"
	}
	if strings.Contains(stackTrace, "error:") || strings.Contains(stackTrace, "at ") {
		return "javascript"
	}

	return "unknown"
}

// determineSeverity 确定严重程度
func (b *BugAnalyzer) determineSeverity(stackTrace string) string {
	errorMsg := strings.ToLower(stackTrace)

	// 严重错误关键词
	criticalKeywords := []string{"panic", "fatal", "critical", "crash", "oom", "out of memory"}
	for _, keyword := range criticalKeywords {
		if strings.Contains(errorMsg, keyword) {
			return "critical"
		}
	}

	// 高优先级错误关键词
	highKeywords := []string{"error", "exception", "failed", "broken"}
	for _, keyword := range highKeywords {
		if strings.Contains(errorMsg, keyword) {
			return "high"
		}
	}

	// 中优先级错误关键词
	mediumKeywords := []string{"warning", "deprecated", "timeout"}
	for _, keyword := range mediumKeywords {
		if strings.Contains(errorMsg, keyword) {
			return "medium"
		}
	}

	return "low"
}

// analyzeRootCause 分析根本原因
func (b *BugAnalyzer) analyzeRootCause(stackTrace string) string {
	errorMsg := strings.ToLower(stackTrace)

	if strings.Contains(errorMsg, "null pointer") || strings.Contains(errorMsg, "nil pointer") {
		return "变量或对象未正确初始化，导致空指针引用"
	}
	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
		return "网络连接失败或超时，可能是网络配置问题或服务不可用"
	}
	if strings.Contains(errorMsg, "permission") || strings.Contains(errorMsg, "access denied") {
		return "权限不足，无法访问所需资源或执行操作"
	}
	if strings.Contains(errorMsg, "out of memory") || strings.Contains(errorMsg, "oom") {
		return "内存使用量超过限制，可能是内存泄漏或配置不当"
	}
	if strings.Contains(errorMsg, "parse") || strings.Contains(errorMsg, "syntax") {
		return "数据格式错误或配置文件语法不正确"
	}

	return "需要进一步分析以确定根本原因"
}

// generateSolutions 生成解决方案
func (b *BugAnalyzer) generateSolutions(stackTrace string) []string {
	var solutions []string
	errorMsg := strings.ToLower(stackTrace)

	if strings.Contains(errorMsg, "null pointer") || strings.Contains(errorMsg, "nil pointer") {
		solutions = append(solutions, "检查变量初始化，确保在使用前已正确赋值")
		solutions = append(solutions, "添加空值检查，避免直接访问可能为空的变量")
		solutions = append(solutions, "使用安全的访问方法，如可选链操作符")
	}

	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
		solutions = append(solutions, "检查网络连接状态和防火墙设置")
		solutions = append(solutions, "验证服务端点是否可访问")
		solutions = append(solutions, "增加连接超时时间或重试机制")
	}

	if strings.Contains(errorMsg, "permission") || strings.Contains(errorMsg, "access denied") {
		solutions = append(solutions, "检查文件或目录权限设置")
		solutions = append(solutions, "验证API密钥或访问令牌的有效性")
		solutions = append(solutions, "确认用户角色和权限配置")
	}

	if strings.Contains(errorMsg, "out of memory") || strings.Contains(errorMsg, "oom") {
		solutions = append(solutions, "增加内存限制或优化内存使用")
		solutions = append(solutions, "检查是否存在内存泄漏")
		solutions = append(solutions, "优化算法或数据结构以减少内存占用")
	}

	if strings.Contains(errorMsg, "parse") || strings.Contains(errorMsg, "syntax") {
		solutions = append(solutions, "验证配置文件格式和语法")
		solutions = append(solutions, "检查数据格式是否符合预期")
		solutions = append(solutions, "使用格式验证工具检查输入数据")
	}

	if len(solutions) == 0 {
		solutions = append(solutions, "查看完整的错误日志获取更多信息")
		solutions = append(solutions, "检查相关文档和最佳实践")
		solutions = append(solutions, "搜索类似问题的解决方案")
	}

	return solutions
}

// generatePrevention 生成预防措施
func (b *BugAnalyzer) generatePrevention(stackTrace string) []string {
	var prevention []string
	errorMsg := strings.ToLower(stackTrace)

	if strings.Contains(errorMsg, "null pointer") || strings.Contains(errorMsg, "nil pointer") {
		prevention = append(prevention, "在代码审查中重点关注空值检查")
		prevention = append(prevention, "使用静态分析工具检测潜在的空指针问题")
		prevention = append(prevention, "建立编码规范，要求显式初始化变量")
	}

	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
		prevention = append(prevention, "实施健康检查和监控机制")
		prevention = append(prevention, "使用连接池和重试机制")
		prevention = append(prevention, "定期测试网络连接和服务可用性")
	}

	if strings.Contains(errorMsg, "permission") || strings.Contains(errorMsg, "access denied") {
		prevention = append(prevention, "实施最小权限原则")
		prevention = append(prevention, "定期审查和更新权限配置")
		prevention = append(prevention, "使用自动化工具检查权限设置")
	}

	if strings.Contains(errorMsg, "out of memory") || strings.Contains(errorMsg, "oom") {
		prevention = append(prevention, "设置合理的内存限制和监控")
		prevention = append(prevention, "定期进行内存使用分析")
		prevention = append(prevention, "实施资源清理和垃圾回收优化")
	}

	if strings.Contains(errorMsg, "parse") || strings.Contains(errorMsg, "syntax") {
		prevention = append(prevention, "使用配置验证工具")
		prevention = append(prevention, "建立数据格式标准和验证流程")
		prevention = append(prevention, "实施自动化测试验证配置正确性")
	}

	if len(prevention) == 0 {
		prevention = append(prevention, "建立完善的日志记录和监控体系")
		prevention = append(prevention, "定期进行代码审查和测试")
		prevention = append(prevention, "建立问题跟踪和知识库")
	}

	return prevention
}

// calculateConfidence 计算置信度
func (b *BugAnalyzer) calculateConfidence(stackTrace string, environment string) float64 {
	confidence := 0.5

	// 基于错误信息的完整性调整置信度
	if stackTrace != "" {
		confidence += 0.2
	}
	if environment != "" {
		confidence += 0.1
	}

	// 基于错误类型的明确性调整置信度
	errorType := b.classifyError(stackTrace)
	if errorType != "未知错误类型" {
		confidence += 0.1
	}

	// 基于语言的检测结果调整置信度
	language := b.detectLanguage(stackTrace)
	if language != "unknown" {
		confidence += 0.1
	}

	return confidence
}

// detectMissingInformation 检测缺失信息
func (b *BugAnalyzer) detectMissingInformation(stackTrace string, environment string) []string {
	var missingInfo []string

	if stackTrace == "" {
		missingInfo = append(missingInfo, "错误堆栈信息")
	}

	if environment == "" {
		missingInfo = append(missingInfo, "环境信息（操作系统、版本等）")
	}

	return missingInfo
}

// generateBasicAnalysis 生成基础分析
func (b *BugAnalyzer) generateBasicAnalysis(stackTrace string, environment string, missingInfo []string) *agent.BugAnalysis {
	analysis := &agent.BugAnalysis{
		ErrorType:  "unknown",
		Language:   "unknown",
		Severity:   "medium",
		RootCause:  "需要更多信息进行分析",
		Solutions:  []string{"请提供完整的错误堆栈信息"},
		Prevention: []string{"建立完善的错误报告机制"},
		Confidence: 0.3,
	}

	if stackTrace != "" {
		analysis.ErrorType = b.classifyError(stackTrace)
		analysis.Language = b.detectLanguage(stackTrace)
		analysis.Severity = b.determineSeverity(stackTrace)
		analysis.Confidence = 0.5
	}

	return analysis
}

// aiAnalyzeBug AI分析Bug
func (b *BugAnalyzer) aiAnalyzeBug(ctx context.Context, stackTrace string, environment string) (*agent.BugAnalysis, error) {
	prompt := fmt.Sprintf(`作为一个技术专家，请分析以下 Bug 信息并提供详细的诊断和解决建议：

错误堆栈：%s
环境信息：%s

请提供：
1. 根本原因分析
2. 详细的解决步骤
3. 预防措施
4. 相关的最佳实践`,
		stackTrace, environment)

	requestBody := map[string]interface{}{
		"model": b.config.OpenAI.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_tokens": 500,
		"temperature": 0.2,
	}

	bodyBytes, _ := json.Marshal(requestBody)
	headers := map[string]string{
		"Authorization": "Bearer " + b.config.OpenAI.APIKey,
		"Content-Type":  "application/json",
	}

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 解析响应
	var aiResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&aiResponse); err != nil {
		return nil, err
	}

	choices, ok := aiResponse["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("AI响应格式错误")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	// 解析AI分析结果
	analysis := &agent.BugAnalysis{
		RootCause:  "AI分析：需要进一步处理",
		Solutions:  []string{content},
		Prevention: []string{"基于AI建议实施预防措施"},
		Confidence: 0.8,
	}

	return analysis, nil
}
