package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/community-governance-mcp-higress/internal/memory"
	"github.com/community-governance-mcp-higress/internal/openai"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"encoding/json"
	"io"
	"github.com/community-governance-mcp-higress/tools"
	"github.com/community-governance-mcp-higress/internal/mcp"
	"github.com/community-governance-mcp-higress/internal/model"
)

// Processor 处理器
type Processor struct {
	openaiClient    *openai.Client
	config          *model.AgentConfig
	logger          *logrus.Logger
	mcpManager      *mcp.Manager
	retrievalManager *RetrievalManager
	memoryManager   *memory.Manager
	fallbackStrategy *FallbackStrategy
}

// NewProcessor 创建新的处理器
func NewProcessor(openaiClient *openai.Client, config *model.AgentConfig) *Processor {
	// 创建记忆管理器
	memoryConfig := memory.MemoryConfig{
		WorkingMemoryMaxItems: config.Memory.WorkingMemoryMaxItems,
		WorkingMemoryTTL:      config.Memory.WorkingMemoryTTL,
		ShortTermMemorySlots:  config.Memory.ShortTermMemorySlots,
		ShortTermMemoryTTL:    config.Memory.ShortTermMemoryTTL,
		CleanupInterval:       config.Memory.CleanupInterval,
		ImportanceThreshold:   config.Memory.ImportanceThreshold,
	}
	memoryManager := memory.NewManager(memoryConfig)

	// 创建检索管理器
	retrievalManager := NewRetrievalManager(&config.Network)

	// 创建备用策略
	fallbackStrategy := NewFallbackStrategy()

	// 创建MCP管理器
	mcpManager := mcp.NewManager(&config.MCP)

	// 创建处理器
	processor := &Processor{
		openaiClient:    openaiClient,
		config:          config,
		logger:          logrus.New(),
		mcpManager:      mcpManager,
		retrievalManager: retrievalManager,
		memoryManager:   memoryManager,
		fallbackStrategy: fallbackStrategy,
	}

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	processor.logger.SetLevel(level)

	return processor
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

	// 0. 检索相关记忆
	relatedMemories, err := p.retrieveRelatedMemories(ctx, request)
	if err != nil {
		p.logger.WithError(err).Warn("检索记忆失败，继续处理")
	}

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

	// 3. 知识融合（包含记忆）
	fusionResult, err := p.fuseKnowledgeWithMemory(ctx, question, sources, relatedMemories)
	if err != nil {
		return nil, fmt.Errorf("知识融合失败: %w", err)
	}

	// 4. 生成回答
	answer, err := p.generateAnswer(ctx, fusionResult)
	if err != nil {
		return nil, fmt.Errorf("生成回答失败: %w", err)
	}

	// 5. 存储相关记忆
	p.storeRelevantMemories(ctx, request, question, answer)

	// 6. 构建响应
	processingTime := time.Since(startTime)
	response := &ProcessResponse{
		ID:              uuid.New().String(),
		QuestionID:      questionID,
		Content:         answer.Content,
		Summary:         answer.Summary,
		Sources:         answer.Sources,
		Confidence:      answer.Confidence,
		ProcessingTime:  processingTime.String(),
		FusionScore:     fusionResult.FusionScore,
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
		"analysis_id":     analysisID,
		"processing_time": processingTime,
		"confidence":      response.Confidence,
	}).Info("问题分析完成")

	return response, nil
}

