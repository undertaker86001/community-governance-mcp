package model

import (
	"time"
)

// QuestionType 问题类型枚举
type QuestionType string

const (
	QuestionTypeIssue   QuestionType = "issue"   // GitHub Issue
	QuestionTypePR      QuestionType = "pr"      // Pull Request
	QuestionTypeText    QuestionType = "text"    // 图文问题
	QuestionTypeUnknown QuestionType = "unknown" // 未知类型
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
	KnowledgeSourceLocal    KnowledgeSource = "local"    // 本地知识库
	KnowledgeSourceHigress  KnowledgeSource = "higress"  // Higress文档
	KnowledgeSourceDeepWiki KnowledgeSource = "deepwiki" // DeepWiki
	KnowledgeSourceGitHub   KnowledgeSource = "github"   // GitHub Issues/PRs
)

// Question 问题结构体
type Question struct {
	ID        string                 `json:"id"`         // 问题唯一标识
	Type      QuestionType           `json:"type"`       // 问题类型
	Title     string                 `json:"title"`      // 问题标题
	Content   string                 `json:"content"`    // 问题内容
	Author    string                 `json:"author"`     // 提问者
	Priority  Priority               `json:"priority"`   // 优先级
	Tags      []string               `json:"tags"`       // 标签
	CreatedAt time.Time              `json:"created_at"` // 创建时间
	UpdatedAt time.Time              `json:"updated_at"` // 更新时间
	Metadata  map[string]interface{} `json:"metadata"`   // 元数据
}

// KnowledgeItem 知识项结构体
type KnowledgeItem struct {
	ID        string                 `json:"id"`         // 知识项ID
	Source    KnowledgeSource        `json:"source"`     // 知识源
	Title     string                 `json:"title"`      // 标题
	Content   string                 `json:"content"`    // 内容
	URL       string                 `json:"url"`        // 来源URL
	Relevance float64                `json:"relevance"`  // 相关性分数
	Tags      []string               `json:"tags"`       // 标签
	CreatedAt time.Time              `json:"created_at"` // 创建时间
	Metadata  map[string]interface{} `json:"metadata"`   // 元数据
}

// Answer 智能回答结构体
type Answer struct {
	Content     string          `json:"content"`      // 回答内容
	Summary     string          `json:"summary"`      // 回答摘要
	Sources     []KnowledgeItem `json:"sources"`      // 参考知识源
	Confidence  float64         `json:"confidence"`   // 置信度
	FusionScore float64         `json:"fusion_score"` // 融合得分
}

// FusionResult 知识融合结果
// 用于多源知识融合后的中间结果
type FusionResult struct {
	Sources     []KnowledgeItem `json:"sources"`      // 融合后的知识源
	FusionScore float64         `json:"fusion_score"` // 融合得分
	Context     string          `json:"context"`      // 相关上下文
}

// ProcessRequest 处理请求结构体
type ProcessRequest struct {
	Type     QuestionType           `json:"type"`     // 问题类型
	Title    string                 `json:"title"`    // 问题标题
	Content  string                 `json:"content"`  // 问题内容
	Author   string                 `json:"author"`   // 提问者
	Priority Priority               `json:"priority"` // 优先级
	Tags     []string               `json:"tags"`     // 标签
	Metadata map[string]interface{} `json:"metadata"` // 元数据
}

// ProcessResponse 处理响应结构体
type ProcessResponse struct {
	ID              string          `json:"id"`              // 响应ID
	QuestionID      string          `json:"question_id"`     // 问题ID
	Content         string          `json:"content"`         // 回答内容
	Summary         string          `json:"summary"`         // 回答摘要
	Sources         []KnowledgeItem `json:"sources"`         // 知识来源
	Confidence      float64         `json:"confidence"`      // 置信度
	ProcessingTime  string          `json:"processing_time"` // 处理时间
	FusionScore     float64         `json:"fusion_score"`    // 融合质量分数
	Recommendations []string        `json:"recommendations"` // 建议列表
}

// BugAnalysisResult Bug分析结果
type BugAnalysisResult struct {
	ErrorType     string               `json:"error_type"`               // 错误类型
	Severity      string               `json:"severity"`                 // 严重程度
	RootCause     string               `json:"root_cause"`               // 根本原因
	Solutions     []string             `json:"solutions"`                // 解决方案
	Prevention    []string             `json:"prevention"`               // 预防措施
	ImageAnalysis *ImageAnalysisResult `json:"image_analysis,omitempty"` // 图片分析结果
}

