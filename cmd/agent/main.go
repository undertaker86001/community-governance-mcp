package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/community-governance-mcp-higress/internal/agent"
	"github.com/community-governance-mcp-higress/internal/memory"
	"github.com/community-governance-mcp-higress/internal/openai"
	"github.com/community-governance-mcp-higress/internal/mcp"
	"github.com/community-governance-mcp-higress/tools"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Server HTTP服务器
type Server struct {
	processor     *agent.Processor
	memoryHandler *memory.Handler
	config        *agent.AgentConfig
	logger        *logrus.Logger
	router        *gin.Engine
}

// NewServer 创建新的服务器
func NewServer(processor *agent.Processor, config *agent.AgentConfig) *Server {
	server := &Server{
		processor: processor,
		config:    config,
		logger:    logrus.New(),
		router:    gin.Default(),
	}

	// 创建记忆处理器
	server.memoryHandler = memory.NewHandler(processor.GetMemoryManager())

	// 设置路由
	server.setupRoutes()

	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// API版本组
	v1 := s.router.Group("/api/v1")
	{
		// 核心功能路由
		v1.POST("/process", s.handleProcess)
		v1.POST("/analyze", s.handleAnalyze)
		v1.GET("/stats", s.handleStats)
		v1.GET("/health", s.handleHealth)
		v1.GET("/config", s.handleConfig)

		// MCP集成路由
		mcp := v1.Group("/mcp")
		{
			mcp.POST("/query", s.handleMCPQuery)
			mcp.POST("/tools", s.handleMCPListTools)
			mcp.POST("/call", s.handleMCPCallTool)
		}
	}

	// 注册记忆组件路由
	s.memoryHandler.RegisterRoutes(s.router)

	// 根路径
	s.router.GET("/", s.handleRoot)
}

// handleProcess 处理问题请求
func (s *Server) handleProcess(c *gin.Context) {
	var request agent.ProcessRequest

	// 解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		s.logger.WithError(err).Error("请求解析失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求格式错误",
			"message": err.Error(),
		})
		return
	}

	// 验证请求
	if err := s.validateRequest(&request); err != nil {
		s.logger.WithError(err).Error("请求验证失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求验证失败",
			"message": err.Error(),
		})
		return
	}

	// 处理问题
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	response, err := s.processor.ProcessQuestion(ctx, &request)
	if err != nil {
		s.logger.WithError(err).Error("问题处理失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "问题处理失败",
			"message": err.Error(),
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, response)
}

// handleAnalyze 处理问题分析请求
func (s *Server) handleAnalyze(c *gin.Context) {
	var request agent.AnalyzeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	analysis := &agent.AnalyzeResponse{
		ID:          uuid.New().String(),
		ProblemType: request.IssueType,
		Diagnosis:   "问题分析中...",
		Solutions:   []string{},
		Confidence:  0.0,
	}

	// 使用Bug分析器
	bugAnalyzer := tools.NewBugAnalyzer(s.config.OpenAI.APIKey)
	bugAnalysis, err := bugAnalyzer.AnalyzeBug(request.Content, request.StackTrace, "production")
	if err == nil {
		analysis.Diagnosis = bugAnalysis.RootCause
		analysis.Solutions = bugAnalysis.Solutions
		analysis.Confidence = 0.8 // 设置默认置信度
	}

	// 如果有图片，使用图片分析器
	if request.ImageURL != "" {
		imageAnalyzer := tools.NewImageAnalyzer(s.config.OpenAI.APIKey)
		imageAnalysis, err := imageAnalyzer.AnalyzeImage(request.ImageURL)
		if err == nil {
			// 将图片分析结果合并到诊断中
			if len(imageAnalysis.Issues) > 0 {
				analysis.Diagnosis += "\n图片分析发现的问题: " + strings.Join(imageAnalysis.Issues, ", ")
			}
			if len(imageAnalysis.Suggestions) > 0 {
				analysis.Solutions = append(analysis.Solutions, imageAnalysis.Suggestions...)
			}
		}
	}

	c.JSON(http.StatusOK, analysis)
}

// handleStats 处理社区统计请求
func (s *Server) handleStats(c *gin.Context) {
	// 获取查询参数
	period := c.DefaultQuery("period", "30d")
	repoOwner := c.DefaultQuery("owner", s.config.Higress.RepoOwner)
	repoName := c.DefaultQuery("repo", s.config.Higress.RepoName)

	// 使用社区统计工具
	statsTool := tools.NewCommunityStats(s.config.GitHub.Token)
	stats, err := statsTool.GetCommunityStats(repoOwner, repoName, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取统计信息失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleHealth 健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   s.config.Agent.Version,
		"services": gin.H{
			"openai":    "connected",
			"deepwiki":  s.config.DeepWiki.Enabled,
			"knowledge": s.config.Knowledge.Enabled,
		},
	})
}

