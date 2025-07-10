package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"community-governance-mcp-higress/internal/openai"
	"community-governance-mcp-higress/tools"
)

// Processor 智能代理处理器
type Processor struct {
	openaiClient *openai.Client
	logger       *logrus.Logger
	config       *AgentConfig
	tools        map[string]interface{}
}

// NewProcessor 创建新的处理器
func NewProcessor(openaiClient *openai.Client, config *AgentConfig) *Processor {
	return &Processor{
		openaiClient: openaiClient,
		logger:       logrus.New(),
		config:       config,
		tools:        make(map[string]interface{}),
	}
}

// RegisterTool 注册工具
func (p *Processor) RegisterTool(name string, tool interface{}) {
	p.tools[name] = tool
}

// ProcessQuestion 处理用户问题（智能问答）
func (p *Processor) ProcessQuestion(ctx context.Context, request *ProcessRequest) (*ProcessResponse, error) {
	startTime := time.Now()
	
	// 生成问题ID
	questionID := uuid.New().String()
	
	p.logger.WithFields(logrus.Fields{
		"question_id": questionID,
		"type":        request.Type,
		"author":      request.Author,
	}).Info("开始处理用户问题")

	// 1. 问题理解和分类
	question, err := p.understandQuestion(ctx, request, questionID)
	if err != nil {
		return nil, fmt.Errorf("问题理解失败: %w", err)
	}

	// 2. 多源知识检索
	sources, err := p.retrieveKnowledge(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("知识检索失败: %w", err)
	}

	// 3. 知识融合
	fusionResult, err := p.fuseKnowledge(ctx, question, sources)
	if err != nil {
		return nil, fmt.Errorf("知识融合失败: %w", err)
	}

	// 4. 生成回答
	answer, err := p.generateAnswer(ctx, fusionResult)
	if err != nil {
		return nil, fmt.Errorf("生成回答失败: %w", err)
	}

	// 5. 构建响应
	processingTime := time.Since(startTime)
	response := &ProcessResponse{
		ID:             uuid.New().String(),
		QuestionID:     questionID,
		Content:        answer.Content,
		Summary:        answer.Summary,
		Sources:        answer.Sources,
		Confidence:     answer.Confidence,
		ProcessingTime: processingTime.String(),
		FusionScore:    fusionResult.FusionScore,
		Recommendations: p.generateRecommendations(question, answer),
	}

	p.logger.WithFields(logrus.Fields{
		"question_id":     questionID,
		"processing_time": processingTime,
		"confidence":      answer.Confidence,
		"sources_count":   len(sources),
	}).Info("问题处理完成")

	return response, nil
}

// AnalyzeProblem 分析问题（Bug分析、图片分析等）
func (p *Processor) AnalyzeProblem(ctx context.Context, request *AnalyzeRequest) (*AnalyzeResponse, error) {
	startTime := time.Now()
	
	analysisID := uuid.New().String()
	
	p.logger.WithFields(logrus.Fields{
		"analysis_id": analysisID,
		"issue_type":  request.IssueType,
	}).Info("开始分析问题")

	var analysis interface{}
	var err error

	// 根据问题类型选择分析工具
	switch request.IssueType {
	case "bug", "stack_trace":
		analysis, err = p.analyzeBug(ctx, request)
	case "image", "screenshot":
		analysis, err = p.analyzeImage(ctx, request)
	case "issue", "github_issue":
		analysis, err = p.classifyIssue(ctx, request)
	default:
		analysis, err = p.analyzeGeneral(ctx, request)
	}

	if err != nil {
		return nil, fmt.Errorf("问题分析失败: %w", err)
	}

	// 构建响应
	processingTime := time.Since(startTime)
	response := &AnalyzeResponse{
		ID:             analysisID,
		ProblemType:    request.IssueType,
		ProcessingTime: processingTime.String(),
	}

	// 根据分析类型填充响应
	switch a := analysis.(type) {
	case *BugAnalysis:
		response.Severity = a.Severity
		response.Diagnosis = a.RootCause
		response.Solutions = a.Solutions
		response.Confidence = a.Confidence
	case *ImageAnalysis:
		response.Diagnosis = strings.Join(a.ErrorMessages, "; ")
		response.Solutions = a.Suggestions
		response.Confidence = a.Confidence
	case *IssueClassification:
		response.Diagnosis = fmt.Sprintf("分类: %s, 优先级: %s", a.Category, a.Priority)
		response.Solutions = []string{"建议分配给: " + strings.Join(a.Assignees, ", ")}
		response.Confidence = a.Confidence
	}

	p.logger.WithFields(logrus.Fields{
		"analysis_id":    analysisID,
		"processing_time": processingTime,
		"confidence":     response.Confidence,
	}).Info("问题分析完成")

	return response, nil
}

