package agent

import (
	"github.com/community-governance-mcp-higress/internal/model"
)

// 重新导出model包中的类型，保持向后兼容
type QuestionType = model.QuestionType
type Priority = model.Priority
type KnowledgeSource = model.KnowledgeSource
type Question = model.Question
type KnowledgeItem = model.KnowledgeItem
type Answer = model.Answer
type FusionResult = model.FusionResult
type ProcessRequest = model.ProcessRequest
type ProcessResponse = model.ProcessResponse
type BugAnalysisResult = model.BugAnalysisResult
type ImageAnalysisResult = model.ImageAnalysisResult
type CommunityStats = model.CommunityStats
type Contributor = model.Contributor
type ActivityData = model.ActivityData
type AnalyzeRequest = model.AnalyzeRequest
type AnalyzeResponse = model.AnalyzeResponse
type BugAnalysis = model.BugAnalysis
type ImageAnalysis = model.ImageAnalysis
type IssueClassification = model.IssueClassification
type AgentConfig = model.AgentConfig
type OpenAIConfig = model.OpenAIConfig
type DeepWikiConfig = model.DeepWikiConfig
type HigressConfig = model.HigressConfig
type GitHubConfig = model.GitHubConfig
type KnowledgeConfig = model.KnowledgeConfig
type MemoryConfig = model.MemoryConfig
type FusionConfig = model.FusionConfig
type LoggingConfig = model.LoggingConfig

// 重新导出常量
const (
	QuestionTypeIssue       = model.QuestionTypeIssue
	QuestionTypePR          = model.QuestionTypePR
	QuestionTypeText        = model.QuestionTypeText
	QuestionTypeUnknown     = model.QuestionTypeUnknown
	PriorityLow             = model.PriorityLow
	PriorityMedium          = model.PriorityMedium
	PriorityHigh            = model.PriorityHigh
	PriorityUrgent          = model.PriorityUrgent
	KnowledgeSourceLocal    = model.KnowledgeSourceLocal
	KnowledgeSourceHigress  = model.KnowledgeSourceHigress
	KnowledgeSourceDeepWiki = model.KnowledgeSourceDeepWiki
	KnowledgeSourceGitHub   = model.KnowledgeSourceGitHub
)