// retrieveRelatedMemories 检索相关记忆
func (p *Processor) retrieveRelatedMemories(ctx context.Context, request *ProcessRequest) ([]memory.MemoryItem, error) {
	// 生成会话ID（基于用户ID）
	sessionID := fmt.Sprintf("session_%s", request.Author)

	// 构建查询
	query := &memory.MemoryQuery{
		SessionID: sessionID,
		UserID:    request.Author,
		Keywords:  p.extractKeywords(request.Content),
		Tags:      request.Tags,
		Limit:     10,
	}

	// 检索工作记忆
	workingResponse, err := p.memoryManager.RetrieveMemory(ctx, &memory.MemoryQuery{
		SessionID: sessionID,
		UserID:    request.Author,
		Type:      memory.WorkingMemory,
		Keywords:  query.Keywords,
		Tags:      query.Tags,
		Limit:     5,
	})
	if err != nil {
		return nil, fmt.Errorf("检索工作记忆失败: %w", err)
	}

	// 检索短期记忆
	shortTermResponse, err := p.memoryManager.RetrieveMemory(ctx, &memory.MemoryQuery{
		SessionID: sessionID,
		UserID:    request.Author,
		Type:      memory.ShortTermMemory,
		Keywords:  query.Keywords,
		Tags:      query.Tags,
		Limit:     5,
	})
	if err != nil {
		return nil, fmt.Errorf("检索短期记忆失败: %w", err)
	}

	// 合并记忆项
	var allMemories []memory.MemoryItem
	allMemories = append(allMemories, workingResponse.Items...)
	allMemories = append(allMemories, shortTermResponse.Items...)

	p.logger.WithField("memory_count", len(allMemories)).Info("检索到相关记忆")
	return allMemories, nil
}

// fuseKnowledgeWithMemory 融合知识和记忆
func (p *Processor) fuseKnowledgeWithMemory(ctx context.Context, question *Question, sources []KnowledgeItem, memories []memory.MemoryItem) (*FusionResult, error) {
	// 原有的知识融合逻辑
	fusionResult, err := p.fuseKnowledge(ctx, question, sources)
	if err != nil {
		return nil, err
	}

	// 如果有相关记忆，将其添加到融合结果中
	if len(memories) > 0 {
		memoryContext := p.buildMemoryContext(memories)
		fusionResult.Context += "\n\n相关历史记忆:\n" + memoryContext

		p.logger.WithField("memory_items", len(memories)).Info("融合记忆到知识中")
	}

	return fusionResult, nil
}

// storeRelevantMemories 存储相关记忆
func (p *Processor) storeRelevantMemories(ctx context.Context, request *ProcessRequest, question *Question, answer *Answer) {
	sessionID := fmt.Sprintf("session_%s", request.Author)

	// 存储问题到工作记忆
	questionMemory := &memory.MemoryRequest{
		SessionID: sessionID,
		UserID:    request.Author,
		Type:      memory.WorkingMemory,
		Content:   request.Content,
		Context:   fmt.Sprintf("问题类型: %s, 优先级: %s", request.Type, request.Priority),
		Tags:      request.Tags,
		Metadata: map[string]interface{}{
			"question_id": question.ID,
			"priority":    request.Priority,
			"type":        request.Type,
		},
	}

	if err := p.memoryManager.StoreMemory(ctx, questionMemory); err != nil {
		p.logger.WithError(err).Warn("存储问题记忆失败")
	}

	// 存储答案到短期记忆
	answerMemory := &memory.MemoryRequest{
		SessionID: sessionID,
		UserID:    request.Author,
		Type:      memory.ShortTermMemory,
		Content:   answer.Content,
		Context:   fmt.Sprintf("回答置信度: %.2f, 融合分数: %.2f", answer.Confidence, answer.FusionScore),
		Tags:      append(request.Tags, "answer"),
		Metadata: map[string]interface{}{
			"question_id":  question.ID,
			"confidence":   answer.Confidence,
			"fusion_score": answer.FusionScore,
		},
	}

	if err := p.memoryManager.StoreMemory(ctx, answerMemory); err != nil {
		p.logger.WithError(err).Warn("存储答案记忆失败")
	}
}

// extractKeywords 提取关键词
func (p *Processor) extractKeywords(content string) []string {
	// 简单的关键词提取逻辑
	// 这里可以集成更复杂的NLP处理
	words := strings.Fields(content)
	var keywords []string

	for _, word := range words {
		if len(word) > 3 && !p.isCommonWord(word) {
			keywords = append(keywords, strings.ToLower(word))
		}
	}

	return keywords
}

// isCommonWord 判断是否为常见词
func (p *Processor) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true, "at": true,
		"to": true, "for": true, "of": true, "with": true, "by": true, "from": true, "this": true,
		"that": true, "is": true, "are": true, "was": true, "were": true, "be": true, "been": true,
		"have": true, "has": true, "had": true, "do": true, "does": true, "did": true, "will": true,
		"would": true, "could": true, "should": true, "can": true, "may": true, "might": true,
	}

	return commonWords[strings.ToLower(word)]
}

