package agent

import (
	"time"
)

// AgentConfig 代理配置
type AgentConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Port    int    `mapstructure:"port"`
	Debug   bool   `mapstructure:"debug"`

	OpenAI   OpenAIConfig   `mapstructure:"openai"`
	DeepWiki DeepWikiConfig `mapstructure:"deepwiki"`
	GitHub   GitHubConfig   `mapstructure:"github"`
	Knowledge KnowledgeConfig `mapstructure:"knowledge"`
	Fusion   FusionConfig   `mapstructure:"fusion"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

// OpenAIConfig OpenAI配置
type OpenAIConfig struct {
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// DeepWikiConfig DeepWiki配置
type DeepWikiConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Endpoint   string `mapstructure:"endpoint"`
	Timeout    string `mapstructure:"timeout"`
	MaxRetries int    `mapstructure:"max_retries"`
}

// GitHubConfig GitHub配置
type GitHubConfig struct {
	Token      string `mapstructure:"token"`
	Owner      string `mapstructure:"owner"`
	Repo       string `mapstructure:"repo"`
	WebhookURL string `mapstructure:"webhook_url"`
}

// KnowledgeConfig 知识库配置
type KnowledgeConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	StoragePath    string `mapstructure:"storage_path"`
	MaxSize        string `mapstructure:"max_size"`
	UpdateInterval string `mapstructure:"update_interval"`
}

// FusionConfig 知识融合配置
type FusionConfig struct {
	Enabled             bool    `mapstructure:"enabled"`
	SimilarityThreshold float64 `mapstructure:"similarity_threshold"`
	MaxSources          int     `mapstructure:"max_sources"`
	ResponseFormat      string  `mapstructure:"response_format"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

// 问题类型枚举
type QuestionType string

const (
	QuestionTypeText   QuestionType = "text"
	QuestionTypeIssue  QuestionType = "issue"
	QuestionTypePR     QuestionType = "pr"
	QuestionTypeBug    QuestionType = "bug"
	QuestionTypeConfig QuestionType = "config"
)

// 优先级枚举
type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityNormal   Priority = "normal"
	PriorityHigh     Priority = "high"
	PriorityUrgent   Priority = "urgent"
	PriorityCritical Priority = "critical"
)

