package google

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// GoogleManager Google API管理器
type GoogleManager struct {
	gmailClient  *GmailClient
	groupsClient *GroupsClient
	config       *GoogleConfig

	// 内存存储
	issueTracking map[string]*IssueTracking
	emailThreads  map[string]*EmailThread
	mappings      map[string]*IssueEmailMapping

	// 统计信息
	stats *GoogleStats

	// 互斥锁
	mu sync.RWMutex
}

// NewGoogleManager 创建Google API管理器
func NewGoogleManager(config *GoogleConfig) (*GoogleManager, error) {
	// 创建Gmail客户端
	gmailClient, err := NewGmailClient(&config.Gmail)
	if err != nil {
		return nil, fmt.Errorf("创建Gmail客户端失败: %v", err)
	}

	// 创建Groups客户端
	groupsClient, err := NewGroupsClient(&config.Groups)
	if err != nil {
		return nil, fmt.Errorf("创建Groups客户端失败: %v", err)
	}

	return &GoogleManager{
		gmailClient:   gmailClient,
		groupsClient:  groupsClient,
		config:        config,
		issueTracking: make(map[string]*IssueTracking),
		emailThreads:  make(map[string]*EmailThread),
		mappings:      make(map[string]*IssueEmailMapping),
		stats: &GoogleStats{
			LastSync: time.Now(),
		},
	}, nil
}

// ProcessGitHubIssue 处理GitHub Issue
func (m *GoogleManager) ProcessGitHubIssue(issueID, issueURL, issueTitle, issueContent string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 创建Issue跟踪记录
	tracking := &IssueTracking{
		IssueID:      issueID,
		IssueURL:     issueURL,
		IssueTitle:   issueTitle,
		IssueContent: issueContent,
		Status:       IssueStatusNew,
		CreatedAt:    time.Now(),
		LastUpdated:  time.Now(),
	}

	// 分析Issue内容
	analysis, err := m.analyzeIssue(issueContent)
	if err != nil {
		return fmt.Errorf("分析Issue失败: %v", err)
	}

	// 如果无法解决，创建邮件会话
	if !analysis.CanResolve {
		err = m.createEmailThreadForIssue(tracking, analysis)
		if err != nil {
			return fmt.Errorf("创建邮件会话失败: %v", err)
		}
		tracking.Status = IssueStatusWaiting
	} else {
		// 如果可以解决，直接处理
		tracking.Status = IssueStatusResolved
	}

	// 保存跟踪记录
	m.issueTracking[issueID] = tracking
	m.updateStats()

	return nil
}

// HandleEmailReply 处理邮件回复
func (m *GoogleManager) HandleEmailReply(threadID string, reply *EmailReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 查找关联的Issue
	issueID := m.findIssueByThreadID(threadID)
	if issueID == "" {
		return fmt.Errorf("未找到关联的Issue: %s", threadID)
	}

	tracking := m.issueTracking[issueID]
	if tracking == nil {
		return fmt.Errorf("未找到Issue跟踪记录: %s", issueID)
	}

	// 添加回复记录
	tracking.MaintainerReplies = append(tracking.MaintainerReplies, *reply)
	tracking.LastUpdated = time.Now()

	// 分析回复内容
	analysis, err := m.analyzeMaintainerReply(reply.Content)
	if err != nil {
		return fmt.Errorf("分析维护者回复失败: %v", err)
	}

	// 生成Issue回复
	issueReply, err := m.generateIssueReply(tracking, reply, analysis)
	if err != nil {
		return fmt.Errorf("生成Issue回复失败: %v", err)
	}

	// 更新状态
	if analysis.IsResolved {
		tracking.Status = IssueStatusResolved
	} else {
		tracking.Status = IssueStatusReplied
	}

	reply.IssueReply = issueReply
	reply.IsProcessed = true

	m.updateStats()

	return nil
}

// SendEmailToGroup 向邮件组发送邮件
func (m *GoogleManager) SendEmailToGroup(subject, content string, threadID string) (*GmailResponse, error) {
	req := &GmailRequest{
		To:       []string{m.config.Gmail.GroupEmail},
		Subject:  subject,
		Content:  content,
		ThreadID: threadID,
	}

	return m.gmailClient.SendEmail(req)
}