// handleConfig 获取配置信息
func (s *Server) handleConfig(c *gin.Context) {
	// 返回安全的配置信息（不包含敏感数据）
	safeConfig := gin.H{
		"name":    s.config.Agent.Name,
		"version": s.config.Agent.Version,
		"port":    s.config.Agent.Port,
		"debug":   s.config.Agent.Debug,
		"features": gin.H{
			"deepwiki_enabled":  s.config.DeepWiki.Enabled,
			"knowledge_enabled": s.config.Knowledge.Enabled,
			"fusion_enabled":    s.config.Fusion.Enabled,
		},
	}

	c.JSON(http.StatusOK, safeConfig)
}

// handleRoot 根路径处理
func (s *Server) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Higress社区治理Agent",
		"version": s.config.Agent.Version,
		"docs":    "/api/v1/",
		"features": []string{
			"智能问答",
			"问题分析",
			"社区统计",
			"知识融合",
		},
	})
}

// validateRequest 验证请求
func (s *Server) validateRequest(request *agent.ProcessRequest) error {
	if request.Title == "" {
		return fmt.Errorf("标题不能为空")
	}

	if request.Content == "" {
		return fmt.Errorf("内容不能为空")
	}

	if request.Author == "" {
		request.Author = "anonymous"
	}

	return nil
}

// handleMCPQuery 处理MCP查询请求
func (s *Server) handleMCPQuery(c *gin.Context) {
	var request mcp.QueryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 创建MCP客户端
	mcpClient := mcp.NewClient(30 * time.Second)

	// 执行查询
	response, err := mcpClient.Query(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "MCP查询失败",
			"message": err.Error(),
		})
		return
	}

	if response.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": response.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"output": response.Output,
	})
}

// handleMCPListTools 处理MCP工具列表请求
func (s *Server) handleMCPListTools(c *gin.Context) {
	var request mcp.ListToolsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 创建MCP客户端
	mcpClient := mcp.NewClient(30 * time.Second)

	// 获取工具列表
	response, err := mcpClient.ListTools(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取工具列表失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": response.Tools,
	})
}

// handleMCPCallTool 处理MCP工具调用请求
func (s *Server) handleMCPCallTool(c *gin.Context) {
	var request mcp.CallToolRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求参数错误",
			"message": err.Error(),
		})
		return
	}

	// 创建MCP客户端
	mcpClient := mcp.NewClient(30 * time.Second)

	// 调用工具
	response, err := mcpClient.CallTool(c.Request.Context(), &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "工具调用失败",
			"message": err.Error(),
		})
		return
	}

	if response.Error != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": response.Error,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"output": response.Output,
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Agent.Port)
	s.logger.WithField("port", s.config.Agent.Port).Info("启动HTTP服务器")

	return s.router.Run(addr)
}

// loadConfig 加载配置
func loadConfig() (*agent.AgentConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置环境变量
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var config agent.AgentConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 手动解析时间字段
	if err := parseTimeFields(&config); err != nil {
		return nil, fmt.Errorf("解析时间字段失败: %w", err)
	}

	// 从环境变量获取敏感信息
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.OpenAI.APIKey = apiKey
	}

	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		config.GitHub.Token = githubToken
	}

	return &config, nil
}

// parseTimeFields 解析时间字段
func parseTimeFields(config *agent.AgentConfig) error {
	// 解析记忆配置中的时间字段
	if config.Memory.WorkingMemoryTTL == 0 {
		config.Memory.WorkingMemoryTTL = 30 * time.Minute
	}
	if config.Memory.ShortTermMemoryTTL == 0 {
		config.Memory.ShortTermMemoryTTL = 2 * time.Hour
	}
	if config.Memory.CleanupInterval == 0 {
		config.Memory.CleanupInterval = 5 * time.Minute
	}

	return nil
}

// setupLogging 设置日志
func setupLogging(config *agent.AgentConfig) *logrus.Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Logging.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if config.Logging.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return logger
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 设置日志
	logger := setupLogging(config)
	logger.Info("启动Higress社区治理Agent")

	// 创建OpenAI客户端
	openaiClient := openai.NewClient(config.OpenAI.APIKey, config.OpenAI.Model)

	// 创建处理器
	processor := agent.NewProcessor(openaiClient, config)

	// 创建服务器
	server := NewServer(processor, config)

	// 启动HTTP服务器
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("服务器启动失败:", err)
		}
	}()

	// 等待信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("正在关闭服务器...")
}
