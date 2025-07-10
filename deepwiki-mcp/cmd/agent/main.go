package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"higress-mcp/internal/agent"
	"higress-mcp/internal/openai"
)

// Server HTTP服务器
type Server struct {
	processor *agent.Processor
	config    *agent.AgentConfig
	logger    *logrus.Logger
	router    *gin.Engine
}

// NewServer 创建新的服务器
func NewServer(processor *agent.Processor, config *agent.AgentConfig) *Server {
	server := &Server{
		processor: processor,
		config:    config,
		logger:    logrus.New(),
		router:    gin.Default(),
	}

	// 设置路由
	server.setupRoutes()

	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// API版本组
	v1 := s.router.Group("/api/v1")
	{
		// 处理问题
		v1.POST("/process", s.handleProcess)
		
		// 健康检查
		v1.GET("/health", s.handleHealth)
		
		// 获取配置信息
		v1.GET("/config", s.handleConfig)
	}

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

// handleHealth 健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   s.config.Version,
	})
}

// handleConfig 获取配置信息
func (s *Server) handleConfig(c *gin.Context) {
	// 返回安全的配置信息（不包含敏感数据）
	safeConfig := gin.H{
		"name":     s.config.Name,
		"version":  s.config.Version,
		"port":     s.config.Port,
		"debug":    s.config.Debug,
		"features": gin.H{
			"deepwiki_enabled": s.config.DeepWiki.Enabled,
			"knowledge_enabled": s.config.Knowledge.Enabled,
			"fusion_enabled":   s.config.Fusion.Enabled,
		},
	}
	
	c.JSON(http.StatusOK, safeConfig)
}

// handleRoot 根路径处理
func (s *Server) handleRoot(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Higress社区治理Agent",
		"version": s.config.Version,
		"docs":    "/api/v1/",
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

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	s.logger.WithField("port", s.config.Port).Info("启动HTTP服务器")
	
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
	
	// 从环境变量获取敏感信息
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		config.OpenAI.APIKey = apiKey
	}
	
	return &config, nil
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
	
	// 设置输出
	if config.Logging.Output == "file" && config.Logging.FilePath != "" {
		file, err := os.OpenFile(config.Logging.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logger.SetOutput(file)
		}
	}
	
	return logger
}

func main() {
	// 加载配置
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}
	
	// 设置日志
	logger := setupLogging(config)
	logger.Info("启动Higress社区治理Agent")
	
	// 创建OpenAI客户端
	openaiClient := openai.NewClient(config.OpenAI.APIKey)
	openaiClient.SetLogger(logger)
	
	// 创建处理器
	processor := agent.NewProcessor(openaiClient, config)
	processor.SetLogger(logger)
	
	// 创建服务器
	server := NewServer(processor, config)
	server.logger = logger
	
	// 设置优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			logger.WithError(err).Fatal("服务器启动失败")
		}
	}()
	
	// 等待关闭信号
	<-quit
	logger.Info("正在关闭服务器...")
	
	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 这里可以添加清理逻辑
	logger.Info("服务器已关闭")
}

// 示例请求和响应结构
type ExampleRequest struct {
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Author   string                 `json:"author"`
	Priority string                 `json:"priority"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

type ExampleResponse struct {
	ID              string                    `json:"id"`
	QuestionID      string                    `json:"question_id"`
	Content         string                    `json:"content"`
	Summary         string                    `json:"summary"`
	Sources         []agent.KnowledgeItem    `json:"sources"`
	Confidence      float64                   `json:"confidence"`
	ProcessingTime  string                    `json:"processing_time"`
	FusionScore     float64                   `json:"fusion_score"`
	Recommendations []string                  `json:"recommendations"`
} 