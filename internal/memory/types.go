package memory

import (
	"time"
)

// MemoryType 记忆类型
type MemoryType string

const (
	WorkingMemory   MemoryType = "working"    // 工作记忆
	ShortTermMemory MemoryType = "short_term" // 短期记忆
)

// MemoryItem 记忆项
type MemoryItem struct {
	ID          string                 `json:"id"`           // 记忆项ID
	Type        MemoryType             `json:"type"`         // 记忆类型
	Content     string                 `json:"content"`      // 记忆内容
	Context     string                 `json:"context"`      // 上下文
	Importance  float64                `json:"importance"`   // 重要性评分 (0-1)
	AccessCount int                    `json:"access_count"` // 访问次数
	CreatedAt   time.Time              `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time              `json:"updated_at"`   // 更新时间
	ExpiresAt   *time.Time             `json:"expires_at"`   // 过期时间
	Tags        []string               `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
}

// WorkingMemorySession 工作记忆会话结构
type WorkingMemorySession struct {
	SessionID  string        `json:"session_id"`  // 会话ID
	UserID     string        `json:"user_id"`     // 用户ID
	Items      []MemoryItem  `json:"items"`       // 记忆项列表
	MaxItems   int           `json:"max_items"`   // 最大记忆项数量
	TTL        time.Duration `json:"ttl"`         // 生存时间
	LastAccess time.Time     `json:"last_access"` // 最后访问时间
}

// ShortTermMemorySession 短期记忆会话结构
type ShortTermMemorySession struct {
	SessionID  string        `json:"session_id"`  // 会话ID
	UserID     string        `json:"user_id"`     // 用户ID
	Slots      []MemorySlot  `json:"slots"`       // 记忆槽
	MaxSlots   int           `json:"max_slots"`   // 最大槽位数
	TTL        time.Duration `json:"ttl"`         // 生存时间
	LastAccess time.Time     `json:"last_access"` // 最后访问时间
}

// MemorySlot 记忆槽
type MemorySlot struct {
	ID         int        `json:"id"`          // 槽位ID
	Item       MemoryItem `json:"item"`        // 记忆项
	IsOccupied bool       `json:"is_occupied"` // 是否被占用
	Priority   int        `json:"priority"`    // 优先级
	LastAccess time.Time  `json:"last_access"` // 最后访问时间
}

// MemoryManager 记忆管理器
type MemoryManager struct {
	WorkingMemories   map[string]*WorkingMemorySession   `json:"working_memories"`    // 工作记忆映射
	ShortTermMemories map[string]*ShortTermMemorySession `json:"short_term_memories"` // 短期记忆映射
	Config            MemoryConfig                       `json:"config"`              // 配置
}

// MemoryConfig 记忆配置
type MemoryConfig struct {
	WorkingMemoryMaxItems int           `json:"working_memory_max_items"` // 工作记忆最大项数
	WorkingMemoryTTL      time.Duration `json:"working_memory_ttl"`       // 工作记忆生存时间
	ShortTermMemorySlots  int           `json:"short_term_memory_slots"`  // 短期记忆槽位数
	ShortTermMemoryTTL    time.Duration `json:"short_term_memory_ttl"`    // 短期记忆生存时间
	CleanupInterval       time.Duration `json:"cleanup_interval"`         // 清理间隔
	ImportanceThreshold   float64       `json:"importance_threshold"`     // 重要性阈值
}

// MemoryRequest 记忆请求
type MemoryRequest struct {
	SessionID string                 `json:"session_id"` // 会话ID
	UserID    string                 `json:"user_id"`    // 用户ID
	Type      MemoryType             `json:"type"`       // 记忆类型
	Content   string                 `json:"content"`    // 内容
	Context   string                 `json:"context"`    // 上下文
	Tags      []string               `json:"tags"`       // 标签
	Metadata  map[string]interface{} `json:"metadata"`   // 元数据
}

// MemoryResponse 记忆响应
type MemoryResponse struct {
	SessionID string       `json:"session_id"` // 会话ID
	UserID    string       `json:"user_id"`    // 用户ID
	Type      MemoryType   `json:"type"`       // 记忆类型
	Items     []MemoryItem `json:"items"`      // 记忆项
	Count     int          `json:"count"`      // 记忆项数量
	Context   string       `json:"context"`    // 上下文摘要
}

// MemoryQuery 记忆查询
type MemoryQuery struct {
	SessionID string     `json:"session_id"` // 会话ID
	UserID    string     `json:"user_id"`    // 用户ID
	Type      MemoryType `json:"type"`       // 记忆类型
	Keywords  []string   `json:"keywords"`   // 关键词
	Tags      []string   `json:"tags"`       // 标签
	Limit     int        `json:"limit"`      // 限制数量
	Since     *time.Time `json:"since"`      // 起始时间
}

// MemoryStats 记忆统计
type MemoryStats struct {
	SessionID            string    `json:"session_id"`              // 会话ID
	UserID               string    `json:"user_id"`                 // 用户ID
	WorkingMemoryCount   int       `json:"working_memory_count"`    // 工作记忆数量
	ShortTermMemoryCount int       `json:"short_term_memory_count"` // 短期记忆数量
	TotalAccessCount     int       `json:"total_access_count"`      // 总访问次数
	LastAccess           time.Time `json:"last_access"`             // 最后访问时间
	MemoryUsage          float64   `json:"memory_usage"`            // 内存使用率
}
