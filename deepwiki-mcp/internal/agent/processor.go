package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"higress-mcp/internal/openai"
)

// Processor Agent处理器
type Processor struct {
	openaiClient *openai.Client
	logger       *logrus.Logger
	config       *AgentConfig
}

// NewProcessor 创建新的处理器
func NewProcessor(openaiClient *openai.Client, config *AgentConfig) *Processor {
	return &Processor{
		openaiClient: openaiClient,
		logger:       logrus.New(),
		config:       config,
	}
}

// ProcessQuestion 处理用户问题
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
		return request.Type
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
	higressKeywords := []string{"gateway", "plugin", "route", "config", "deployment"}
	for _, keyword := range higressKeywords {
		if strings.Contains(content, keyword) {
			tags = append(tags, keyword)
		}
	}

	// 技术标签
	techKeywords := []string{"kubernetes", "docker", "helm", "yaml", "json"}
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
		return request.Priority
	}

	// 基于关键词确定优先级
	content := strings.ToLower(request.Content + " " + request.Title)
	
	// 紧急关键词
	urgentKeywords := []string{"urgent", "critical", "broken", "crash", "production"}
	for _, keyword := range urgentKeywords {
		if strings.Contains(content, keyword) {
			return PriorityUrgent
		}
	}

	// 高优先级关键词
	highKeywords := []string{"important", "blocking", "security", "performance"}
	for _, keyword := range highKeywords {
		if strings.Contains(content, keyword) {
			return PriorityHigh
		}
	}

	// 中优先级关键词
	mediumKeywords := []string{"feature", "enhancement", "improvement"}
	for _, keyword := range mediumKeywords {
		if strings.Contains(content, keyword) {
			return PriorityMedium
		}
	}

	// 默认为低优先级
	return PriorityLow
}

// retrieveKnowledge 检索知识
func (p *Processor) retrieveKnowledge(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	var allSources []KnowledgeItem

	// 1. 从本地知识库检索
	if p.config.Knowledge.Enabled {
		localSources, err := p.retrieveLocalKnowledge(ctx, question)
		if err != nil {
			p.logger.WithError(err).Warn("本地知识库检索失败")
		} else {
			allSources = append(allSources, localSources...)
		}
	}

	// 2. 从Higress文档检索
	if p.config.Higress.DocsURL != "" {
		higressSources, err := p.retrieveHigressDocs(ctx, question)
		if err != nil {
			p.logger.WithError(err).Warn("Higress文档检索失败")
		} else {
			allSources = append(allSources, higressSources...)
		}
	}

	// 3. 从DeepWiki检索
	if p.config.DeepWiki.Enabled {
		deepwikiSources, err := p.retrieveDeepWiki(ctx, question)
		if err != nil {
			p.logger.WithError(err).Warn("DeepWiki检索失败")
		} else {
			allSources = append(allSources, deepwikiSources...)
		}
	}

	p.logger.WithField("sources_count", len(allSources)).Info("知识检索完成")
	return allSources, nil
}

// retrieveLocalKnowledge 检索本地知识库
func (p *Processor) retrieveLocalKnowledge(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	// TODO: 实现本地知识库检索
	// 这里可以集成向量数据库或本地文件系统
	return []KnowledgeItem{}, nil
}

// retrieveHigressDocs 检索Higress文档
func (p *Processor) retrieveHigressDocs(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	// TODO: 实现Higress文档检索
	// 这里可以爬取Higress官方文档或使用API
	return []KnowledgeItem{}, nil
}

// retrieveDeepWiki 检索DeepWiki
func (p *Processor) retrieveDeepWiki(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	// 构建DeepWiki问题
	repoName := "alibaba/higress" // Higress官方仓库
	questionText := fmt.Sprintf("%s: %s", question.Title, question.Content)

	// 调用DeepWiki
	answer, err := p.openaiClient.AskDeepWiki(ctx, repoName, questionText)
	if err != nil {
		return nil, fmt.Errorf("DeepWiki查询失败: %w", err)
	}

	// 构建知识项
	knowledgeItem := KnowledgeItem{
		ID:        uuid.New().String(),
		Source:    KnowledgeSourceDeepWiki,
		Title:     "DeepWiki回答",
		Content:   answer,
		URL:       fmt.Sprintf("https://github.com/%s", repoName),
		Relevance: 0.9, // DeepWiki回答通常相关性很高
		Tags:      question.Tags,
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"repo_name": repoName,
			"question":  questionText,
		},
	}

	return []KnowledgeItem{knowledgeItem}, nil
}

