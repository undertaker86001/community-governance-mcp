package test

import (
	"bytes"
	"context"
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
	router := gin.Default()
	
	// 设置路由
	router.POST("/api/v1/process", handleProcess)
	router.POST("/api/v1/analyze", handleAnalyze)
	router.GET("/api/v1/health", handleHealth)
	
	return &TestServer{
		router: router,
	}
}

// handleProcess 处理问答请求
func handleProcess(c *gin.Context) {
	var request agent.ProcessRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 模拟处理
	response := agent.ProcessResponse{
		ID:              "test-id",
		QuestionID:      "test-question-id",
		Content:         "这是一个测试回答",
		Summary:         "测试摘要",
		Sources:         []agent.KnowledgeItem{},
		Confidence:      0.8,
		ProcessingTime:  "1s",
		FusionScore:     0.7,
		Recommendations: []string{"建议1", "建议2"},
	}

	c.JSON(http.StatusOK, response)
}

// handleAnalyze 处理分析请求
func handleAnalyze(c *gin.Context) {
	var request agent.AnalyzeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 模拟分析
	response := agent.AnalyzeResponse{
		ID:             "test-analysis-id",
		ProblemType:    "bug",
		Severity:       "medium",
		Diagnosis:      "测试诊断",
		Solutions:      []string{"解决方案1", "解决方案2"},
		Confidence:     0.7,
		ProcessingTime: "1s",
		RelatedIssues:  []string{"相关Issue1"},
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

// TestProcessQuestion 测试问答功能
func TestProcessQuestion(t *testing.T) {
	server := NewTestServer()

	request := agent.ProcessRequest{
		Title:   "测试问题",
		Content: "这是一个测试问题",
		Author:  "test-user",
		Type:    "question",
	}

	requestBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/process", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response agent.ProcessResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Content)
}

// TestAnalyzeProblem 测试问题分析功能
func TestAnalyzeProblem(t *testing.T) {
	server := NewTestServer()

	request := agent.AnalyzeRequest{
		StackTrace: "panic: runtime error: invalid memory address or nil pointer dereference",
		Environment: "Go 1.21, Linux",
		IssueType:  "bug",
	}

	requestBody, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response agent.AnalyzeResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "bug", response.ProblemType)
}

// TestHealthCheck 测试健康检查
func TestHealthCheck(t *testing.T) {
	server := NewTestServer()

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

// TestOpenAIClient 测试OpenAI客户端
func TestOpenAIClient(t *testing.T) {
	config := &agent.OpenAIConfig{
		APIKey:      "test-key",
		Model:       "gpt-4o",
		MaxTokens:   1000,
		Temperature: 0.7,
	}

	client := openai.NewClient(config)
	assert.NotNil(t, client)
}

// TestBugAnalyzer 测试Bug分析器
func TestBugAnalyzer(t *testing.T) {
	analyzer := NewBugAnalyzer()
	assert.NotNil(t, analyzer)

	// 测试空指针异常分析
	stackTrace := "panic: runtime error: invalid memory address or nil pointer dereference"
	environment := "Go 1.21, Linux"

	analysis, err := analyzer.Analyze(context.Background(), stackTrace, environment)
	assert.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "空指针异常", analysis.ErrorType)
	assert.Equal(t, "go", analysis.Language)
	assert.Equal(t, "critical", analysis.Severity)
}

// TestProcessor 测试处理器
func TestProcessor(t *testing.T) {
	config := &agent.AgentConfig{
		Name:    "test-agent",
		Version: "1.0.0",
		Port:    8080,
		Debug:   true,
		OpenAI: agent.OpenAIConfig{
			APIKey:      "test-key",
			Model:       "gpt-4o",
			MaxTokens:   1000,
			Temperature: 0.7,
		},
	}

	openaiClient := openai.NewClient(&config.OpenAI)
	processor := agent.NewProcessor(openaiClient, config)
	assert.NotNil(t, processor)

	// 测试工具注册
	processor.RegisterTool("test_tool", "test_value")
	tool, exists := processor.tools["test_tool"]
	assert.True(t, exists)
	assert.Equal(t, "test_value", tool)
}

// BenchmarkProcessQuestion 问答功能性能测试
func BenchmarkProcessQuestion(b *testing.B) {
	server := NewTestServer()
	request := agent.ProcessRequest{
		Title:   "性能测试问题",
		Content: "这是一个性能测试问题",
		Author:  "benchmark-user",
		Type:    "question",
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

// BenchmarkAnalyzeProblem 问题分析性能测试
func BenchmarkAnalyzeProblem(b *testing.B) {
	server := NewTestServer()
	request := agent.AnalyzeRequest{
		StackTrace: "panic: runtime error: invalid memory address or nil pointer dereference",
		Environment: "Go 1.21, Linux",
		IssueType:  "bug",
	}

	requestBody, _ := json.Marshal(request)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
	}
} 