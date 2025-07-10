package memory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Manager 记忆管理器实现
type Manager struct {
	workingMemories   map[string]*WorkingMemorySession
	shortTermMemories map[string]*ShortTermMemorySession
	config            MemoryConfig
	logger            *logrus.Logger
	mutex             sync.RWMutex
	cleanupTicker     *time.Ticker
	stopCleanup       chan bool
}

// NewManager 创建新的记忆管理器
func NewManager(config MemoryConfig) *Manager {
	manager := &Manager{
		workingMemories:   make(map[string]*WorkingMemorySession),
		shortTermMemories: make(map[string]*ShortTermMemorySession),
		config:            config,
		logger:            logrus.New(),
		stopCleanup:       make(chan bool),
	}

	// 启动清理协程
	go manager.startCleanupRoutine()

	return manager
}

// StoreMemory 存储记忆
func (m *Manager) StoreMemory(ctx context.Context, request *MemoryRequest) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	memoryItem := MemoryItem{
		ID:          uuid.New().String(),
		Type:        request.Type,
		Content:     request.Content,
		Context:     request.Context,
		Importance:  m.calculateImportance(request),
		AccessCount: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        request.Tags,
		Metadata:    request.Metadata,
	}

	// 根据记忆类型存储
	switch request.Type {
	case WorkingMemory:
		return m.storeWorkingMemory(request.SessionID, request.UserID, memoryItem)
	case ShortTermMemory:
		return m.storeShortTermMemory(request.SessionID, request.UserID, memoryItem)
	default:
		return fmt.Errorf("不支持的记忆类型: %s", request.Type)
	}
}

// RetrieveMemory 检索记忆
func (m *Manager) RetrieveMemory(ctx context.Context, query *MemoryQuery) (*MemoryResponse, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var items []MemoryItem

	switch query.Type {
	case WorkingMemory:
		items = m.retrieveWorkingMemory(query.SessionID, query.UserID, query)
	case ShortTermMemory:
		items = m.retrieveShortTermMemory(query.SessionID, query.UserID, query)
	default:
		return nil, fmt.Errorf("不支持的记忆类型: %s", query.Type)
	}

	// 按重要性排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Importance > items[j].Importance
	})

	// 限制返回数量
	if query.Limit > 0 && len(items) > query.Limit {
		items = items[:query.Limit]
	}

	// 更新访问次数
	for i := range items {
		items[i].AccessCount++
		items[i].UpdatedAt = time.Now()
	}

	// 生成上下文摘要
	contextSummary := m.generateContextSummary(items)

	return &MemoryResponse{
		SessionID: query.SessionID,
		UserID:    query.UserID,
		Type:      query.Type,
		Items:     items,
		Count:     len(items),
		Context:   contextSummary,
	}, nil
}

// GetMemoryStats 获取记忆统计
func (m *Manager) GetMemoryStats(sessionID, userID string) *MemoryStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	workingCount := 0
	shortTermCount := 0
	totalAccessCount := 0
	var lastAccess time.Time

	// 统计工作记忆
	if working, exists := m.workingMemories[sessionID]; exists {
		workingCount = len(working.Items)
		for _, item := range working.Items {
			totalAccessCount += item.AccessCount
			if item.UpdatedAt.After(lastAccess) {
				lastAccess = item.UpdatedAt
			}
		}
	}

	// 统计短期记忆
	if shortTerm, exists := m.shortTermMemories[sessionID]; exists {
		shortTermCount = len(shortTerm.Slots)
		for _, slot := range shortTerm.Slots {
			if slot.IsOccupied {
				totalAccessCount += slot.Item.AccessCount
				if slot.Item.UpdatedAt.After(lastAccess) {
					lastAccess = slot.Item.UpdatedAt
				}
			}
		}
	}

	// 计算内存使用率
	totalCapacity := m.config.WorkingMemoryMaxItems + m.config.ShortTermMemorySlots
	totalUsed := workingCount + shortTermCount
	memoryUsage := float64(totalUsed) / float64(totalCapacity)

	return &MemoryStats{
		SessionID:            sessionID,
		UserID:               userID,
		WorkingMemoryCount:   workingCount,
		ShortTermMemoryCount: shortTermCount,
		TotalAccessCount:     totalAccessCount,
		LastAccess:           lastAccess,
		MemoryUsage:          memoryUsage,
	}
}