// buildMemoryContext 构建记忆上下文
func (p *Processor) buildMemoryContext(memories []memory.MemoryItem) string {
	if len(memories) == 0 {
		return ""
	}

	var contextParts []string
	for i, memory := range memories {
		if i >= 3 { // 最多显示3个记忆
			break
		}
		contextParts = append(contextParts, fmt.Sprintf("- %s (重要性: %.2f)", memory.Content, memory.Importance))
	}

	return strings.Join(contextParts, "\n")
}

// GetCommunityStats 获取社区统计
func (p *Processor) GetCommunityStats(ctx context.Context) (*CommunityStats, error) {
	return nil, fmt.Errorf("社区统计功能未实现")
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
	highKeywords := []string{"important", "blocking", "issue", "error"}
	for _, keyword := range highKeywords {
		if strings.Contains(content, keyword) {
			return PriorityHigh
		}
	}

	// 中优先级关键词
	mediumKeywords := []string{"question", "help", "how", "what"}
	for _, keyword := range mediumKeywords {
		if strings.Contains(content, keyword) {
			return PriorityMedium
		}
	}

	return PriorityLow
}

// retrieveKnowledge 检索知识
func (p *Processor) retrieveKnowledge(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	var allSources []KnowledgeItem

	// 1. 检索本地知识库
	if p.config.Knowledge.Enabled {
		localSources, err := p.retrieveLocalKnowledge(ctx, question)
		if err != nil {
			p.logger.WithError(err).Warn("本地知识库检索失败")
		} else {
			allSources = append(allSources, localSources...)
		}
	}

	// 2. 检索Higress文档
	higressSources, err := p.retrieveHigressDocs(ctx, question)
	if err != nil {
		p.logger.WithError(err).Warn("Higress文档检索失败")
	} else {
		allSources = append(allSources, higressSources...)
	}

	// 3. 检索DeepWiki
	if p.config.DeepWiki.Enabled {
		deepwikiSources, err := p.retrieveDeepWiki(ctx, question)
		if err != nil {
			p.logger.WithError(err).Warn("DeepWiki检索失败")
		} else {
			allSources = append(allSources, deepwikiSources...)
		}
	}

	// 4. 计算相关性并排序
	for i := range allSources {
		allSources[i].Relevance = p.calculateRelevance(question, &allSources[i])
	}
	p.sortByRelevance(allSources)

	// 限制返回数量
	if len(allSources) > p.config.Fusion.MaxSources {
		allSources = allSources[:p.config.Fusion.MaxSources]
	}

	p.logger.WithField("sources_count", len(allSources)).Info("知识检索完成")
	return allSources, nil
}

// retrieveLocalKnowledge 检索本地知识库
func (p *Processor) retrieveLocalKnowledge(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	p.logger.Info("开始检索本地知识库")
	
	// 使用现有的知识库工具
	knowledgeBase := tools.NewKnowledgeBase(p.config.OpenAI.APIKey)
	
	// 构建查询
	query := question.Title + " " + question.Content
	
	// 执行搜索
	searchResult, err := knowledgeBase.SearchKnowledge(query, 5)
	if err != nil {
		p.logger.WithError(err).Warn("本地知识库检索失败")
		return []KnowledgeItem{}, nil // 返回空结果而不是错误
	}
	
	// 转换为KnowledgeItem
	var items []KnowledgeItem
	for _, result := range searchResult.Results {
		item := KnowledgeItem{
			ID:        result.DocumentID,
			Source:    KnowledgeSourceLocal,
			Title:     result.Title,
			Content:   result.Content,
			URL:       "", // 本地知识库没有URL
			Relevance: result.RelevanceScore,
			Tags:      []string{}, // 可以从文档内容中提取标签
			CreatedAt: time.Now(),
			Metadata: map[string]interface{}{
				"snippet": result.Snippet,
			},
		}
		items = append(items, item)
	}
	
	p.logger.WithField("results_count", len(items)).Info("本地知识库检索完成")
	return items, nil
}

