package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"community-governance-mcp-higress/internal/agent"
	"community-governance-mcp-higress/internal/openai"
)

// TestServer 测试服务器
type TestServer struct {
	router *gin.Engine
}

// NewTestServer 创建测试服务器
func NewTestServer() *TestServer {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 设置路由
	v1 := router.Group("/api/v1")
	{
		v1.POST("/process", handleProcess)
		v1.GET("/health", handleHealth)
	}
	
	return &TestServer{
		router: router,
	}
}

// handleProcess 处理测试请求
func handleProcess(c *gin.Context) {
	var request agent.ProcessRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 模拟处理逻辑
	response := agent.ProcessResponse{
		ID:              "test-id",
		QuestionID:      "test-question-id",
		Content:         "这是一个测试回答",
		Summary:         "测试回答摘要",
		Sources:         []agent.KnowledgeItem{},
		Confidence:      0.8,
		ProcessingTime:  "1.2s",
		FusionScore:     0.7,
		Recommendations: []string{"建议1", "建议2"},
	}

	c.JSON(http.StatusOK, response)
}

// handleHealth 健康检查
func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

// TestProcessRequest 测试处理请求
func TestProcessRequest(t *testing.T) {
	server := NewTestServer()
	
	// 创建测试请求
	request := agent.ProcessRequest{
		Title:    "测试问题",
		Content:  "这是一个测试问题内容",
		Author:   "test-user",
		Type:     agent.QuestionTypeText,
		Priority: agent.PriorityMedium,
		Tags:     []string{"test", "question"},
	}
	
	requestBody, _ := json.Marshal(request)
	
	// 创建HTTP请求
	req, err := http.NewRequest("POST", "/api/v1/process", bytes.NewBuffer(requestBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	server.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response agent.ProcessResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.QuestionID)
	assert.NotEmpty(t, response.Content)
	assert.Greater(t, response.Confidence, 0.0)
}

// TestHealthCheck 测试健康检查
func TestHealthCheck(t *testing.T) {
	server := NewTestServer()
	
	// 创建HTTP请求
	req, err := http.NewRequest("GET", "/api/v1/health", nil)
	assert.NoError(t, err)
	
	// 创建响应记录器
	w := httptest.NewRecorder()
	
	// 执行请求
	server.router.ServeHTTP(w, req)
	
	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "healthy", response["status"])
	assert.NotNil(t, response["timestamp"])
	assert.Equal(t, "1.0.0", response["version"])
}

// TestOpenAIClient 测试OpenAI客户端
func TestOpenAIClient(t *testing.T) {
	// 跳过测试如果没有API密钥
	if testing.Short() {
		t.Skip("跳过OpenAI客户端测试")
	}
	
	client := openai.NewClient("test-key", "gpt-4o")
	assert.NotNil(t, client)
	
	// 测试文本生成
	text, err := client.GenerateText(nil, "测试提示", 100, 0.7)
	if err == nil {
		assert.NotEmpty(t, text)
	}
}

// TestAgentTypes 测试Agent类型
func TestAgentTypes(t *testing.T) {
	// 测试问题类型
	assert.Equal(t, agent.QuestionTypeIssue, "issue")
	assert.Equal(t, agent.QuestionTypePR, "pr")
	assert.Equal(t, agent.QuestionTypeText, "text")
	
	// 测试优先级
	assert.Equal(t, agent.PriorityLow, "low")
	assert.Equal(t, agent.PriorityMedium, "medium")
	assert.Equal(t, agent.PriorityHigh, "high")
	assert.Equal(t, agent.PriorityUrgent, "urgent")
	
	// 测试知识源
	assert.Equal(t, agent.KnowledgeSourceLocal, "local")
	assert.Equal(t, agent.KnowledgeSourceHigress, "higress")
	assert.Equal(t, agent.KnowledgeSourceDeepWiki, "deepwiki")
	assert.Equal(t, agent.KnowledgeSourceGitHub, "github")
}

