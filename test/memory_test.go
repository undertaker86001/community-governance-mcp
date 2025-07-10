package test

import (
	"context"
	"testing"
	"time"

	"github.com/community-governance-mcp-higress/internal/memory"
	"github.com/stretchr/testify/assert"
)

func TestMemoryManager(t *testing.T) {
	// 创建记忆配置
	config := memory.MemoryConfig{
		WorkingMemoryMaxItems: 10,
		WorkingMemoryTTL:      30 * time.Minute,
		ShortTermMemorySlots:  16,
		ShortTermMemoryTTL:    2 * time.Hour,
		CleanupInterval:       5 * time.Minute,
		ImportanceThreshold:   0.3,
	}

	// 创建记忆管理器
	manager := memory.NewManager(config)
	defer manager.Stop()

	ctx := context.Background()

	t.Run("测试工作记忆存储和检索", func(t *testing.T) {
		// 存储工作记忆
		request := &memory.MemoryRequest{
			SessionID: "test_session",
			UserID:    "test_user",
			Type:      memory.WorkingMemory,
			Content:   "这是一个测试问题",
			Context:   "测试上下文",
			Tags:      []string{"test", "question"},
			Metadata: map[string]interface{}{
				"priority": "high",
			},
		}

		err := manager.StoreMemory(ctx, request)
		assert.NoError(t, err)

		// 检索工作记忆
		query := &memory.MemoryQuery{
			SessionID: "test_session",
			UserID:    "test_user",
			Type:      memory.WorkingMemory,
			Keywords:  []string{"测试"},
			Limit:     5,
		}

		response, err := manager.RetrieveMemory(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, response.Items, 1)
		assert.Equal(t, "这是一个测试问题", response.Items[0].Content)
		assert.Equal(t, memory.WorkingMemory, response.Items[0].Type)
	})

	t.Run("测试短期记忆存储和检索", func(t *testing.T) {
		// 存储短期记忆
		request := &memory.MemoryRequest{
			SessionID: "test_session",
			UserID:    "test_user",
			Type:      memory.ShortTermMemory,
			Content:   "这是一个测试答案",
			Context:   "测试答案上下文",
			Tags:      []string{"test", "answer"},
			Metadata: map[string]interface{}{
				"confidence": 0.8,
			},
		}

		err := manager.StoreMemory(ctx, request)
		assert.NoError(t, err)

		// 检索短期记忆
		query := &memory.MemoryQuery{
			SessionID: "test_session",
			UserID:    "test_user",
			Type:      memory.ShortTermMemory,
			Keywords:  []string{"测试"},
			Limit:     5,
		}

		response, err := manager.RetrieveMemory(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, response.Items, 1)
		assert.Equal(t, "这是一个测试答案", response.Items[0].Content)
		assert.Equal(t, memory.ShortTermMemory, response.Items[0].Type)
	})

	t.Run("测试记忆统计", func(t *testing.T) {
		stats := manager.GetMemoryStats("test_session", "test_user")
		assert.NotNil(t, stats)
		assert.Equal(t, "test_session", stats.SessionID)
		assert.Equal(t, "test_user", stats.UserID)
		assert.Greater(t, stats.WorkingMemoryCount, 0)
		assert.Greater(t, stats.ShortTermMemoryCount, 0)
	})

	t.Run("测试记忆清除", func(t *testing.T) {
		// 清除工作记忆
		err := manager.ClearMemory("test_session", "test_user", memory.WorkingMemory)
		assert.NoError(t, err)

		// 验证工作记忆已被清除
		query := &memory.MemoryQuery{
			SessionID: "test_session",
			UserID:    "test_user",
			Type:      memory.WorkingMemory,
		}

		response, err := manager.RetrieveMemory(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, response.Items, 0)
	})

	t.Run("测试重要性计算", func(t *testing.T) {
		// 测试高优先级内容
		highPriorityRequest := &memory.MemoryRequest{
			SessionID: "test_session",
			UserID:    "test_user",
			Type:      memory.WorkingMemory,
			Content:   "这是一个非常重要的长内容，包含很多详细信息",
			Context:   "重要上下文",
			Tags:      []string{"important", "urgent", "critical"},
			Metadata: map[string]interface{}{
				"priority": "high",
			},
		}

		err := manager.StoreMemory(ctx, highPriorityRequest)
		assert.NoError(t, err)

		// 验证重要性评分
		query := &memory.MemoryQuery{
			SessionID: "test_session",
			UserID:    "test_user",
			Type:      memory.WorkingMemory,
		}

		response, err := manager.RetrieveMemory(ctx, query)
		assert.NoError(t, err)
		assert.Len(t, response.Items, 1)
		assert.Greater(t, response.Items[0].Importance, 0.7) // 高重要性
	})
}

func TestMemoryHandler(t *testing.T) {
	// 创建记忆配置
	config := memory.MemoryConfig{
		WorkingMemoryMaxItems: 10,
		WorkingMemoryTTL:      30 * time.Minute,
		ShortTermMemorySlots:  16,
		ShortTermMemoryTTL:    2 * time.Hour,
		CleanupInterval:       5 * time.Minute,
		ImportanceThreshold:   0.3,
	}

	// 创建记忆管理器
	manager := memory.NewManager(config)
	defer manager.Stop()

	// 创建记忆处理器
	handler := memory.NewHandler(manager)

	t.Run("测试记忆处理器创建", func(t *testing.T) {
		assert.NotNil(t, handler)
	})
}