// ProcessRequest 智能问答请求
type ProcessRequest struct {
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Author   string                 `json:"author"`
	Type     string                 `json:"type"`
	Priority string                 `json:"priority"`
	Tags     []string               `json:"tags"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ProcessResponse 智能问答响应
type ProcessResponse struct {
	ID              string                    `json:"id"`
	QuestionID      string                    `json:"question_id"`
	Content         string                    `json:"content"`
	Summary         string                    `json:"summary"`
	Sources         []KnowledgeItem           `json:"sources"`
	Confidence      float64                   `json:"confidence"`
	ProcessingTime  string                    `json:"processing_time"`
	FusionScore     float64                   `json:"fusion_score"`
	Recommendations []string                  `json:"recommendations"`
}

// AnalyzeRequest 问题分析请求
type AnalyzeRequest struct {
	StackTrace string                 `json:"stack_trace"`
	Environment string                `json:"environment"`
	Version    string                 `json:"version"`
	ImageURL   string                 `json:"image_url"`
	IssueType  string                 `json:"issue_type"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// AnalyzeResponse 问题分析响应
type AnalyzeResponse struct {
	ID              string   `json:"id"`
	ProblemType     string   `json:"problem_type"`
	Severity        string   `json:"severity"`
	Diagnosis       string   `json:"diagnosis"`
	Solutions       []string `json:"solutions"`
	Confidence      float64  `json:"confidence"`
	ProcessingTime  string   `json:"processing_time"`
	RelatedIssues   []string `json:"related_issues"`
}

// CommunityStats 社区统计
type CommunityStats struct {
	TotalIssues     int                    `json:"total_issues"`
	OpenIssues      int                    `json:"open_issues"`
	ClosedIssues    int                    `json:"closed_issues"`
	TotalPRs        int                    `json:"total_prs"`
	OpenPRs         int                    `json:"open_prs"`
	MergedPRs       int                    `json:"merged_prs"`
	Contributors    int                    `json:"contributors"`
	ActiveUsers     int                    `json:"active_users"`
	TopContributors []Contributor          `json:"top_contributors"`
	IssueTrends     []IssueTrend           `json:"issue_trends"`
	PRTrends        []PRTrend              `json:"pr_trends"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// Contributor 贡献者信息
type Contributor struct {
	Username    string `json:"username"`
	AvatarURL   string `json:"avatar_url"`
	Contributions int  `json:"contributions"`
	Issues      int   `json:"issues"`
	PRs         int   `json:"prs"`
}

// IssueTrend 问题趋势
type IssueTrend struct {
	Date   string `json:"date"`
	Opened int    `json:"opened"`
	Closed int    `json:"closed"`
}

// PRTrend PR趋势
type PRTrend struct {
	Date   string `json:"date"`
	Opened int    `json:"opened"`
	Merged int    `json:"merged"`
}

// Question 问题对象
type Question struct {
	ID        string                 `json:"id"`
	Type      QuestionType           `json:"type"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Author    string                 `json:"author"`
	Priority  Priority               `json:"priority"`
	Tags      []string               `json:"tags"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// KnowledgeItem 知识项
type KnowledgeItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	URL         string  `json:"url"`
	Source      string  `json:"source"`
	Relevance   float64 `json:"relevance"`
	Confidence  float64 `json:"confidence"`
	LastUpdated time.Time `json:"last_updated"`
}

// FusionResult 知识融合结果
type FusionResult struct {
	Question     *Question       `json:"question"`
	Sources      []KnowledgeItem `json:"sources"`
	FusionScore  float64         `json:"fusion_score"`
	Answer       *Answer         `json:"answer"`
}

// Answer 回答对象
type Answer struct {
	Content    string          `json:"content"`
	Summary    string          `json:"summary"`
	Sources    []KnowledgeItem `json:"sources"`
	Confidence float64         `json:"confidence"`
}

// BugAnalysis Bug分析结果
type BugAnalysis struct {
	ErrorType    string   `json:"error_type"`
	Language     string   `json:"language"`
	Severity     string   `json:"severity"`
	RootCause    string   `json:"root_cause"`
	Solutions    []string `json:"solutions"`
	Prevention   []string `json:"prevention"`
	Confidence   float64  `json:"confidence"`
}

// ImageAnalysis 图片分析结果
type ImageAnalysis struct {
	DetectedElements []string `json:"detected_elements"`
	ErrorMessages    []string `json:"error_messages"`
	UIElements       []string `json:"ui_elements"`
	Suggestions      []string `json:"suggestions"`
	Confidence       float64  `json:"confidence"`
}

// IssueClassification Issue分类结果
type IssueClassification struct {
	Category    string   `json:"category"`
	Priority    string   `json:"priority"`
	Labels      []string `json:"labels"`
	Assignees   []string `json:"assignees"`
	Milestone   string   `json:"milestone"`
	Confidence  float64  `json:"confidence"`
}

// GitHubIssue GitHub Issue信息
type GitHubIssue struct {
	ID          int       `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	Author      string    `json:"author"`
	Labels      []string  `json:"labels"`
	Assignees   []string  `json:"assignees"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
}

// GitHubPR GitHub PR信息
type GitHubPR struct {
	ID          int       `json:"id"`
	Number      int       `json:"number"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	State       string    `json:"state"`
	Author      string    `json:"author"`
	Labels      []string  `json:"labels"`
	Reviewers   []string  `json:"reviewers"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	MergedAt    *time.Time `json:"merged_at"`
	BaseBranch  string    `json:"base_branch"`
	HeadBranch  string    `json:"head_branch"`
} 