// fuseKnowledge 融合知识
func (p *Processor) fuseKnowledge(ctx context.Context, question *Question, sources []KnowledgeItem) (*FusionResult, error) {
	// 计算相关性分数
	for i := range sources {
		sources[i].Relevance = p.calculateRelevance(question, &sources[i])
	}

	// 按相关性排序
	p.sortByRelevance(sources)

	// 限制源数量
	if len(sources) > p.config.Fusion.MaxSources {
		sources = sources[:p.config.Fusion.MaxSources]
	}

	// 计算融合质量分数
	fusionScore := p.calculateFusionScore(sources)

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
	// TODO: 实现更复杂的语义相似度计算
	
	questionText := strings.ToLower(question.Title + " " + question.Content)
	sourceText := strings.ToLower(source.Title + " " + source.Content)

	// 计算关键词匹配度
	questionWords := strings.Fields(questionText)
	sourceWords := strings.Fields(sourceText)

	matches := 0
	for _, qWord := range questionWords {
		for _, sWord := range sourceWords {
			if qWord == sWord && len(qWord) > 2 {
				matches++
				break
			}
		}
	}

	if len(questionWords) == 0 {
		return 0.0
	}

	return float64(matches) / float64(len(questionWords))
}

// sortByRelevance 按相关性排序
func (p *Processor) sortByRelevance(sources []KnowledgeItem) {
	// 简单的冒泡排序
	for i := 0; i < len(sources)-1; i++ {
		for j := 0; j < len(sources)-i-1; j++ {
			if sources[j].Relevance < sources[j+1].Relevance {
				sources[j], sources[j+1] = sources[j+1], sources[j]
			}
		}
	}
}

// calculateFusionScore 计算融合质量分数
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
	summary := p.buildAnswerSummary(content)

	// 计算置信度
	confidence := p.calculateConfidence(fusionResult)

	answer := &Answer{
		ID:         uuid.New().String(),
		QuestionID: fusionResult.Question.ID,
		Content:    content,
		Summary:    summary,
		Sources:    fusionResult.Sources,
		Confidence: confidence,
		CreatedAt:  time.Now(),
		Metadata: map[string]interface{}{
			"fusion_score": fusionResult.FusionScore,
		},
	}

	return answer, nil
}

// buildAnswerContent 构建回答内容
func (p *Processor) buildAnswerContent(fusionResult *FusionResult) string {
	var content strings.Builder

	// 添加问题总结
	content.WriteString(fmt.Sprintf("## 问题分析\n\n"))
	content.WriteString(fmt.Sprintf("您的问题是关于 **%s** 的 %s 类型问题。\n\n", 
		fusionResult.Question.Title, fusionResult.Question.Type))

	// 添加主要回答
	content.WriteString(fmt.Sprintf("## 解决方案\n\n"))
	
	// 基于知识源生成回答
	for i, source := range fusionResult.Sources {
		if source.Relevance > p.config.Fusion.SimilarityThreshold {
			content.WriteString(fmt.Sprintf("### 来源 %d: %s\n\n", i+1, source.Title))
			content.WriteString(source.Content)
			content.WriteString("\n\n")
		}
	}

	// 添加建议
	content.WriteString(fmt.Sprintf("## 建议\n\n"))
	content.WriteString("- 请根据您的具体环境调整配置\n")
	content.WriteString("- 建议查看官方文档获取最新信息\n")
	content.WriteString("- 如有疑问，欢迎在社区讨论\n")

	return content.String()
}

// buildAnswerSummary 构建回答摘要
func (p *Processor) buildAnswerSummary(content string) string {
	// 简单的摘要生成：取前200个字符
	if len(content) <= 200 {
		return content
	}
	return content[:200] + "..."
}

// calculateConfidence 计算置信度
func (p *Processor) calculateConfidence(fusionResult *FusionResult) float64 {
	// 基于融合分数和源数量计算置信度
	baseConfidence := fusionResult.FusionScore
	
	// 根据源数量调整
	sourceCount := len(fusionResult.Sources)
	if sourceCount > 0 {
		baseConfidence += float64(sourceCount) * 0.1
	}
	
	// 限制在0-1范围内
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}
	
	return baseConfidence
}

// generateRecommendations 生成建议
func (p *Processor) generateRecommendations(question *Question, answer *Answer) []string {
	var recommendations []string

	// 基于问题类型生成建议
	switch question.Type {
	case QuestionTypeIssue:
		recommendations = append(recommendations, 
			"建议提供详细的错误日志和复现步骤",
			"检查是否是最新版本的问题",
			"考虑在GitHub上创建Issue")
	case QuestionTypePR:
		recommendations = append(recommendations,
			"确保代码符合项目规范",
			"添加必要的测试用例",
			"更新相关文档")
	case QuestionTypeText:
		recommendations = append(recommendations,
			"建议查看官方文档获取详细信息",
			"可以在社区论坛寻求帮助")
	}

	// 基于置信度生成建议
	if answer.Confidence < 0.7 {
		recommendations = append(recommendations,
			"建议提供更多上下文信息以获得更准确的回答")
	}

	return recommendations
}

// SetLogger 设置日志记录器
func (p *Processor) SetLogger(logger *logrus.Logger) {
	p.logger = logger
} 