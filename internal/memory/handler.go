package memory

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler 记忆处理器
type Handler struct {
	manager *Manager
	logger  *logrus.Logger
}

// NewHandler 创建新的记忆处理器
func NewHandler(manager *Manager) *Handler {
	return &Handler{
		manager: manager,
		logger:  logrus.New(),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// 记忆管理API
	memory := router.Group("/api/v1/memory")
	{
		// 存储记忆
		memory.POST("/store", h.handleStoreMemory)

		// 检索记忆
		memory.POST("/retrieve", h.handleRetrieveMemory)

		// 获取记忆统计
		memory.GET("/stats/:session_id", h.handleGetMemoryStats)

		// 清除记忆
		memory.DELETE("/clear/:session_id", h.handleClearMemory)

		// 获取记忆列表
		memory.GET("/list/:session_id", h.handleListMemory)
	}
}

// handleStoreMemory 处理存储记忆请求
func (h *Handler) handleStoreMemory(c *gin.Context) {
	var request MemoryRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("解析存储记忆请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求格式错误",
			"message": err.Error(),
		})
		return
	}

	// 验证请求
	if err := h.validateMemoryRequest(&request); err != nil {
		h.logger.WithError(err).Error("验证存储记忆请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求验证失败",
			"message": err.Error(),
		})
		return
	}

	// 存储记忆
	ctx := c.Request.Context()
	if err := h.manager.StoreMemory(ctx, &request); err != nil {
		h.logger.WithError(err).Error("存储记忆失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "存储记忆失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "记忆存储成功",
		"session_id": request.SessionID,
		"type":       request.Type,
		"timestamp":  time.Now().Unix(),
	})
}

// handleRetrieveMemory 处理检索记忆请求
func (h *Handler) handleRetrieveMemory(c *gin.Context) {
	var query MemoryQuery

	if err := c.ShouldBindJSON(&query); err != nil {
		h.logger.WithError(err).Error("解析检索记忆请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求格式错误",
			"message": err.Error(),
		})
		return
	}

	// 验证请求
	if err := h.validateMemoryQuery(&query); err != nil {
		h.logger.WithError(err).Error("验证检索记忆请求失败")
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "请求验证失败",
			"message": err.Error(),
		})
		return
	}

	// 检索记忆
	ctx := c.Request.Context()
	response, err := h.manager.RetrieveMemory(ctx, &query)
	if err != nil {
		h.logger.WithError(err).Error("检索记忆失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "检索记忆失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleGetMemoryStats 处理获取记忆统计请求
func (h *Handler) handleGetMemoryStats(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID := c.Query("user_id")

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "会话ID不能为空",
		})
		return
	}

	// 获取记忆统计
	stats := h.manager.GetMemoryStats(sessionID, userID)

	c.JSON(http.StatusOK, stats)
}

// handleClearMemory 处理清除记忆请求
func (h *Handler) handleClearMemory(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID := c.Query("user_id")
	memoryType := MemoryType(c.Query("type"))

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "会话ID不能为空",
		})
		return
	}

	// 如果没有指定类型，清除所有记忆
	if memoryType == "" {
		// 清除工作记忆
		if err := h.manager.ClearMemory(sessionID, userID, WorkingMemory); err != nil {
			h.logger.WithError(err).Error("清除工作记忆失败")
		}

		// 清除短期记忆
		if err := h.manager.ClearMemory(sessionID, userID, ShortTermMemory); err != nil {
			h.logger.WithError(err).Error("清除短期记忆失败")
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "所有记忆已清除",
			"session_id": sessionID,
		})
		return
	}

	// 清除指定类型的记忆
	if err := h.manager.ClearMemory(sessionID, userID, memoryType); err != nil {
		h.logger.WithError(err).Error("清除记忆失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "清除记忆失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "记忆已清除",
		"session_id": sessionID,
		"type":       memoryType,
	})
}

// handleListMemory 处理获取记忆列表请求
func (h *Handler) handleListMemory(c *gin.Context) {
	sessionID := c.Param("session_id")
	userID := c.Query("user_id")
	memoryType := MemoryType(c.Query("type"))

	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "会话ID不能为空",
		})
		return
	}

	// 构建查询
	query := &MemoryQuery{
		SessionID: sessionID,
		UserID:    userID,
		Type:      memoryType,
		Limit:     50, // 默认限制50条
	}

	// 如果没有指定类型，获取所有类型的记忆
	if memoryType == "" {
		// 获取工作记忆
		workingResponse, err := h.manager.RetrieveMemory(c.Request.Context(), &MemoryQuery{
			SessionID: sessionID,
			UserID:    userID,
			Type:      WorkingMemory,
			Limit:     25,
		})
		if err != nil {
			h.logger.WithError(err).Error("获取工作记忆失败")
		}

		// 获取短期记忆
		shortTermResponse, err := h.manager.RetrieveMemory(c.Request.Context(), &MemoryQuery{
			SessionID: sessionID,
			UserID:    userID,
			Type:      ShortTermMemory,
			Limit:     25,
		})
		if err != nil {
			h.logger.WithError(err).Error("获取短期记忆失败")
		}

		// 合并结果
		var allItems []MemoryItem
		if workingResponse != nil {
			allItems = append(allItems, workingResponse.Items...)
		}
		if shortTermResponse != nil {
			allItems = append(allItems, shortTermResponse.Items...)
		}

		c.JSON(http.StatusOK, gin.H{
			"session_id": sessionID,
			"user_id":    userID,
			"items":      allItems,
			"count":      len(allItems),
			"working_count": func() int {
				if workingResponse != nil {
					return workingResponse.Count
				}
				return 0
			}(),
			"short_term_count": func() int {
				if shortTermResponse != nil {
					return shortTermResponse.Count
				}
				return 0
			}(),
		})
		return
	}

	// 获取指定类型的记忆
	response, err := h.manager.RetrieveMemory(c.Request.Context(), query)
	if err != nil {
		h.logger.WithError(err).Error("获取记忆列表失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "获取记忆列表失败",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// validateMemoryRequest 验证记忆请求
func (h *Handler) validateMemoryRequest(request *MemoryRequest) error {
	if request.SessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	if request.Content == "" {
		return fmt.Errorf("记忆内容不能为空")
	}

	if request.Type == "" {
		return fmt.Errorf("记忆类型不能为空")
	}

	// 验证记忆类型
	switch request.Type {
	case WorkingMemory, ShortTermMemory:
		// 有效类型
	default:
		return fmt.Errorf("不支持的记忆类型: %s", request.Type)
	}

	return nil
}

// validateMemoryQuery 验证记忆查询
func (h *Handler) validateMemoryQuery(query *MemoryQuery) error {
	if query.SessionID == "" {
		return fmt.Errorf("会话ID不能为空")
	}

	if query.Type == "" {
		return fmt.Errorf("记忆类型不能为空")
	}

	// 验证记忆类型
	switch query.Type {
	case WorkingMemory, ShortTermMemory:
		// 有效类型
	default:
		return fmt.Errorf("不支持的记忆类型: %s", query.Type)
	}

	return nil
}