// ClearMemory 清除记忆
func (m *Manager) ClearMemory(sessionID, userID string, memoryType MemoryType) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch memoryType {
	case WorkingMemory:
		delete(m.workingMemories, sessionID)
	case ShortTermMemory:
		delete(m.shortTermMemories, sessionID)
	default:
		return fmt.Errorf("不支持的记忆类型: %s", memoryType)
	}

	m.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"type":       memoryType,
	}).Info("记忆已清除")

	return nil
}

// storeWorkingMemory 存储工作记忆
func (m *Manager) storeWorkingMemory(sessionID, userID string, item MemoryItem) error {
	// 获取或创建工作记忆
	working, exists := m.workingMemories[sessionID]
	if !exists {
		working = &WorkingMemorySession{
			SessionID:  sessionID,
			UserID:     userID,
			Items:      make([]MemoryItem, 0),
			MaxItems:   m.config.WorkingMemoryMaxItems,
			TTL:        m.config.WorkingMemoryTTL,
			LastAccess: time.Now(),
		}
		m.workingMemories[sessionID] = working
	}

	// 检查是否已存在相同内容
	for i, existingItem := range working.Items {
		if existingItem.Content == item.Content {
			// 更新现有项
			working.Items[i].UpdatedAt = time.Now()
			working.Items[i].AccessCount++
			working.Items[i].Importance = item.Importance
			working.Items[i].Tags = item.Tags
			working.Items[i].Metadata = item.Metadata
			return nil
		}
	}

	// 添加新项
	working.Items = append(working.Items, item)
	working.LastAccess = time.Now()

	// 如果超出最大数量，移除最不重要的项
	if len(working.Items) > working.MaxItems {
		m.removeLeastImportantItems(working)
	}

	return nil
}

// storeShortTermMemory 存储短期记忆
func (m *Manager) storeShortTermMemory(sessionID, userID string, item MemoryItem) error {
	// 获取或创建短期记忆
	shortTerm, exists := m.shortTermMemories[sessionID]
	if !exists {
		shortTerm = &ShortTermMemorySession{
			SessionID:  sessionID,
			UserID:     userID,
			Slots:      make([]MemorySlot, m.config.ShortTermMemorySlots),
			MaxSlots:   m.config.ShortTermMemorySlots,
			TTL:        m.config.ShortTermMemoryTTL,
			LastAccess: time.Now(),
		}
		// 初始化槽位
		for i := range shortTerm.Slots {
			shortTerm.Slots[i] = MemorySlot{
				ID:         i,
				IsOccupied: false,
				Priority:   0,
				LastAccess: time.Now(),
			}
		}
		m.shortTermMemories[sessionID] = shortTerm
	}

	// 寻找空闲槽位或优先级最低的槽位
	var targetSlot *MemorySlot
	var minPriority = int(^uint(0) >> 1) // 最大整数

	for i := range shortTerm.Slots {
		slot := &shortTerm.Slots[i]
		if !slot.IsOccupied {
			targetSlot = slot
			break
		}
		if slot.Priority < minPriority {
			minPriority = slot.Priority
			targetSlot = slot
		}
	}

	if targetSlot != nil {
		// 计算优先级（基于重要性和访问次数）
		priority := int(item.Importance*100) + item.AccessCount

		targetSlot.Item = item
		targetSlot.IsOccupied = true
		targetSlot.Priority = priority
		targetSlot.LastAccess = time.Now()
		shortTerm.LastAccess = time.Now()
	}

	return nil
}

// retrieveWorkingMemory 检索工作记忆
func (m *Manager) retrieveWorkingMemory(sessionID, userID string, query *MemoryQuery) []MemoryItem {
	working, exists := m.workingMemories[sessionID]
	if !exists {
		return []MemoryItem{}
	}

	var items []MemoryItem
	for _, item := range working.Items {
		if m.matchesQuery(item, query) {
			items = append(items, item)
		}
	}

	return items
}

// retrieveShortTermMemory 检索短期记忆
func (m *Manager) retrieveShortTermMemory(sessionID, userID string, query *MemoryQuery) []MemoryItem {
	shortTerm, exists := m.shortTermMemories[sessionID]
	if !exists {
		return []MemoryItem{}
	}

	var items []MemoryItem
	for _, slot := range shortTerm.Slots {
		if slot.IsOccupied && m.matchesQuery(slot.Item, query) {
			items = append(items, slot.Item)
		}
	}

	return items
}

