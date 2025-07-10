package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/community-governance-mcp-higress/internal/model"
	"github.com/community-governance-mcp-higress/internal/openai"
)

// KnowledgeBase 知识库工具
type KnowledgeBase struct {
	openaiClient *openai.Client
	documents    []model.Document
}

// NewKnowledgeBase 创建新的知识库
func NewKnowledgeBase(apiKey string) *KnowledgeBase {
	return &KnowledgeBase{
		openaiClient: openai.NewClient(apiKey, "gpt-4o"),
		documents:    []model.Document{},
	}
}

// AddDocument 添加文档到知识库
func (kb *KnowledgeBase) AddDocument(doc model.Document) {
	kb.documents = append(kb.documents, doc)
}

// SearchKnowledge 搜索知识库
func (kb *KnowledgeBase) SearchKnowledge(query string, maxResults int) (*model.KnowledgeSearchResult, error) {
	if len(kb.documents) == 0 {
		return &model.KnowledgeSearchResult{
			Query:     query,
			Results:   []model.SearchResult{},
			TotalHits: 0,
		}, nil
	}

	// 使用AI进行语义搜索
	results, err := kb.semanticSearch(query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("语义搜索失败: %w", err)
	}

	return &model.KnowledgeSearchResult{
		Query:     query,
		Results:   results,
		TotalHits: len(results),
	}, nil
}

// semanticSearch 语义搜索
func (kb *KnowledgeBase) semanticSearch(query string, maxResults int) ([]model.SearchResult, error) {
	// 构建搜索提示
	prompt := kb.buildSearchPrompt(query, maxResults)

	// 使用AI进行搜索
	response, err := kb.openaiClient.GenerateText(context.Background(), prompt, 800, 0.3)
	if err != nil {
		return nil, fmt.Errorf("AI搜索失败: %w", err)
	}

	// 解析搜索结果
	results := kb.parseSearchResults(response, query)

	// 限制结果数量
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}

// buildSearchPrompt 构建搜索提示
func (kb *KnowledgeBase) buildSearchPrompt(query string, maxResults int) string {
	// 构建文档内容
	var docContent strings.Builder
	for i, doc := range kb.documents {
		docContent.WriteString(fmt.Sprintf("文档%d:\n标题: %s\n内容: %s\n\n", i+1, doc.Title, doc.Content))
	}

	return fmt.Sprintf(`请从以下知识库文档中搜索与查询最相关的内容：

查询: %s

知识库文档:
%s

请返回最相关的%d个结果，格式如下：
{
  "results": [
    {
      "document_id": "文档ID",
      "title": "文档标题",
      "content": "相关内容片段",
      "relevance_score": 0.95,
      "snippet": "匹配的文本片段"
    }
  ]
}`, query, docContent.String(), maxResults)
}

// parseSearchResults 解析搜索结果
func (kb *KnowledgeBase) parseSearchResults(response string, query string) []model.SearchResult {
	var results []model.SearchResult

	// 尝试解析JSON响应
	if strings.Contains(response, "{") && strings.Contains(response, "}") {
		start := strings.Index(response, "{")
		end := strings.LastIndex(response, "}") + 1
		if start >= 0 && end > start {
			jsonStr := response[start:end]

			var result map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
				if resultsArray, ok := result["results"].([]interface{}); ok {
					for _, item := range resultsArray {
						if resultMap, ok := item.(map[string]interface{}); ok {
							searchResult := model.SearchResult{
								DocumentID:     getStringFromMap(resultMap, "document_id"),
								Title:          getStringFromMap(resultMap, "title"),
								Content:        getStringFromMap(resultMap, "content"),
								RelevanceScore: getFloatFromMap(resultMap, "relevance_score"),
								Snippet:        getStringFromMap(resultMap, "snippet"),
							}
							results = append(results, searchResult)
						}
					}
				}
			}
		}
	}

	// 如果没有解析到JSON，使用简单的文本匹配
	if len(results) == 0 {
		results = kb.fallbackTextSearch(query)
	}

	return results
}

// fallbackTextSearch 备用文本搜索
func (kb *KnowledgeBase) fallbackTextSearch(query string) []model.SearchResult {
	var results []model.SearchResult
	query = strings.ToLower(query)

	for i, doc := range kb.documents {
		content := strings.ToLower(doc.Content)
		title := strings.ToLower(doc.Title)

		// 简单的关键词匹配
		relevance := 0.0
		if strings.Contains(content, query) || strings.Contains(title, query) {
			relevance = 0.8
		}

		// 检查部分匹配
		words := strings.Fields(query)
		for _, word := range words {
			if strings.Contains(content, word) || strings.Contains(title, word) {
				relevance += 0.2
			}
		}

		if relevance > 0.0 {
			// 生成片段
			snippet := kb.generateSnippet(doc.Content, query)

			results = append(results, model.SearchResult{
				DocumentID:     fmt.Sprintf("doc_%d", i),
				Title:          doc.Title,
				Content:        doc.Content,
				RelevanceScore: relevance,
				Snippet:        snippet,
			})
		}
	}

	return results
}

// generateSnippet 生成文本片段
func (kb *KnowledgeBase) generateSnippet(content string, query string) string {
	// 简单的片段生成
	words := strings.Fields(query)
	for _, word := range words {
		if idx := strings.Index(strings.ToLower(content), strings.ToLower(word)); idx != -1 {
			start := idx - 50
			if start < 0 {
				start = 0
			}
			end := idx + 100
			if end > len(content) {
				end = len(content)
			}
			return content[start:end] + "..."
		}
	}

	// 如果没有找到匹配，返回前100个字符
	if len(content) > 100 {
		return content[:100] + "..."
	}
	return content
}

// GetDocument 获取文档
func (kb *KnowledgeBase) GetDocument(documentID string) (*model.Document, error) {
	for _, doc := range kb.documents {
		if doc.ID == documentID {
			return &doc, nil
		}
	}
	return nil, fmt.Errorf("文档未找到: %s", documentID)
}

// UpdateDocument 更新文档
func (kb *KnowledgeBase) UpdateDocument(documentID string, updates model.Document) error {
	for i, doc := range kb.documents {
		if doc.ID == documentID {
			kb.documents[i] = updates
			return nil
		}
	}
	return fmt.Errorf("文档未找到: %s", documentID)
}

// DeleteDocument 删除文档
func (kb *KnowledgeBase) DeleteDocument(documentID string) error {
	for i, doc := range kb.documents {
		if doc.ID == documentID {
			kb.documents = append(kb.documents[:i], kb.documents[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("文档未找到: %s", documentID)
}

// GetDocumentCount 获取文档数量
func (kb *KnowledgeBase) GetDocumentCount() int {
	return len(kb.documents)
}

// GetDocuments 获取所有文档
func (kb *KnowledgeBase) GetDocuments() []model.Document {
	return kb.documents
}

// ClearDocuments 清空所有文档
func (kb *KnowledgeBase) ClearDocuments() {
	kb.documents = []model.Document{}
}

// ExportDocuments 导出文档
func (kb *KnowledgeBase) ExportDocuments() ([]byte, error) {
	return json.Marshal(kb.documents)
}

// ImportDocuments 导入文档
func (kb *KnowledgeBase) ImportDocuments(data []byte) error {
	var documents []model.Document
	if err := json.Unmarshal(data, &documents); err != nil {
		return fmt.Errorf("解析文档数据失败: %w", err)
	}
	kb.documents = documents
	return nil
}

// getStringFromMap 安全获取字符串值
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// getFloatFromMap 安全获取浮点数值
func getFloatFromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	return 0.0
}