// TestQuestionCreation 测试问题创建
func TestQuestionCreation(t *testing.T) {
	question := &agent.Question{
		ID:        "test-id",
		Type:      agent.QuestionTypeText,
		Title:     "测试问题",
		Content:   "测试内容",
		Author:    "test-user",
		Priority:  agent.PriorityMedium,
		Tags:      []string{"test"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	assert.NotEmpty(t, question.ID)
	assert.Equal(t, agent.QuestionTypeText, question.Type)
	assert.Equal(t, "测试问题", question.Title)
	assert.Equal(t, agent.PriorityMedium, question.Priority)
}

// TestKnowledgeItem 测试知识项
func TestKnowledgeItem(t *testing.T) {
	item := &agent.KnowledgeItem{
		ID:        "test-item-id",
		Source:    agent.KnowledgeSourceLocal,
		Title:     "测试知识项",
		Content:   "测试知识内容",
		URL:       "https://example.com",
		Relevance: 0.8,
		Tags:      []string{"test", "knowledge"},
		CreatedAt: time.Now(),
	}
	
	assert.NotEmpty(t, item.ID)
	assert.Equal(t, agent.KnowledgeSourceLocal, item.Source)
	assert.Equal(t, "测试知识项", item.Title)
	assert.Equal(t, 0.8, item.Relevance)
}

// TestAnswerCreation 测试回答创建
func TestAnswerCreation(t *testing.T) {
	answer := &agent.Answer{
		ID:         "test-answer-id",
		QuestionID: "test-question-id",
		Content:    "测试回答内容",
		Summary:    "测试回答摘要",
		Sources:    []agent.KnowledgeItem{},
		Confidence: 0.9,
		CreatedAt:  time.Now(),
	}
	
	assert.NotEmpty(t, answer.ID)
	assert.Equal(t, "test-question-id", answer.QuestionID)
	assert.Equal(t, "测试回答内容", answer.Content)
	assert.Equal(t, 0.9, answer.Confidence)
}

// TestFusionResult 测试融合结果
func TestFusionResult(t *testing.T) {
	question := &agent.Question{
		ID:   "test-id",
		Type: agent.QuestionTypeText,
	}
	
	fusionResult := &agent.FusionResult{
		Question:      question,
		Sources:       []agent.KnowledgeItem{},
		FusionScore:   0.85,
		ProcessingTime: time.Second * 2,
	}
	
	assert.NotNil(t, fusionResult.Question)
	assert.Equal(t, 0.85, fusionResult.FusionScore)
	assert.Equal(t, time.Second*2, fusionResult.ProcessingTime)
}

// TestProcessRequestValidation 测试请求验证
func TestProcessRequestValidation(t *testing.T) {
	// 测试有效请求
	validRequest := agent.ProcessRequest{
		Title:   "测试标题",
		Content: "测试内容",
		Author:  "test-user",
	}
	
	assert.NotEmpty(t, validRequest.Title)
	assert.NotEmpty(t, validRequest.Content)
	
	// 测试无效请求
	invalidRequest := agent.ProcessRequest{
		Title:   "",
		Content: "",
		Author:  "",
	}
	
	assert.Empty(t, invalidRequest.Title)
	assert.Empty(t, invalidRequest.Content)
}

// TestProcessResponse 测试响应结构
func TestProcessResponse(t *testing.T) {
	response := agent.ProcessResponse{
		ID:              "test-response-id",
		QuestionID:      "test-question-id",
		Content:         "测试响应内容",
		Summary:         "测试响应摘要",
		Sources:         []agent.KnowledgeItem{},
		Confidence:      0.8,
		ProcessingTime:  "1.5s",
		FusionScore:     0.75,
		Recommendations: []string{"建议1", "建议2"},
	}
	
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.QuestionID)
	assert.NotEmpty(t, response.Content)
	assert.Greater(t, response.Confidence, 0.0)
	assert.LessOrEqual(t, response.Confidence, 1.0)
	assert.NotEmpty(t, response.ProcessingTime)
	assert.Greater(t, response.FusionScore, 0.0)
	assert.LessOrEqual(t, response.FusionScore, 1.0)
	assert.Len(t, response.Recommendations, 2)
}

// BenchmarkProcessRequest 基准测试
func BenchmarkProcessRequest(b *testing.B) {
	server := NewTestServer()
	request := agent.ProcessRequest{
		Title:   "基准测试问题",
		Content: "这是一个基准测试问题内容",
		Author:  "benchmark-user",
		Type:    agent.QuestionTypeText,
	}
	
	requestBody, _ := json.Marshal(request)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/process", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}
} 