// matchesQuery 检查记忆项是否匹配查询条件
func (m *Manager) matchesQuery(item MemoryItem, query *MemoryQuery) bool {
	// 检查关键词
	if len(query.Keywords) > 0 {
		content := strings.ToLower(item.Content)
		context := strings.ToLower(item.Context)

		matched := false
		for _, keyword := range query.Keywords {
			if strings.Contains(content, strings.ToLower(keyword)) ||
				strings.Contains(context, strings.ToLower(keyword)) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查标签
	if len(query.Tags) > 0 {
		matched := false
		for _, queryTag := range query.Tags {
			for _, itemTag := range item.Tags {
				if queryTag == itemTag {
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查时间范围
	if query.Since != nil && item.CreatedAt.Before(*query.Since) {
		return false
	}

	return true
}

// calculateImportance 计算重要性评分
func (m *Manager) calculateImportance(request *MemoryRequest) float64 {
	importance := 0.5 // 基础重要性

	// 基于内容长度调整
	contentLength := len(request.Content)
	if contentLength > 1000 {
		importance += 0.2
	} else if contentLength > 500 {
		importance += 0.1
	}

	// 基于标签数量调整
	if len(request.Tags) > 0 {
		importance += float64(len(request.Tags)) * 0.05
	}

	// 基于元数据调整
	if metadata, exists := request.Metadata["priority"]; exists {
		if priority, ok := metadata.(string); ok {
			switch priority {
			case "high":
				importance += 0.3
			case "medium":
				importance += 0.1
			case "low":
				importance -= 0.1
			}
		}
	}

	// 确保重要性在0-1范围内
	if importance > 1.0 {
		importance = 1.0
	} else if importance < 0.0 {
		importance = 0.0
	}

	return importance
}

// removeLeastImportantItems 移除最不重要的项
func (m *Manager) removeLeastImportantItems(working *WorkingMemorySession) {
	// 按重要性排序
	sort.Slice(working.Items, func(i, j int) bool {
		return working.Items[i].Importance < working.Items[j].Importance
	})

	// 移除最不重要的项
	excess := len(working.Items) - working.MaxItems
	working.Items = working.Items[excess:]
}

// generateContextSummary 生成上下文摘要
func (m *Manager) generateContextSummary(items []MemoryItem) string {
	if len(items) == 0 {
		return ""
	}

	var summaries []string
	for _, item := range items {
		if item.Context != "" {
			summaries = append(summaries, item.Context)
		}
	}

	if len(summaries) == 0 {
		return ""
	}

	// 简单的摘要：取前3个上下文
	if len(summaries) > 3 {
		summaries = summaries[:3]
	}

	return strings.Join(summaries, "; ")
}

// startCleanupRoutine 启动清理协程
func (m *Manager) startCleanupRoutine() {
	m.cleanupTicker = time.NewTicker(m.config.CleanupInterval)
	defer m.cleanupTicker.Stop()

	for {
		select {
		case <-m.cleanupTicker.C:
			m.cleanupExpiredMemories()
		case <-m.stopCleanup:
			return
		}
	}
}

// cleanupExpiredMemories 清理过期的记忆
func (m *Manager) cleanupExpiredMemories() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()

	// 清理工作记忆
	for sessionID, working := range m.workingMemories {
		if now.Sub(working.LastAccess) > working.TTL {
			delete(m.workingMemories, sessionID)
			m.logger.WithField("session_id", sessionID).Info("清理过期的工作记忆")
		} else {
			// 清理过期的记忆项
			var validItems []MemoryItem
			for _, item := range working.Items {
				if item.ExpiresAt == nil || now.Before(*item.ExpiresAt) {
					validItems = append(validItems, item)
				}
			}
			working.Items = validItems
		}
	}

	// 清理短期记忆
	for sessionID, shortTerm := range m.shortTermMemories {
		if now.Sub(shortTerm.LastAccess) > shortTerm.TTL {
			delete(m.shortTermMemories, sessionID)
			m.logger.WithField("session_id", sessionID).Info("清理过期的短期记忆")
		} else {
			// 清理过期的槽位
			for i := range shortTerm.Slots {
				slot := &shortTerm.Slots[i]
				if slot.IsOccupied && slot.Item.ExpiresAt != nil && now.After(*slot.Item.ExpiresAt) {
					slot.IsOccupied = false
					slot.Priority = 0
				}
			}
		}
	}
}

// Stop 停止记忆管理器
func (m *Manager) Stop() {
	if m.cleanupTicker != nil {
		m.cleanupTicker.Stop()
	}
	close(m.stopCleanup)
}
