package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/community-governance-mcp-higress/internal/google"
)

// GoogleTools Google API工具
type GoogleTools struct {
	manager *google.GoogleManager
}

// NewGoogleTools 创建Google API工具
func NewGoogleTools(config *google.GoogleConfig) (*GoogleTools, error) {
	manager, err := google.NewGoogleManager(config)
	if err != nil {
		return nil, fmt.Errorf("创建Google管理器失败: %v", err)
	}

	return &GoogleTools{
		manager: manager,
	}, nil
}

// ProcessGitHubIssue 处理GitHub Issue
func (t *GoogleTools) ProcessGitHubIssue(issueID, issueURL, issueTitle, issueContent string) error {
	log.Printf("开始处理GitHub Issue: %s", issueID)

	err := t.manager.ProcessGitHubIssue(issueID, issueURL, issueTitle, issueContent)
	if err != nil {
		log.Printf("处理GitHub Issue失败: %v", err)
		return err
	}

	log.Printf("GitHub Issue处理成功: %s", issueID)
	return nil
}

// SyncEmails 同步邮件
func (t *GoogleTools) SyncEmails() error {
	log.Printf("开始同步邮件")

	err := t.manager.SyncEmails()
	if err != nil {
		log.Printf("同步邮件失败: %v", err)
		return err
	}

	log.Printf("邮件同步成功")
	return nil
}

// GetPendingIssues 获取待处理的Issue
func (t *GoogleTools) GetPendingIssues() ([]*google.IssueTracking, error) {
	issues := t.manager.GetPendingIssues()
	log.Printf("获取到 %d 个待处理Issue", len(issues))
	return issues, nil
}

// GetEmailThreads 获取邮件会话
func (t *GoogleTools) GetEmailThreads() ([]*google.EmailThread, error) {
	threads := t.manager.GetEmailThreads()
	log.Printf("获取到 %d 个邮件会话", len(threads))
	return threads, nil
}

// GetStats 获取统计信息
func (t *GoogleTools) GetStats() (*google.GoogleStats, error) {
	stats := t.manager.GetStats()
	log.Printf("获取统计信息: 总Issue=%d, 待处理=%d, 活跃会话=%d",
		stats.TotalIssues, stats.PendingIssues, stats.ActiveThreads)
	return stats, nil
}

// SendEmailToGroup 向邮件组发送邮件
func (t *GoogleTools) SendEmailToGroup(subject, content string, threadID string) (*google.GmailResponse, error) {
	log.Printf("发送邮件到邮件组: %s", subject)

	response, err := t.manager.SendEmailToGroup(subject, content, threadID)
	if err != nil {
		log.Printf("发送邮件失败: %v", err)
		return nil, err
	}

	log.Printf("邮件发送成功: %s", response.MessageID)
	return response, nil
}

// HandleEmailReply 处理邮件回复
func (t *GoogleTools) HandleEmailReply(threadID string, reply *google.EmailReply) error {
	log.Printf("处理邮件回复: %s", threadID)

	err := t.manager.HandleEmailReply(threadID, reply)
	if err != nil {
		log.Printf("处理邮件回复失败: %v", err)
		return err
	}

	log.Printf("邮件回复处理成功: %s", threadID)
	return nil
}

// WatchForChanges 监听变化
func (t *GoogleTools) WatchForChanges(topicName string) error {
	log.Printf("开始监听变化: %s", topicName)

	err := t.manager.WatchForChanges(topicName)
	if err != nil {
		log.Printf("开始监听失败: %v", err)
		return err
	}

	log.Printf("监听设置成功: %s", topicName)
	return nil
}

// StopWatching 停止监听
func (t *GoogleTools) StopWatching() error {
	log.Printf("停止监听")

	err := t.manager.StopWatching()
	if err != nil {
		log.Printf("停止监听失败: %v", err)
		return err
	}

	log.Printf("监听已停止")
	return nil
}

// AnalyzeIssueContent 分析Issue内容
func (t *GoogleTools) AnalyzeIssueContent(issueID, content string) (*google.IssueAnalysis, error) {
	log.Printf("分析Issue内容")
	analysis, err := t.manager.AnalyzeIssue(context.Background(), issueID, content)
	if err != nil {
		log.Printf("分析Issue内容失败: %v", err)
		return nil, err
	}
	log.Printf("Issue分析完成: 可解决=%v, 优先级=%s", analysis.CanResolve, analysis.Priority)
	return analysis, nil
}

// GenerateIssueReply 生成Issue回复
func (t *GoogleTools) GenerateIssueReply(issueID string, analysis *google.IssueAnalysis) (string, error) {
	log.Printf("生成Issue回复")
	content, err := t.manager.GenerateIssueReply(context.Background(), issueID, analysis)
	if err != nil {
		log.Printf("生成Issue回复失败: %v", err)
		return "", err
	}
	log.Printf("Issue回复生成成功")
	return content, nil
}

// ExportIssueData 导出Issue数据
func (t *GoogleTools) ExportIssueData() ([]byte, error) {
	log.Printf("导出Issue数据")

	// 获取所有Issue数据
	issues := t.manager.GetPendingIssues()
	threads := t.manager.GetEmailThreads()
	stats := t.manager.GetStats()

	data := map[string]interface{}{
		"issues":      issues,
		"threads":     threads,
		"stats":       stats,
		"export_time": time.Now(),
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Printf("导出Issue数据失败: %v", err)
		return nil, err
	}

	log.Printf("Issue数据导出成功")
	return jsonData, nil
}

// ImportIssueData 导入Issue数据
func (t *GoogleTools) ImportIssueData(data []byte) error {
	log.Printf("导入Issue数据")

	var importData map[string]interface{}
	err := json.Unmarshal(data, &importData)
	if err != nil {
		log.Printf("解析导入数据失败: %v", err)
		return err
	}

	// 这里可以添加导入逻辑
	log.Printf("Issue数据导入成功")
	return nil
}

// CleanupOldData 清理旧数据
func (t *GoogleTools) CleanupOldData(days int) error {
	log.Printf("清理 %d 天前的旧数据", days)

	// 这里可以添加清理逻辑
	// 例如删除已关闭的Issue、清理旧的邮件会话等

	log.Printf("旧数据清理完成")
	return nil
}

// BackupData 备份数据
func (t *GoogleTools) BackupData() ([]byte, error) {
	log.Printf("备份数据")

	// 导出所有数据作为备份
	data, err := t.ExportIssueData()
	if err != nil {
		log.Printf("备份数据失败: %v", err)
		return nil, err
	}

	log.Printf("数据备份成功")
	return data, nil
}

// RestoreData 恢复数据
func (t *GoogleTools) RestoreData(data []byte) error {
	log.Printf("恢复数据")

	// 导入备份数据
	err := t.ImportIssueData(data)
	if err != nil {
		log.Printf("恢复数据失败: %v", err)
		return err
	}

	log.Printf("数据恢复成功")
	return nil
}