// GetCommunityStats 获取社区统计
func (p *Processor) GetCommunityStats(ctx context.Context) (*CommunityStats, error) {
	p.logger.Info("开始获取社区统计")

	// 获取GitHub统计工具
	statsTool, ok := p.tools["community_stats"].(*tools.CommunityStats)
	if !ok {
		return nil, fmt.Errorf("社区统计工具未找到")
	}

	// 获取统计数据
	stats, err := statsTool.GetStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取社区统计失败: %w", err)
	}

	p.logger.Info("社区统计获取完成")
	return stats, nil
}

// understandQuestion 理解问题
func (p *Processor) understandQuestion(ctx context.Context, request *ProcessRequest, questionID string) (*Question, error) {
	// 确定问题类型
	questionType := p.determineQuestionType(request)
	
	// 提取关键词和标签
	tags := p.extractTags(request)
	
	// 确定优先级
	priority := p.determinePriority(request)
	
	// 构建问题对象
	question := &Question{
		ID:        questionID,
		Type:      questionType,
		Title:     request.Title,
		Content:   request.Content,
		Author:    request.Author,
		Priority:  priority,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Metadata:  request.Metadata,
	}

	p.logger.WithFields(logrus.Fields{
		"question_id": questionID,
		"type":        questionType,
		"priority":    priority,
		"tags":        tags,
	}).Info("问题理解完成")

	return question, nil
}

// determineQuestionType 确定问题类型
func (p *Processor) determineQuestionType(request *ProcessRequest) QuestionType {
	// 基于请求类型确定
	if request.Type != "" {
		return QuestionType(request.Type)
	}

	// 基于内容分析确定类型
	content := strings.ToLower(request.Content)
	title := strings.ToLower(request.Title)

	// 检查是否为Issue
	if strings.Contains(content, "issue") || strings.Contains(title, "issue") ||
		strings.Contains(content, "bug") || strings.Contains(title, "bug") ||
		strings.Contains(content, "error") || strings.Contains(title, "error") {
		return QuestionTypeIssue
	}

	// 检查是否为PR
	if strings.Contains(content, "pull request") || strings.Contains(title, "pull request") ||
		strings.Contains(content, "pr") || strings.Contains(title, "pr") ||
		strings.Contains(content, "merge") || strings.Contains(title, "merge") {
		return QuestionTypePR
	}

	// 检查是否为配置问题
	if strings.Contains(content, "config") || strings.Contains(title, "config") ||
		strings.Contains(content, "configuration") || strings.Contains(title, "configuration") {
		return QuestionTypeConfig
	}

	// 默认为文本问题
	return QuestionTypeText
}

// extractTags 提取标签
func (p *Processor) extractTags(request *ProcessRequest) []string {
	tags := make([]string, 0)
	
	// 添加用户提供的标签
	if request.Tags != nil {
		tags = append(tags, request.Tags...)
	}

	// 基于内容提取标签
	content := strings.ToLower(request.Content + " " + request.Title)
	
	// Higress相关标签
	higressKeywords := []string{"gateway", "plugin", "route", "config", "deployment", "ingress"}
	for _, keyword := range higressKeywords {
		if strings.Contains(content, keyword) {
			tags = append(tags, keyword)
		}
	}

	// 技术标签
	techKeywords := []string{"kubernetes", "docker", "helm", "yaml", "json", "go", "java"}
	for _, keyword := range techKeywords {
		if strings.Contains(content, keyword) {
			tags = append(tags, keyword)
		}
	}

	return tags
}

// determinePriority 确定优先级
func (p *Processor) determinePriority(request *ProcessRequest) Priority {
	// 如果用户指定了优先级，直接使用
	if request.Priority != "" {
		return Priority(request.Priority)
	}

	// 基于关键词确定优先级
	content := strings.ToLower(request.Content + " " + request.Title)
	
	// 紧急关键词
	urgentKeywords := []string{"urgent", "critical", "broken", "crash", "production", "blocking"}
	for _, keyword := range urgentKeywords {
		if strings.Contains(content, keyword) {
			return PriorityUrgent
		}
	}

	// 高优先级关键词
	highKeywords := []string{"important", "high", "major", "significant"}
	for _, keyword := range highKeywords {
		if strings.Contains(content, keyword) {
			return PriorityHigh
		}
	}

	// 低优先级关键词
	lowKeywords := []string{"minor", "low", "nice-to-have", "enhancement"}
	for _, keyword := range lowKeywords {
		if strings.Contains(content, keyword) {
			return PriorityLow
		}
	}

	// 默认为普通优先级
	return PriorityNormal
}

