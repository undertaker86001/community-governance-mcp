package tools

import (
	"context"
	"fmt"
	"github.com/community-governance-mcp-higress/internal/model"
	"github.com/community-governance-mcp-higress/internal/openai"
	"strings"
)

// BugAnalyzer Bug分析器
type BugAnalyzer struct {
	openaiClient *openai.Client
}

// NewBugAnalyzer 创建新的Bug分析器
func NewBugAnalyzer(apiKey string) *BugAnalyzer {
	return &BugAnalyzer{
		openaiClient: openai.NewClient(apiKey, "gpt-4o"),
	}
}

// AnalyzeBug 分析Bug
func (b *BugAnalyzer) AnalyzeBug(stackTrace string, environment string, version string) (*model.BugAnalysisResult, error) {
	// 检测信息完整性
	missingInfo := b.detectMissingInformation(stackTrace, environment)
	if len(missingInfo) > 0 {
		// 如果有缺失信息，返回基础分析
		return b.generateBasicAnalysis(stackTrace, environment, version, missingInfo), nil
	}

	// 分析错误信息
	analysis := b.analyzeBug(stackTrace, environment, version)

	// 使用AI进行深度分析
	if b.openaiClient != nil {
		aiAnalysis, err := b.aiAnalyzeBug(context.Background(), stackTrace, environment, version)
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
func (b *BugAnalyzer) analyzeBug(stackTrace string, environment string, version string) *model.BugAnalysisResult {
	analysis := &model.BugAnalysisResult{
		ErrorType:  b.classifyError(stackTrace),
		Severity:   b.determineSeverity(stackTrace),
		RootCause:  b.analyzeRootCause(stackTrace),
		Solutions:  b.generateSolutions(stackTrace),
		Prevention: b.generatePrevention(stackTrace),
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
		solutions = append(solutions, "检查配置文件语法和格式")
		solutions = append(solutions, "验证输入数据的有效性")
		solutions = append(solutions, "使用适当的解析器和错误处理")
	}

	if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "404") {
		solutions = append(solutions, "检查文件路径或URL是否正确")
		solutions = append(solutions, "确认资源是否存在且可访问")
		solutions = append(solutions, "验证配置中的路径设置")
	}

	return solutions
}

// generatePrevention 生成预防措施
func (b *BugAnalyzer) generatePrevention(stackTrace string) []string {
	var prevention []string
	errorMsg := strings.ToLower(stackTrace)

	if strings.Contains(errorMsg, "null pointer") || strings.Contains(errorMsg, "nil pointer") {
		prevention = append(prevention, "建立代码审查流程，确保变量正确初始化")
		prevention = append(prevention, "使用静态分析工具检测潜在的空指针问题")
		prevention = append(prevention, "编写单元测试覆盖边界情况")
	}

	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
		prevention = append(prevention, "实施网络连接监控和告警机制")
		prevention = append(prevention, "配置适当的超时和重试策略")
		prevention = append(prevention, "定期检查网络配置和依赖服务状态")
	}

	if strings.Contains(errorMsg, "permission") || strings.Contains(errorMsg, "access denied") {
		prevention = append(prevention, "实施最小权限原则，定期审查权限配置")
		prevention = append(prevention, "使用自动化工具管理权限和访问控制")
		prevention = append(prevention, "建立权限变更的审批流程")
	}

	if strings.Contains(errorMsg, "out of memory") || strings.Contains(errorMsg, "oom") {
		prevention = append(prevention, "设置内存使用监控和告警")
		prevention = append(prevention, "定期进行内存使用分析和优化")
		prevention = append(prevention, "实施资源限制和配额管理")
	}

	return prevention
}

// detectMissingInformation 检测缺失信息
func (b *BugAnalyzer) detectMissingInformation(stackTrace string, environment string) []string {
	var missingInfo []string

	if stackTrace == "" {
		missingInfo = append(missingInfo, "错误堆栈信息")
	}

	if environment == "" {
		missingInfo = append(missingInfo, "环境信息")
	}

	return missingInfo
}

// generateBasicAnalysis 生成基础分析
func (b *BugAnalyzer) generateBasicAnalysis(stackTrace string, environment string, version string, missingInfo []string) *model.BugAnalysisResult {
	analysis := &model.BugAnalysisResult{
		ErrorType: "未知错误",
		Severity:  "medium",
		RootCause: "由于信息不完整，无法进行详细分析",
		Solutions: []string{
			"请提供完整的错误堆栈信息",
			"提供详细的环境配置信息",
			"包含相关的日志文件",
		},
		Prevention: []string{
			"建立标准化的错误报告流程",
			"配置自动化的错误信息收集",
			"定期进行系统健康检查",
		},
	}

	// 如果有一些信息，尝试基础分析
	if stackTrace != "" {
		analysis.ErrorType = b.classifyError(stackTrace)
		analysis.Severity = b.determineSeverity(stackTrace)
	}

	return analysis
}

// aiAnalyzeBug AI分析Bug
func (b *BugAnalyzer) aiAnalyzeBug(ctx context.Context, stackTrace string, environment string, version string) (*model.BugAnalysisResult, error) {
	prompt := fmt.Sprintf(`请分析以下错误信息，并提供详细的分析结果：

环境信息：%s
版本信息：%s
错误堆栈：
%s

请提供以下格式的分析结果：
1. 错误类型
2. 严重程度
3. 根本原因
4. 解决方案（3-5条）
5. 预防措施（3-5条）`, environment, version, stackTrace)

	_, err := b.openaiClient.GenerateText(ctx, prompt, 1000, 0.3)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析AI响应
	analysis := &model.BugAnalysisResult{
		ErrorType:  "AI分析结果",
		Severity:   "medium",
		RootCause:  "AI分析的根本原因",
		Solutions:  []string{"AI建议的解决方案"},
		Prevention: []string{"AI建议的预防措施"},
	}

	// 这里可以添加更复杂的响应解析逻辑
	// 目前返回基础结构，实际应用中可以根据AI响应格式进行解析

	return analysis, nil
}