// ImageAnalysisResult 图片分析结果
type ImageAnalysisResult struct {
	Description string   `json:"description"` // 图片描述
	Issues      []string `json:"issues"`      // 发现的问题
	Suggestions []string `json:"suggestions"` // 改进建议
	Confidence  float64  `json:"confidence"`  // 分析置信度
}

// CommunityStats 社区统计结果
type CommunityStats struct {
	Period          string                 `json:"period"`           // 统计周期
	TotalIssues     int                    `json:"total_issues"`     // 总Issue数
	OpenIssues      int                    `json:"open_issues"`      // 开放Issue数
	ClosedIssues    int                    `json:"closed_issues"`    // 关闭Issue数
	TotalPRs        int                    `json:"total_prs"`        // 总PR数
	OpenPRs         int                    `json:"open_prs"`         // 开放PR数
	MergedPRs       int                    `json:"merged_prs"`       // 合并PR数
	Contributors    int                    `json:"contributors"`     // 贡献者数
	TopContributors []Contributor          `json:"top_contributors"` // 顶级贡献者
	ActivityTrend   []ActivityData         `json:"activity_trend"`   // 活跃度趋势
	HealthScore     float64                `json:"health_score"`     // 社区健康度
	Metadata        map[string]interface{} `json:"metadata"`         // 元数据
}

// Contributor 贡献者信息
type Contributor struct {
	Username      string `json:"username"`      // 用户名
	AvatarURL     string `json:"avatar_url"`    // 头像URL
	Contributions int    `json:"contributions"` // 贡献数
	LastActive    string `json:"last_active"`   // 最后活跃时间
}

// ActivityData 活跃度数据
type ActivityData struct {
	Date     string `json:"date"`     // 日期
	Issues   int    `json:"issues"`   // Issue数
	PRs      int    `json:"prs"`      // PR数
	Comments int    `json:"comments"` // 评论数
}

// AnalyzeRequest 问题分析请求
// 用于Bug分析、图片分析、Issue分类等
type AnalyzeRequest struct {
	IssueType  string   `json:"issue_type"`  // 问题类型（bug、image、issue等）
	Content    string   `json:"content"`     // 问题内容
	StackTrace string   `json:"stack_trace"` // 错误堆栈
	ImageURL   string   `json:"image_url"`   // 图片URL
	Title      string   `json:"title"`       // 标题
	Assignees  []string `json:"assignees"`   // 预分配人
	Tags       []string `json:"tags"`        // 标签
	Priority   string   `json:"priority"`    // 优先级
}

// AnalyzeResponse 问题分析响应
type AnalyzeResponse struct {
	ID             string   `json:"id"`                 // 分析ID
	ProblemType    string   `json:"problem_type"`       // 问题类型
	Severity       string   `json:"severity,omitempty"` // 严重程度
	Diagnosis      string   `json:"diagnosis"`          // 诊断结果
	Solutions      []string `json:"solutions"`          // 解决方案
	Confidence     float64  `json:"confidence"`         // 置信度
	ProcessingTime string   `json:"processing_time"`    // 分析耗时
}

// BugAnalysis Bug分析结果
type BugAnalysis struct {
	Severity   string   `json:"severity"`   // 严重程度
	RootCause  string   `json:"root_cause"` // 根因
	Solutions  []string `json:"solutions"`  // 解决方案
	Prevention []string `json:"prevention"` // 预防建议
	Confidence float64  `json:"confidence"` // 置信度
}

// ImageAnalysis 图片分析结果
type ImageAnalysis struct {
	ErrorMessages []string `json:"error_messages"` // 错误信息
	Suggestions   []string `json:"suggestions"`    // 建议
	Confidence    float64  `json:"confidence"`     // 置信度
}

// IssueClassification Issue分类结果
type IssueClassification struct {
	Category   string   `json:"category"`   // 分类（bug/feature/文档等）
	Priority   string   `json:"priority"`   // 优先级
	Severity   string   `json:"severity"`   // 严重程度
	Type       string   `json:"type"`       // 类型
	Labels     []string `json:"labels"`     // 建议标签
	Assignees  []string `json:"assignees"`  // 推荐分配人
	Confidence float64  `json:"confidence"` // 置信度
	Reasoning  string   `json:"reasoning"`  // 分类理由
}