// retrieveHigressDocs 检索Higress文档
func (p *Processor) retrieveHigressDocs(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	p.logger.Info("开始检索Higress文档")
	
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	
	// 使用多个备用API端点，避免网络限制
	endpoints := []string{
		"https://higress.io/docs",
		"https://higress.cn/docs", 
		"https://api.github.com/repos/alibaba/higress/contents/docs",
	}
	
	// 使用多端点检索
	multiRetrieval := NewMultiEndpointRetrieval(endpoints, DefaultRetrievalConfig())
	result, err := multiRetrieval.Retrieve(timeoutCtx, p.retrievalManager)
	
	if err == nil && result.Success {
		// 解析响应内容
		items, err := p.parseHigressResponse(strings.NewReader(string(result.Data)), question)
		if err == nil && len(items) > 0 {
			p.logger.WithField("endpoint", "multi").Info("Higress文档检索成功")
			return items, nil
		}
	}
	
	// 如果所有端点都失败，使用本地缓存或模拟数据
	fallbackItems := p.fallbackStrategy.GetHigressFallbackData()
	items := p.convertFallbackToKnowledgeItems(question, fallbackItems, KnowledgeSourceHigress)
	p.logger.Info("使用Higress文档备用数据")
	return items, nil
}

// convertFallbackToKnowledgeItems 将备用数据转换为知识项
func (p *Processor) convertFallbackToKnowledgeItems(question *Question, fallbackData map[string]string, source KnowledgeSource) []KnowledgeItem {
	var items []KnowledgeItem
	query := strings.ToLower(question.Title + " " + question.Content)
	
	for keyword, content := range fallbackData {
		if strings.Contains(query, strings.ToLower(keyword)) {
			item := KnowledgeItem{
				ID:        fmt.Sprintf("fallback_%s", keyword),
				Source:    source,
				Title:     fmt.Sprintf("%s指南", keyword),
				Content:   content,
				URL:       "", // 备用数据没有URL
				Relevance: 0.8, // 较高的相关性
				Tags:      []string{string(source), keyword},
				CreatedAt: time.Now(),
				Metadata: map[string]interface{}{
					"source": "fallback_cache",
				},
			}
			items = append(items, item)
		}
	}
	
	return items
}

// parseHigressResponse 解析Higress响应
func (p *Processor) parseHigressResponse(body io.Reader, question *Question) ([]KnowledgeItem, error) {
	// 读取响应内容
	content, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}
	
	// 简单的文本解析（实际项目中可以使用更复杂的解析）
	text := string(content)
	
	// 提取相关内容片段
	snippets := p.extractRelevantSnippets(text, question.Title+" "+question.Content)
	
	var items []KnowledgeItem
	for i, snippet := range snippets {
		item := KnowledgeItem{
			ID:        fmt.Sprintf("higress_%d", i),
			Source:    KnowledgeSourceHigress,
			Title:     fmt.Sprintf("Higress文档片段 %d", i+1),
			Content:   snippet,
			URL:       "https://higress.io/docs",
			Relevance: p.calculateRelevance(question, &KnowledgeItem{Content: snippet}),
			Tags:      []string{"higress", "documentation"},
			CreatedAt: time.Now(),
			Metadata: map[string]interface{}{
				"source": "higress_docs",
			},
		}
		items = append(items, item)
	}
	
	return items, nil
}

// extractRelevantSnippets 提取相关内容片段
func (p *Processor) extractRelevantSnippets(text, query string) []string {
	query = strings.ToLower(query)
	text = strings.ToLower(text)
	
	// 简单的关键词匹配
	words := strings.Fields(query)
	var snippets []string
	
	// 按段落分割
	paragraphs := strings.Split(text, "\n\n")
	
	for _, paragraph := range paragraphs {
		if len(paragraph) < 50 { // 忽略太短的段落
			continue
		}
		
		// 检查是否包含查询关键词
		matches := 0
		for _, word := range words {
			if len(word) < 3 {
				continue
			}
			if strings.Contains(paragraph, word) {
				matches++
			}
		}
		
		// 如果匹配度足够高，添加到结果中
		if float64(matches)/float64(len(words)) > 0.3 {
			snippets = append(snippets, paragraph)
		}
		
		// 限制结果数量
		if len(snippets) >= 5 {
			break
		}
	}
	
	return snippets
}