// GetPendingIssues 获取待处理的Issue列表
func (m *GoogleManager) GetPendingIssues() []*IssueTracking {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []*IssueTracking
	for _, tracking := range m.issueTracking {
		if tracking.Status == IssueStatusWaiting || tracking.Status == IssueStatusReplied {
			pending = append(pending, tracking)
		}
	}

	return pending
}

// GetEmailThreads 获取邮件会话列表
func (m *GoogleManager) GetEmailThreads() []*EmailThread {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var threads []*EmailThread
	for _, thread := range m.emailThreads {
		threads = append(threads, thread)
	}

	return threads
}

// GetStats 获取统计信息
func (m *GoogleManager) GetStats() *GoogleStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.stats
}

// SyncEmails 同步邮件
func (m *GoogleManager) SyncEmails() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取新邮件
	emails, err := m.gmailClient.GetEmails("is:unread", 50)
	if err != nil {
		return fmt.Errorf("获取邮件失败: %v", err)
	}

	// 处理新邮件
	for _, email := range emails {
		err = m.processNewEmail(email)
		if err != nil {
			log.Printf("处理新邮件失败 %s: %v", email.ID, err)
		}
	}

	m.stats.LastSync = time.Now()
	m.updateStats()

	return nil
}

// WatchForChanges 监听变化
func (m *GoogleManager) WatchForChanges(topicName string) error {
	return m.gmailClient.WatchInbox(topicName)
}

// StopWatching 停止监听
func (m *GoogleManager) StopWatching() error {
	return m.gmailClient.StopWatching()
}

// AnalyzeIssue 分析Issue
func (m *GoogleManager) AnalyzeIssue(ctx context.Context, issueID string, content string) (*IssueAnalysis, error) {
	return m.analyzeIssue(content)
}

// GenerateIssueReply 生成Issue回复
func (m *GoogleManager) GenerateIssueReply(ctx context.Context, issueID string, analysis *IssueAnalysis) (string, error) {
	// 创建模拟的tracking和reply对象
	tracking := &IssueTracking{
		IssueID: issueID,
	}
	reply := &EmailReply{
		From:    "maintainer@example.com",
		Content: "模拟维护者回复",
	}
	replyAnalysis := &ReplyAnalysis{
		IsResolved: false,
		Action:     "reply",
		Summary:    "维护者已回复",
	}

	return m.generateIssueReply(tracking, reply, replyAnalysis)
}

// analyzeIssue 分析Issue内容
func (m *GoogleManager) analyzeIssue(content string) (*IssueAnalysis, error) {
	// 这里可以集成AI分析功能
	// 暂时使用简单的关键词匹配
	analysis := &IssueAnalysis{
		CanResolve: false,
		Priority:   "medium",
		Tags:       []string{},
		Summary:    "需要维护者协助",
	}

	// 简单的关键词分析
	if containsKeywords(content, []string{"bug", "error", "crash", "fail"}) {
		analysis.Priority = "high"
		analysis.Tags = append(analysis.Tags, "bug")
	}

	if containsKeywords(content, []string{"feature", "enhancement", "improvement"}) {
		analysis.Tags = append(analysis.Tags, "feature")
	}

	// 如果包含特定关键词，标记为可解决
	if containsKeywords(content, []string{"documentation", "typo", "format"}) {
		analysis.CanResolve = true
		analysis.Summary = "可以自动处理"
	}

	return analysis, nil
}

// analyzeMaintainerReply 分析维护者回复
func (m *GoogleManager) analyzeMaintainerReply(content string) (*ReplyAnalysis, error) {
	analysis := &ReplyAnalysis{
		IsResolved: false,
		Action:     "reply",
		Summary:    "维护者已回复",
	}

	// 简单的关键词分析
	if containsKeywords(content, []string{"fixed", "resolved", "done", "complete"}) {
		analysis.IsResolved = true
		analysis.Action = "close"
		analysis.Summary = "问题已解决"
	}

	return analysis, nil
}