// retrieveKnowledge 检索知识
func (p *Processor) retrieveKnowledge(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	sources := make([]KnowledgeItem, 0)

	// 1. 检索本地知识库
	if p.config.Knowledge.Enabled {
		localSources, err := p.retrieveLocalKnowledge(ctx, question)
		if err != nil {
			p.logger.WithError(err).Warn("本地知识库检索失败")
		} else {
			sources = append(sources, localSources...)
		}
	}

	// 2. 检索Higress文档
	higressSources, err := p.retrieveHigressDocs(ctx, question)
	if err != nil {
		p.logger.WithError(err).Warn("Higress文档检索失败")
	} else {
		sources = append(sources, higressSources...)
	}

	// 3. 检索DeepWiki知识
	if p.config.DeepWiki.Enabled {
		deepwikiSources, err := p.retrieveDeepWiki(ctx, question)
		if err != nil {
			p.logger.WithError(err).Warn("DeepWiki检索失败")
		} else {
			sources = append(sources, deepwikiSources...)
		}
	}

	// 4. 按相关性排序
	p.sortByRelevance(sources)

	// 5. 限制来源数量
	if len(sources) > p.config.Fusion.MaxSources {
		sources = sources[:p.config.Fusion.MaxSources]
	}

	p.logger.WithField("sources_count", len(sources)).Info("知识检索完成")
	return sources, nil
}

// retrieveLocalKnowledge 检索本地知识库
func (p *Processor) retrieveLocalKnowledge(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	// 这里实现本地知识库检索逻辑
	// 可以使用向量数据库或简单的关键词匹配
	return []KnowledgeItem{}, nil
}

// retrieveHigressDocs 检索Higress文档
func (p *Processor) retrieveHigressDocs(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	// 这里实现Higress文档检索逻辑
	// 可以调用Higress官方文档API或爬虫
	return []KnowledgeItem{}, nil
}

// retrieveDeepWiki 检索DeepWiki知识
func (p *Processor) retrieveDeepWiki(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	// 这里实现DeepWiki检索逻辑
	// 可以调用DeepWiki MCP服务
	return []KnowledgeItem{}, nil
}

// fuseKnowledge 融合知识
func (p *Processor) fuseKnowledge(ctx context.Context, question *Question, sources []KnowledgeItem) (*FusionResult, error) {
	// 计算融合分数
	fusionScore := p.calculateFusionScore(sources)

	// 构建融合结果
	fusionResult := &FusionResult{
		Question:    question,
		Sources:     sources,
		FusionScore: fusionScore,
	}

	p.logger.WithFields(logrus.Fields{
		"fusion_score": fusionScore,
		"sources_count": len(sources),
	}).Info("知识融合完成")

	return fusionResult, nil
}

// calculateRelevance 计算相关性
func (p *Processor) calculateRelevance(question *Question, source *KnowledgeItem) float64 {
	// 简单的关键词匹配算法
	// 在实际应用中可以使用更复杂的语义相似度算法
	
	questionText := strings.ToLower(question.Title + " " + question.Content)
	sourceText := strings.ToLower(source.Title + " " + source.Content)
	
	// 计算关键词匹配度
	keywords := append(question.Tags, strings.Fields(questionText)...)
	matches := 0
	
	for _, keyword := range keywords {
		if strings.Contains(sourceText, keyword) {
			matches++
		}
	}
	
	if len(keywords) == 0 {
		return 0.0
	}
	
	return float64(matches) / float64(len(keywords))
}

// sortByRelevance 按相关性排序
func (p *Processor) sortByRelevance(sources []KnowledgeItem) {
	// 这里实现排序逻辑
	// 可以使用sort.Slice进行排序
}

// calculateFusionScore 计算融合分数
func (p *Processor) calculateFusionScore(sources []KnowledgeItem) float64 {
	if len(sources) == 0 {
		return 0.0
	}
	
	// 计算平均相关性
	totalRelevance := 0.0
	for _, source := range sources {
		totalRelevance += source.Relevance
	}
	
	return totalRelevance / float64(len(sources))
}

// generateAnswer 生成回答
func (p *Processor) generateAnswer(ctx context.Context, fusionResult *FusionResult) (*Answer, error) {
	// 构建回答内容
	content := p.buildAnswerContent(fusionResult)
	
	// 生成摘要
	summary := p.buildAnswerSummary(content)
	
	// 计算置信度
	confidence := p.calculateConfidence(fusionResult)
	
	answer := &Answer{
		Content:    content,
		Summary:    summary,
		Sources:    fusionResult.Sources,
		Confidence: confidence,
	}

	return answer, nil
}