// retrieveDeepWiki 检索DeepWiki
func (p *Processor) retrieveDeepWiki(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	p.logger.Info("开始检索DeepWiki")
	
	// 创建带超时的上下文
	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	
	// 检查DeepWiki配置
	if !p.config.DeepWiki.Enabled {
		p.logger.Info("DeepWiki未启用，跳过检索")
		return []KnowledgeItem{}, nil
	}
	
	// 使用MCP管理器进行查询
	items, err := p.mcpManager.QueryWithFallback(
		timeoutCtx,
		"deepwiki",
		question.Title+" "+question.Content,
		"modelcontextprotocol/modelcontextprotocol", // 默认仓库，可根据需要调整
		func() ([]model.KnowledgeItem, error) {
			// 备用方案：直接HTTP调用
			httpItems, err := p.retrieveFromDeepWikiHTTP(timeoutCtx, question)
			if err != nil {
				// 如果HTTP调用也失败，使用备用数据
				fallbackData := p.fallbackStrategy.GetDeepWikiFallbackData()
				items := p.convertFallbackToKnowledgeItems(question, fallbackData, KnowledgeSourceDeepWiki)
				return items, nil
			}
			return httpItems, nil
		},
	)
	
	if err != nil {
		p.logger.WithError(err).Warn("DeepWiki检索失败，使用备用数据")
		fallbackData := p.fallbackStrategy.GetDeepWikiFallbackData()
		items = p.convertFallbackToKnowledgeItems(question, fallbackData, KnowledgeSourceDeepWiki)
	}
	
	p.logger.WithField("results_count", len(items)).Info("DeepWiki检索完成")
	return items, nil
}