// AgentConfig Agent配置结构体
type AgentConfig struct {
	Agent     AgentInfo        `json:"agent"`     // Agent基础信息
	OpenAI    OpenAIConfig     `json:"openai"`    // OpenAI配置
	DeepWiki  DeepWikiConfig   `json:"deepwiki"`  // DeepWiki配置
	Higress   HigressConfig    `json:"higress"`   // Higress配置
	GitHub    GitHubConfig     `json:"github"`    // GitHub配置
	Knowledge KnowledgeConfig  `json:"knowledge"` // 知识库配置
	Fusion    FusionConfig     `json:"fusion"`    // 融合配置
	Logging   LoggingConfig    `json:"logging"`   // 日志配置
	Memory    MemoryConfig     `json:"memory"`    // 记忆组件配置
	Network   NetworkConfig    `json:"network"`   // 网络配置
	MCP       MCPConfig        `json:"mcp"`       // MCP集成配置
}

// MCPConfig MCP集成配置
type MCPConfig struct {
	Enabled string                 `json:"enabled"` // 是否启用MCP
	Timeout string                 `json:"timeout"` // 超时时间
	Servers map[string]MCPServer  `json:"servers"` // MCP服务器配置
}

// MCPServer MCP服务器配置
type MCPServer struct {
	Enabled         bool              `json:"enabled"`         // 是否启用
	ServerURL       string            `json:"server_url"`      // 服务器URL
	ServerLabel     string            `json:"server_label"`    // 服务器标签
	RequireApproval string            `json:"require_approval"` // 审批要求
	AllowedTools    []string          `json:"allowed_tools"`   // 允许的工具
	Headers         map[string]string `json:"headers"`         // 请求头
}

// AgentInfo Agent基础信息
type AgentInfo struct {
	Name    string `json:"name"`    // Agent名称
	Version string `json:"version"`  // 版本号
	Port    int    `json:"port"`    // 服务端口
	Debug   bool   `json:"debug"`   // 调试模式
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
	Enabled    bool   `json:"enabled"`
	Endpoint   string `json:"endpoint"`
	APIKey     string `json:"api_key"`
	Timeout    string `json:"timeout"`
	MaxRetries int    `json:"max_retries"`
}

// HigressConfig Higress配置
type HigressConfig struct {
	DocsURL               string `json:"docs_url"`
	RepoOwner             string `json:"repo_owner"`
	RepoName              string `json:"repo_name"`
	CacheDuration         string `json:"cache_duration"`
	MaxConcurrentRequests int    `json:"max_concurrent_requests"`
}

// GitHubConfig GitHub配置
type GitHubConfig struct {
	Token      string `json:"token"`
	APIURL     string `json:"api_url"`
	Timeout    string `json:"timeout"`
	MaxRetries int    `json:"max_retries"`
}

// KnowledgeConfig 知识库配置
type KnowledgeConfig struct {
	Enabled        bool   `json:"enabled"`
	StoragePath    string `json:"storage_path"`
	MaxSize        string `json:"max_size"`
	UpdateInterval string `json:"update_interval"`
}

// MemoryConfig 记忆组件配置
type MemoryConfig struct {
	WorkingMemoryMaxItems int           `json:"working_memory_max_items"` // 工作记忆最大项数
	WorkingMemoryTTL      time.Duration `json:"working_memory_ttl"`       // 工作记忆生存时间
	ShortTermMemorySlots  int           `json:"short_term_memory_slots"`  // 短期记忆槽位数
	ShortTermMemoryTTL    time.Duration `json:"short_term_memory_ttl"`    // 短期记忆生存时间
	CleanupInterval       time.Duration `json:"cleanup_interval"`         // 清理间隔
	ImportanceThreshold   float64       `json:"importance_threshold"`     // 重要性阈值
}

