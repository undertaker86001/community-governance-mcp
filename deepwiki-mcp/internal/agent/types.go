package agent

import (
	"time"
)

// QuestionType 问题类型枚举
type QuestionType string

const (
	QuestionTypeIssue    QuestionType = "issue"    // GitHub Issue
	QuestionTypePR       QuestionType = "pr"       // Pull Request
	QuestionTypeText     QuestionType = "text"     // 图文问题
	QuestionTypeUnknown  QuestionType = "unknown"  // 未知类型
)

// Priority 优先级枚举
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// KnowledgeSource 知识源类型
type KnowledgeSource string

const (
	KnowledgeSourceLocal    KnowledgeSource = "local"     // 本地知识库
	KnowledgeSourceHigress  KnowledgeSource = "higress"   // Higress文档
	KnowledgeSourceDeepWiki KnowledgeSource = "deepwiki"  // DeepWiki
)

// Question 问题结构体
type Question struct {
	ID          string       `json:"id"`           // 问题唯一标识
	Type        QuestionType `json:"type"`         // 问题类型
	Title       string       `json:"title"`        // 问题标题
	Content     string       `json:"content"`      // 问题内容
	Author      string       `json:"author"`       // 提问者
	Priority    Priority     `json:"priority"`     // 优先级
	Tags        []string     `json:"tags"`         // 标签
	CreatedAt   time.Time    `json:"created_at"`   // 创建时间
	UpdatedAt   time.Time    `json:"updated_at"`   // 更新时间
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
}

// KnowledgeItem 知识项结构体
type KnowledgeItem struct {
	ID          string            `json:"id"`           // 知识项ID
	Source      KnowledgeSource   `json:"source"`       // 知识源
	Title       string            `json:"title"`        // 标题
	Content     string            `json:"content"`      // 内容
	URL         string            `json:"url"`          // 来源URL
	Relevance   float64           `json:"relevance"`    // 相关性分数
	Tags        []string          `json:"tags"`         // 标签
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
}

// Answer 回答结构体
type Answer struct {
	ID          string            `json:"id"`           // 回答ID
	QuestionID  string            `json:"question_id"`  // 对应问题ID
	Content     string            `json:"content"`      // 回答内容
	Summary     string            `json:"summary"`      // 回答摘要
	Sources     []KnowledgeItem   `json:"sources"`      // 知识来源
	Confidence  float64           `json:"confidence"`   // 置信度
	CreatedAt   time.Time         `json:"created_at"`   // 创建时间
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
}

// FusionResult 知识融合结果
type FusionResult struct {
	Question    *Question         `json:"question"`     // 原始问题
	Answer      *Answer           `json:"answer"`       // 融合后的回答
	Sources     []KnowledgeItem   `json:"sources"`      // 所有知识源
	FusionScore float64           `json:"fusion_score"` // 融合质量分数
	ProcessingTime time.Duration  `json:"processing_time"` // 处理时间
}

// ProcessRequest 处理请求结构体
type ProcessRequest struct {
	Type        QuestionType     `json:"type"`         // 问题类型
	Title       string           `json:"title"`        // 问题标题
	Content     string           `json:"content"`      // 问题内容
	Author      string           `json:"author"`       // 提问者
	Priority    Priority         `json:"priority"`     // 优先级
	Tags        []string         `json:"tags"`         // 标签
	Metadata    map[string]interface{} `json:"metadata"` // 元数据
}

// ProcessResponse 处理响应结构体
type ProcessResponse struct {
	ID              string        `json:"id"`               // 响应ID
	QuestionID      string        `json:"question_id"`      // 问题ID
	Content         string        `json:"content"`          // 回答内容
	Summary         string        `json:"summary"`          // 回答摘要
	Sources         []KnowledgeItem `json:"sources"`        // 知识来源
	Confidence      float64       `json:"confidence"`       // 置信度
	ProcessingTime  string        `json:"processing_time"`  // 处理时间
	FusionScore     float64       `json:"fusion_score"`    // 融合质量分数
	Recommendations []string      `json:"recommendations"`  // 建议列表
}

// AgentConfig Agent配置结构体
type AgentConfig struct {
	Name        string            `json:"name"`         // Agent名称
	Version     string            `json:"version"`      // 版本号
	Port        int               `json:"port"`         // 服务端口
	Debug       bool              `json:"debug"`        // 调试模式
	OpenAI      OpenAIConfig      `json:"openai"`      // OpenAI配置
	DeepWiki    DeepWikiConfig    `json:"deepwiki"`    // DeepWiki配置
	Higress     HigressConfig     `json:"higress"`     // Higress配置
	Knowledge   KnowledgeConfig   `json:"knowledge"`   // 知识库配置
	Fusion      FusionConfig      `json:"fusion"`      // 融合配置
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

// DeepWikiConfig DeepWiki配置
type DeepWikiConfig struct {
	Enabled     bool   `json:"enabled"`
	Endpoint    string `json:"endpoint"`
	APIKey      string `json:"api_key"`
	Timeout     string `json:"timeout"`
	MaxRetries  int    `json:"max_retries"`
}

// HigressConfig Higress配置
type HigressConfig struct {
	DocsURL     string `json:"docs_url"`
	CacheDuration string `json:"cache_duration"`
	MaxConcurrentRequests int `json:"max_concurrent_requests"`
}

// KnowledgeConfig 知识库配置
type KnowledgeConfig struct {
	Enabled     bool   `json:"enabled"`
	StoragePath string `json:"storage_path"`
	MaxSize     string `json:"max_size"`
	UpdateInterval string `json:"update_interval"`
}

// FusionConfig 融合配置
type FusionConfig struct {
	Enabled     bool    `json:"enabled"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	MaxSources  int     `json:"max_sources"`
	ResponseFormat string `json:"response_format"`
} 