// retrieveFromDeepWikiHTTP 通过HTTP调用检索DeepWiki
func (p *Processor) retrieveFromDeepWikiHTTP(ctx context.Context, question *Question) ([]KnowledgeItem, error) {
	// 构建HTTP客户端
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			DisableCompression:  true,
		},
	}
	
	// 构建请求URL
	baseURL := p.config.DeepWiki.Endpoint
	if baseURL == "" {
		baseURL = "https://api.deepwiki.com" // 默认端点
	}
	
	query := url.QueryEscape(question.Title + " " + question.Content)
	requestURL := fmt.Sprintf("%s/search?q=%s&limit=5", baseURL, query)
	
	// 构建请求
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	
	// 设置请求头
	req.Header.Set("User-Agent", "HigressBot/1.0")
	req.Header.Set("Accept", "application/json")
	if p.config.DeepWiki.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.config.DeepWiki.APIKey)
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DeepWiki API返回错误: %d", resp.StatusCode)
	}
	
	// 解析JSON响应
	var response struct {
		Results []struct {
			Title   string  `json:"title"`
			Content string  `json:"content"`
			URL     string  `json:"url"`
			Score   float64 `json:"score"`
		} `json:"results"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	
	// 转换为KnowledgeItem
	var items []KnowledgeItem
	for _, result := range response.Results {
		item := KnowledgeItem{
			ID:        fmt.Sprintf("deepwiki_%s", result.Title),
			Source:    KnowledgeSourceDeepWiki,
			Title:     result.Title,
			Content:   result.Content,
			URL:       result.URL,
			Relevance: result.Score,
			Tags:      []string{"deepwiki"},
			CreatedAt: time.Now(),
			Metadata: map[string]interface{}{
				"source": "deepwiki_api",
			},
		}
		items = append(items, item)
	}
	
	return items, nil
}

// fuseKnowledge 融合知识
func (p *Processor) fuseKnowledge(ctx context.Context, question *Question, sources []KnowledgeItem) (*FusionResult, error) {
	// 计算融合分数
	fusionScore := p.calculateFusionScore(sources)
	// 构建融合结果
	fResult := &FusionResult{
		Sources:     sources,
		FusionScore: fusionScore,
		Context:     "", // 可根据需要补充上下文
	}
	p.logger.WithFields(logrus.Fields{
		"fusion_score":  fusionScore,
		"sources_count": len(sources),
	}).Info("知识融合完成")
	return fResult, nil
}

// calculateRelevance 计算相关性
func (p *Processor) calculateRelevance(question *Question, source *KnowledgeItem) float64 {
	// 简单的关键词匹配算法
	questionText := strings.ToLower(question.Title + " " + question.Content)
	sourceText := strings.ToLower(source.Title + " " + source.Content)

	// 计算关键词匹配度
	questionWords := strings.Fields(questionText)
	sourceWords := strings.Fields(sourceText)

	matches := 0
	for _, qWord := range questionWords {
		if len(qWord) < 3 { // 忽略短词
			continue
		}
		for _, sWord := range sourceWords {
			if strings.Contains(sWord, qWord) || strings.Contains(qWord, sWord) {
				matches++
				break
			}
		}
	}

	if len(questionWords) == 0 {
		return 0.0
	}

	relevance := float64(matches) / float64(len(questionWords))

	// 标签匹配加分
	for _, qTag := range question.Tags {
		for _, sTag := range source.Tags {
			if strings.EqualFold(qTag, sTag) {
				relevance += 0.2
				break
			}
		}
	}

	// 确保分数在0-1之间
	if relevance > 1.0 {
		relevance = 1.0
	}

	return relevance
}

// sortByRelevance 按相关性排序
func (p *Processor) sortByRelevance(sources []KnowledgeItem) {
	// 简单的冒泡排序，按相关性降序
	for i := 0; i < len(sources)-1; i++ {
		for j := 0; j < len(sources)-1-i; j++ {
			if sources[j].Relevance < sources[j+1].Relevance {
				sources[j], sources[j+1] = sources[j+1], sources[j]
			}
		}
	}
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

	avgRelevance := totalRelevance / float64(len(sources))

	// 考虑来源多样性
	diversityBonus := 0.0
	sourceTypes := make(map[KnowledgeSource]bool)
	for _, source := range sources {
		sourceTypes[source.Source] = true
	}

	if len(sourceTypes) > 1 {
		diversityBonus = 0.1 * float64(len(sourceTypes)-1)
	}

	fusionScore := avgRelevance + diversityBonus
	if fusionScore > 1.0 {
		fusionScore = 1.0
	}

	return fusionScore
}

// generateAnswer 生成回答
func (p *Processor) generateAnswer(ctx context.Context, fusionResult *FusionResult) (*Answer, error) {
	content := p.buildAnswerContent(fusionResult)
	summary := p.buildAnswerSummary(content)
	confidence := p.calculateConfidence(fusionResult)
	answer := &Answer{
		Content:     content,
		Summary:     summary,
		Sources:     fusionResult.Sources,
		Confidence:  confidence,
		FusionScore: fusionResult.FusionScore,
	}
	return answer, nil
}

// buildAnswerContent 构建回答内容
func (p *Processor) buildAnswerContent(fusionResult *FusionResult) string {
	if len(fusionResult.Sources) == 0 {
		return "抱歉，我没有找到相关的信息来回答您的问题。请尝试重新描述您的问题，或者联系社区管理员获取帮助。"
	}

	var content strings.Builder

	// 添加主要回答
	content.WriteString("根据我的分析，以下是针对您问题的回答：\n\n")

	// 基于最相关的源构建回答
	if len(fusionResult.Sources) > 0 {
		primarySource := fusionResult.Sources[0]
		content.WriteString(primarySource.Content)
		content.WriteString("\n\n")
	}

	// 如果有多个来源，添加补充信息
	if len(fusionResult.Sources) > 1 {
		content.WriteString("补充信息：\n")
		for i, source := range fusionResult.Sources[1:] {
			if i >= 2 { // 最多显示3个来源
				break
			}
			content.WriteString(fmt.Sprintf("- %s\n", source.Title))
		}
		content.WriteString("\n")
	}

	// 添加来源链接
	content.WriteString("参考来源：\n")
	for _, source := range fusionResult.Sources {
		if source.URL != "" {
			content.WriteString(fmt.Sprintf("- [%s](%s)\n", source.Title, source.URL))
		}
	}

	return content.String()
}

// buildAnswerSummary 构建回答摘要
func (p *Processor) buildAnswerSummary(content string) string {
	// 简单的摘要生成：取前200个字符
	if len(content) <= 200 {
		return content
	}

	summary := content[:200]
	// 尝试在句号处截断
	if lastPeriod := strings.LastIndex(summary, "。"); lastPeriod > 150 {
		summary = summary[:lastPeriod+3] // 包含句号
	}

	return summary + "..."
}

// calculateConfidence 计算置信度
func (p *Processor) calculateConfidence(fusionResult *FusionResult) float64 {
	// 基于融合分数和来源质量计算置信度
	confidence := fusionResult.FusionScore

	// 根据来源数量调整
	if len(fusionResult.Sources) > 0 {
		avgRelevance := 0.0
		for _, source := range fusionResult.Sources {
			avgRelevance += source.Relevance
		}
		avgRelevance /= float64(len(fusionResult.Sources))

		// 融合分数和平均相关性的加权平均
		confidence = (fusionResult.FusionScore*0.7 + avgRelevance*0.3)
	}

	// 确保置信度在0-1之间
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateRecommendations 生成建议
func (p *Processor) generateRecommendations(question *Question, answer *Answer) []string {
	var recommendations []string

	// 基于问题类型生成建议
	switch question.Type {
	case QuestionTypeIssue:
		recommendations = append(recommendations,
			"如果问题仍然存在，请提供更详细的错误信息和环境配置",
			"考虑在GitHub上创建Issue以获得更多社区支持")
	case QuestionTypePR:
		recommendations = append(recommendations,
			"确保您的PR符合项目的贡献指南",
			"添加适当的测试用例和文档更新")
	case QuestionTypeText:
		recommendations = append(recommendations,
			"如果这个回答对您有帮助，请标记为已解决",
			"您可以在社区论坛中分享您的经验")
	}

	// 基于置信度生成建议
	if answer.Confidence < 0.7 {
		recommendations = append(recommendations,
			"这个回答的置信度较低，建议您进一步验证信息",
			"考虑联系项目维护者获取更准确的指导")
	}

	// 基于标签生成建议
	for _, tag := range question.Tags {
		switch tag {
		case "gateway":
			recommendations = append(recommendations, "查看Higress网关配置文档")
		case "plugin":
			recommendations = append(recommendations, "参考插件开发指南")
		case "kubernetes":
			recommendations = append(recommendations, "确保您的Kubernetes环境配置正确")
		}
	}

	return recommendations
}

// analyzeBug 分析Bug
func (p *Processor) analyzeBug(ctx context.Context, request *AnalyzeRequest) (*BugAnalysis, error) {
	return &BugAnalysis{
		Severity:   "high",
		RootCause:  "模拟根因",
		Solutions:  []string{"模拟解决方案1", "模拟解决方案2"},
		Prevention: []string{"模拟预防建议"},
		Confidence: 0.9,
	}, nil
}

// analyzeImage 分析图片
func (p *Processor) analyzeImage(ctx context.Context, request *AnalyzeRequest) (*ImageAnalysis, error) {
	return &ImageAnalysis{
		ErrorMessages: []string{"模拟图片错误"},
		Suggestions:   []string{"模拟图片建议"},
		Confidence:    0.8,
	}, nil
}

// classifyIssue 分类Issue
func (p *Processor) classifyIssue(ctx context.Context, request *AnalyzeRequest) (*IssueClassification, error) {
	return &IssueClassification{
		Category:   "bug",
		Priority:   "high",
		Assignees:  []string{"user1"},
		Confidence: 0.85,
	}, nil
}

// analyzeGeneral 通用分析
func (p *Processor) analyzeGeneral(ctx context.Context, request *AnalyzeRequest) (interface{}, error) {
	return &BugAnalysis{
		Severity:   "medium",
		RootCause:  "模拟一般问题根因",
		Solutions:  []string{"模拟一般问题解决方案"},
		Prevention: []string{"模拟一般问题预防"},
		Confidence: 0.7,
	}, nil
}

// SetLogger 设置日志器
func (p *Processor) SetLogger(logger *logrus.Logger) {
	p.logger = logger
}

// GetMemoryManager 获取记忆管理器
func (p *Processor) GetMemoryManager() *memory.Manager {
	return p.memoryManager
}