// createEmailThreadForIssue 为Issue创建邮件会话
func (m *GoogleManager) createEmailThreadForIssue(tracking *IssueTracking, analysis *IssueAnalysis) error {
	// 生成邮件主题
	subject := fmt.Sprintf("[Issue #%s] %s", tracking.IssueID, tracking.IssueTitle)

	// 生成邮件内容
	content := fmt.Sprintf(`Issue详情:
- URL: %s
- 标题: %s
- 内容: %s
- 优先级: %s
- 标签: %v

分析结果: %s

请协助处理此Issue。`,
		tracking.IssueURL,
		tracking.IssueTitle,
		tracking.IssueContent,
		analysis.Priority,
		analysis.Tags,
		analysis.Summary,
	)

	// 发送邮件
	response, err := m.SendEmailToGroup(subject, content, "")
	if err != nil {
		return err
	}

	// 创建邮件会话记录
	thread := &EmailThread{
		ID:        response.ThreadID,
		Subject:   subject,
		IssueID:   tracking.IssueID,
		Status:    ThreadStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存映射关系
	mapping := &IssueEmailMapping{
		IssueID:   tracking.IssueID,
		ThreadID:  response.ThreadID,
		Subject:   subject,
		CreatedAt: time.Now(),
	}

	m.emailThreads[response.ThreadID] = thread
	m.mappings[tracking.IssueID] = mapping
	tracking.EmailThreadID = response.ThreadID

	return nil
}

// generateIssueReply 生成Issue回复
func (m *GoogleManager) generateIssueReply(tracking *IssueTracking, reply *EmailReply, analysis *ReplyAnalysis) (string, error) {
	// 这里可以集成AI生成回复内容
	content := fmt.Sprintf(`感谢维护者的回复:

%s

维护者回复: %s

状态: %s`,
		reply.From,
		reply.Content,
		analysis.Summary,
	)

	return content, nil
}

// processNewEmail 处理新邮件
func (m *GoogleManager) processNewEmail(email *EmailMessage) error {
	// 检查是否是维护者回复
	if m.isMaintainerReply(email) {
		reply := &EmailReply{
			From:        email.From,
			Content:     email.Content,
			Timestamp:   email.Timestamp,
			IsProcessed: false,
		}

		return m.HandleEmailReply(email.ThreadID, reply)
	}

	return nil
}

// isMaintainerReply 检查是否是维护者回复
func (m *GoogleManager) isMaintainerReply(email *EmailMessage) bool {
	// 检查发件人是否是维护者
	// 这里可以根据实际需求配置维护者邮箱列表
	maintainers := []string{"maintainer@example.com"}

	for _, maintainer := range maintainers {
		if email.From == maintainer {
			return true
		}
	}

	return false
}

// findIssueByThreadID 根据会话ID查找Issue
func (m *GoogleManager) findIssueByThreadID(threadID string) string {
	for issueID, mapping := range m.mappings {
		if mapping.ThreadID == threadID {
			return issueID
		}
	}
	return ""
}

// updateStats 更新统计信息
func (m *GoogleManager) updateStats() {
	m.stats.TotalIssues = len(m.issueTracking)
	m.stats.PendingIssues = 0
	m.stats.ActiveThreads = 0

	for _, tracking := range m.issueTracking {
		if tracking.Status == IssueStatusWaiting || tracking.Status == IssueStatusReplied {
			m.stats.PendingIssues++
		}
	}

	for _, thread := range m.emailThreads {
		if thread.Status == ThreadStatusPending || thread.Status == ThreadStatusReplied {
			m.stats.ActiveThreads++
		}
	}

	m.stats.TotalEmails = len(m.emailThreads)
}

// containsKeywords 检查是否包含关键词
func containsKeywords(content string, keywords []string) bool {
	content = strings.ToLower(content)
	for _, keyword := range keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// IssueAnalysis Issue分析结果
type IssueAnalysis struct {
	CanResolve bool     `json:"can_resolve"`
	Priority   string   `json:"priority"`
	Tags       []string `json:"tags"`
	Summary    string   `json:"summary"`
}

// ReplyAnalysis 回复分析结果
type ReplyAnalysis struct {
	IsResolved bool   `json:"is_resolved"`
	Action     string `json:"action"`
	Summary    string `json:"summary"`
}