// FusionConfig 融合配置
type FusionConfig struct {
	Enabled             bool    `json:"enabled"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	MaxSources          int     `json:"max_sources"`
	ResponseFormat      string  `json:"response_format"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `json:"level"`
	Format   string `json:"format"`
	Output   string `json:"output"`
	FilePath string `json:"file_path"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ProxyEnabled bool   `json:"proxy_enabled"` // 是否启用代理
	ProxyURL     string `json:"proxy_url"`     // 代理URL
	ProxyType    string `json:"proxy_type"`    // 代理类型 (http, https, socks5)
}

// Document 文档结构体
type Document struct {
	ID        string                 `json:"id"`         // 文档ID
	Title     string                 `json:"title"`      // 文档标题
	Content   string                 `json:"content"`    // 文档内容
	URL       string                 `json:"url"`        // 文档URL
	Source    string                 `json:"source"`     // 文档来源
	Tags      []string               `json:"tags"`       // 文档标签
	CreatedAt time.Time              `json:"created_at"` // 创建时间
	UpdatedAt time.Time              `json:"updated_at"` // 更新时间
	Metadata  map[string]interface{} `json:"metadata"`   // 元数据
}

// GitHubIssue GitHub Issue结构体
type GitHubIssue struct {
	ID         int           `json:"id"`         // Issue ID
	Number     int           `json:"number"`     // Issue编号
	Title      string        `json:"title"`      // 标题
	Body       string        `json:"body"`       // 内容
	State      string        `json:"state"`      // 状态
	CreatedAt  string        `json:"created_at"` // 创建时间
	UpdatedAt  string        `json:"updated_at"` // 更新时间
	ClosedAt   string        `json:"closed_at"`  // 关闭时间
	User       *GitHubUser   `json:"user"`       // 创建者
	Labels     []string      `json:"labels"`     // 标签
	Assignees  []*GitHubUser `json:"assignees"`  // 分配者
	Comments   int           `json:"comments"`   // 评论数
	HTMLURL    string        `json:"html_url"`   // HTML URL
	Repository string        `json:"repository"` // 仓库名
}

// GitHubComment GitHub评论结构体
type GitHubComment struct {
	ID        int         `json:"id"`         // 评论ID
	Body      string      `json:"body"`       // 评论内容
	User      *GitHubUser `json:"user"`       // 评论者
	CreatedAt string      `json:"created_at"` // 创建时间
	UpdatedAt string      `json:"updated_at"` // 更新时间
	HTMLURL   string      `json:"html_url"`   // HTML URL
}

// GitHubUser GitHub用户结构体
type GitHubUser struct {
	ID        int    `json:"id"`         // 用户ID
	Login     string `json:"login"`      // 用户名
	AvatarURL string `json:"avatar_url"` // 头像URL
	HTMLURL   string `json:"html_url"`   // HTML URL
	Type      string `json:"type"`       // 用户类型
}

// Repository GitHub仓库结构体
type Repository struct {
	ID          int    `json:"id"`          // 仓库ID
	Name        string `json:"name"`        // 仓库名
	FullName    string `json:"full_name"`   // 完整名称
	Description string `json:"description"` // 描述
	Private     bool   `json:"private"`     // 是否私有
	Fork        bool   `json:"fork"`        // 是否Fork
	Stars       int    `json:"stars"`       // 星标数
	Forks       int    `json:"forks"`       // Fork数
	Watchers    int    `json:"watchers"`    // 关注者数
	OpenIssues  int    `json:"open_issues"` // 开放Issue数
	Language    string `json:"language"`    // 主要语言
	CreatedAt   string `json:"created_at"`  // 创建时间
	UpdatedAt   string `json:"updated_at"`  // 更新时间
	HTMLURL     string `json:"html_url"`    // HTML URL
}

// RepositoryStats 仓库统计结构体
type RepositoryStats struct {
	Repository   *Repository `json:"repository"`    // 仓库信息
	OpenIssues   int         `json:"open_issues"`   // 开放Issue数
	ClosedIssues int         `json:"closed_issues"` // 关闭Issue数
	TotalIssues  int         `json:"total_issues"`  // 总Issue数
	LastUpdated  string      `json:"last_updated"`  // 最后更新时间
}

// SearchResult 搜索结果结构体
type SearchResult struct {
	DocumentID     string  `json:"document_id"`     // 文档ID
	Title          string  `json:"title"`           // 标题
	Content        string  `json:"content"`         // 内容
	RelevanceScore float64 `json:"relevance_score"` // 相关性分数
	Snippet        string  `json:"snippet"`         // 片段
}

// KnowledgeSearchResult 知识库搜索结果结构体
type KnowledgeSearchResult struct {
	Query     string         `json:"query"`      // 查询内容
	Results   []SearchResult `json:"results"`    // 搜索结果
	TotalHits int            `json:"total_hits"` // 总命中数
}

// ClassificationStats 分类统计结构体
type ClassificationStats struct {
	CategoryCounts    map[string]int `json:"category_counts"`    // 分类统计
	PriorityCounts    map[string]int `json:"priority_counts"`    // 优先级统计
	SeverityCounts    map[string]int `json:"severity_counts"`    // 严重程度统计
	TypeCounts        map[string]int `json:"type_counts"`        // 类型统计
	TotalIssues       int            `json:"total_issues"`       // 总Issue数
	AverageConfidence float64        `json:"average_confidence"` // 平均置信度
}

// IssueInfo Issue信息结构体
type IssueInfo struct {
	Title  string   `json:"title"`  // 标题
	Body   string   `json:"body"`   // 内容
	Labels []string `json:"labels"` // 标签
}

// Config 配置结构体
type Config struct {
	OpenAIKey   string `json:"openai_key"`   // OpenAI API密钥
	GitHubToken string `json:"github_token"` // GitHub Token
}