// buildAnswerContent 构建回答内容
func (p *Processor) buildAnswerContent(fusionResult *FusionResult) string {
	// 这里实现内容构建逻辑
	// 可以使用OpenAI API生成回答
	return "基于检索到的知识，这里是回答内容..."
}

// buildAnswerSummary 构建回答摘要
func (p *Processor) buildAnswerSummary(content string) string {
	// 这里实现摘要生成逻辑
	// 可以使用OpenAI API生成摘要
	if len(content) > 200 {
		return content[:200] + "..."
	}
	return content
}

// calculateConfidence 计算置信度
func (p *Processor) calculateConfidence(fusionResult *FusionResult) float64 {
	// 基于融合分数和来源质量计算置信度
	baseConfidence := fusionResult.FusionScore
	
	// 根据来源数量调整置信度
	sourceCount := len(fusionResult.Sources)
	if sourceCount > 0 {
		baseConfidence *= float64(sourceCount) / float64(p.config.Fusion.MaxSources)
	}
	
	return baseConfidence
}

// generateRecommendations 生成建议
func (p *Processor) generateRecommendations(question *Question, answer *Answer) []string {
	recommendations := make([]string, 0)
	
	// 基于问题类型生成建议
	switch question.Type {
	case QuestionTypeIssue:
		recommendations = append(recommendations, "建议在GitHub上创建Issue")
		recommendations = append(recommendations, "提供详细的错误信息和重现步骤")
	case QuestionTypePR:
		recommendations = append(recommendations, "确保代码符合项目规范")
		recommendations = append(recommendations, "添加必要的测试用例")
	case QuestionTypeConfig:
		recommendations = append(recommendations, "检查配置文件语法")
		recommendations = append(recommendations, "参考官方文档进行配置")
	}
	
	// 基于置信度生成建议
	if answer.Confidence < 0.7 {
		recommendations = append(recommendations, "建议提供更多上下文信息")
		recommendations = append(recommendations, "可以尝试重新描述问题")
	}
	
	return recommendations
}

// analyzeBug 分析Bug
func (p *Processor) analyzeBug(ctx context.Context, request *AnalyzeRequest) (*BugAnalysis, error) {
	// 获取Bug分析工具
	bugAnalyzer, ok := p.tools["bug_analyzer"].(*tools.BugAnalyzer)
	if !ok {
		return nil, fmt.Errorf("Bug分析工具未找到")
	}
	
	// 分析Bug
	analysis, err := bugAnalyzer.Analyze(ctx, request.StackTrace, request.Environment)
	if err != nil {
		return nil, err
	}
	
	return analysis, nil
}

// analyzeImage 分析图片
func (p *Processor) analyzeImage(ctx context.Context, request *AnalyzeRequest) (*ImageAnalysis, error) {
	// 获取图片分析工具
	imageAnalyzer, ok := p.tools["image_analyzer"].(*tools.ImageAnalyzer)
	if !ok {
		return nil, fmt.Errorf("图片分析工具未找到")
	}
	
	// 分析图片
	analysis, err := imageAnalyzer.Analyze(ctx, request.ImageURL)
	if err != nil {
		return nil, err
	}
	
	return analysis, nil
}

// classifyIssue 分类Issue
func (p *Processor) classifyIssue(ctx context.Context, request *AnalyzeRequest) (*IssueClassification, error) {
	// 获取Issue分类工具
	issueClassifier, ok := p.tools["issue_classifier"].(*tools.IssueClassifier)
	if !ok {
		return nil, fmt.Errorf("Issue分类工具未找到")
	}
	
	// 分类Issue
	classification, err := issueClassifier.Classify(ctx, request.StackTrace)
	if err != nil {
		return nil, err
	}
	
	return classification, nil
}

// analyzeGeneral 通用分析
func (p *Processor) analyzeGeneral(ctx context.Context, request *AnalyzeRequest) (interface{}, error) {
	// 通用分析逻辑
	return &BugAnalysis{
		ErrorType:  "unknown",
		Language:   "unknown",
		Severity:   "medium",
		RootCause:  "需要更多信息进行分析",
		Solutions:  []string{"请提供更多详细信息"},
		Prevention: []string{"定期检查系统状态"},
		Confidence: 0.3,
	}, nil
}

// SetLogger 设置日志器
func (p *Processor) SetLogger(logger *logrus.Logger) {
	p.logger = logger
} 