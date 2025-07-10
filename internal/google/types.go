package google

import (
	"time"
)

// GmailConfig Gmail配置
type GmailConfig struct {
	CredentialsFile string   `json:"credentials_file"` // 服务账号凭证文件
	TokenFile       string   `json:"token_file"`       // 访问令牌文件
	GroupEmail      string   `json:"group_email"`      // 邮件组地址
	Scopes          []string `json:"scopes"`           // 权限范围
}

// GroupsConfig Google Groups配置
type GroupsConfig struct {
	AdminEmail string `json:"admin_email"` // 管理员邮箱
	GroupKey   string `json:"group_key"`   // 邮件组标识
	Domain     string `json:"domain"`      // 域名
}

// GoogleConfig Google API配置
type GoogleConfig struct {
	Gmail  GmailConfig  `json:"gmail"`  // Gmail配置
	Groups GroupsConfig `json:"groups"` // Groups配置
}

// EmailMessage 邮件消息
type EmailMessage struct {
	ID        string    `json:"id"`         // 邮件ID
	ThreadID  string    `json:"thread_id"`  // 会话ID
	From      string    `json:"from"`       // 发件人
	To        []string  `json:"to"`         // 收件人
	Subject   string    `json:"subject"`    // 主题
	Content   string    `json:"content"`    // 内容
	Timestamp time.Time `json:"timestamp"`  // 时间戳
	Labels    []string  `json:"labels"`     // 标签
	IsRead    bool      `json:"is_read"`    // 是否已读
	IsReplied bool      `json:"is_replied"` // 是否已回复
}

// EmailThread 邮件会话
type EmailThread struct {
	ID        string         `json:"id"`         // 会话ID
	Subject   string         `json:"subject"`    // 主题
	Messages  []EmailMessage `json:"messages"`   // 邮件列表
	IssueID   string         `json:"issue_id"`   // 关联的Issue ID
	Status    ThreadStatus   `json:"status"`     // 会话状态
	CreatedAt time.Time      `json:"created_at"` // 创建时间
	UpdatedAt time.Time      `json:"updated_at"` // 更新时间
}

// ThreadStatus 会话状态
type ThreadStatus string

const (
	ThreadStatusPending  ThreadStatus = "pending"  // 等待回复
	ThreadStatusReplied  ThreadStatus = "replied"  // 已回复
	ThreadStatusResolved ThreadStatus = "resolved" // 已解决
	ThreadStatusClosed   ThreadStatus = "closed"   // 已关闭
)

// IssueTracking Issue跟踪
type IssueTracking struct {
	IssueID           string       `json:"issue_id"`           // Issue ID
	IssueURL          string       `json:"issue_url"`          // Issue URL
	IssueTitle        string       `json:"issue_title"`        // Issue标题
	IssueContent      string       `json:"issue_content"`      // Issue内容
	EmailThreadID     string       `json:"email_thread_id"`    // 邮件会话ID
	Status            IssueStatus  `json:"status"`             // Issue状态
	Priority          string       `json:"priority"`           // 优先级
	Tags              []string     `json:"tags"`               // 标签
	CreatedAt         time.Time    `json:"created_at"`         // 创建时间
	LastUpdated       time.Time    `json:"last_updated"`       // 最后更新时间
	MaintainerReplies []EmailReply `json:"maintainer_replies"` // Maintainer回复
}

// IssueStatus Issue状态
type IssueStatus string

const (
	IssueStatusNew       IssueStatus = "new"       // 新建
	IssueStatusAnalyzing IssueStatus = "analyzing" // 分析中
	IssueStatusWaiting   IssueStatus = "waiting"   // 等待回复
	IssueStatusReplied   IssueStatus = "replied"   // 已回复
	IssueStatusResolved  IssueStatus = "resolved"  // 已解决
	IssueStatusClosed    IssueStatus = "closed"    // 已关闭
)

// EmailReply 邮件回复
type EmailReply struct {
	From        string    `json:"from"`         // 发件人
	Content     string    `json:"content"`      // 内容
	Timestamp   time.Time `json:"timestamp"`    // 时间戳
	IssueReply  string    `json:"issue_reply"`  // Issue回复内容
	IsProcessed bool      `json:"is_processed"` // 是否已处理
}

// GmailRequest Gmail请求
type GmailRequest struct {
	To       []string `json:"to"`                  // 收件人
	Subject  string   `json:"subject"`             // 主题
	Content  string   `json:"content"`             // 内容
	ThreadID string   `json:"thread_id,omitempty"` // 会话ID（回复时）
}

// GmailResponse Gmail响应
type GmailResponse struct {
	MessageID string `json:"message_id"`      // 邮件ID
	ThreadID  string `json:"thread_id"`       // 会话ID
	Success   bool   `json:"success"`         // 是否成功
	Error     string `json:"error,omitempty"` // 错误信息
}

// IssueEmailMapping Issue邮件映射
type IssueEmailMapping struct {
	IssueID   string    `json:"issue_id"`   // Issue ID
	ThreadID  string    `json:"thread_id"`  // 邮件会话ID
	Subject   string    `json:"subject"`    // 邮件主题
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

// GoogleStats Google API统计
type GoogleStats struct {
	TotalIssues   int       `json:"total_issues"`   // 总Issue数
	PendingIssues int       `json:"pending_issues"` // 待处理Issue数
	ActiveThreads int       `json:"active_threads"` // 活跃会话数
	TotalEmails   int       `json:"total_emails"`   // 总邮件数
	LastSync      time.Time `json:"last_sync"`      // 最后同步时间
	SuccessRate   float64   `json:"success_rate"`   // 成功率
}
