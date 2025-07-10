package google

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GmailClient Gmail客户端
type GmailClient struct {
	service *gmail.Service
	config  *GmailConfig
	userID  string
}

// NewGmailClient 创建Gmail客户端
func NewGmailClient(config *GmailConfig) (*GmailClient, error) {
	// 读取服务账号凭证
	credentials, err := ioutil.ReadFile(config.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("无法读取凭证文件: %v", err)
	}

	// 创建JWT配置
	jwtConfig, err := google.JWTConfigFromJSON(credentials, config.Scopes...)
	if err != nil {
		return nil, fmt.Errorf("无法创建JWT配置: %v", err)
	}

	// 创建HTTP客户端
	client := jwtConfig.Client(context.Background())

	// 创建Gmail服务
	service, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("无法创建Gmail服务: %v", err)
	}

	return &GmailClient{
		service: service,
		config:  config,
		userID:  "me", // 使用"me"表示当前用户
	}, nil
}

// SendEmail 发送邮件
func (c *GmailClient) SendEmail(req *GmailRequest) (*GmailResponse, error) {
	// 构建邮件内容
	message := &gmail.Message{
		Raw: base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf(
			"To: %s\r\n"+
				"Subject: %s\r\n"+
				"Content-Type: text/plain; charset=UTF-8\r\n"+
				"\r\n"+
				"%s",
			strings.Join(req.To, ", "),
			req.Subject,
			req.Content,
		))),
	}

	// 如果是回复邮件，设置In-Reply-To和References头
	if req.ThreadID != "" {
		// 获取原始邮件信息
		thread, err := c.service.Users.Threads.Get(c.userID, req.ThreadID).Do()
		if err != nil {
			return nil, fmt.Errorf("无法获取会话信息: %v", err)
		}

		if len(thread.Messages) > 0 {
			originalMessageID := thread.Messages[0].Id
			message.Raw = base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf(
				"To: %s\r\n"+
					"Subject: %s\r\n"+
					"In-Reply-To: <%s>\r\n"+
					"References: <%s>\r\n"+
					"Content-Type: text/plain; charset=UTF-8\r\n"+
					"\r\n"+
					"%s",
				strings.Join(req.To, ", "),
				req.Subject,
				originalMessageID,
				originalMessageID,
				req.Content,
			)))
		}
	}

	// 发送邮件
	sentMessage, err := c.service.Users.Messages.Send(c.userID, message).Do()
	if err != nil {
		return nil, fmt.Errorf("发送邮件失败: %v", err)
	}

	return &GmailResponse{
		MessageID: sentMessage.Id,
		ThreadID:  sentMessage.ThreadId,
		Success:   true,
	}, nil
}

// GetEmails 获取邮件列表
func (c *GmailClient) GetEmails(query string, maxResults int64) ([]*EmailMessage, error) {
	// 构建查询条件
	if query == "" {
		query = "in:inbox"
	}

	// 获取邮件列表
	response, err := c.service.Users.Messages.List(c.userID).Q(query).MaxResults(maxResults).Do()
	if err != nil {
		return nil, fmt.Errorf("获取邮件列表失败: %v", err)
	}

	var emails []*EmailMessage
	for _, msg := range response.Messages {
		email, err := c.GetEmail(msg.Id)
		if err != nil {
			log.Printf("获取邮件详情失败 %s: %v", msg.Id, err)
			continue
		}
		emails = append(emails, email)
	}

	return emails, nil
}

// GetEmail 获取单个邮件详情
func (c *GmailClient) GetEmail(messageID string) (*EmailMessage, error) {
	message, err := c.service.Users.Messages.Get(c.userID, messageID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("获取邮件详情失败: %v", err)
	}

	// 解析邮件头信息
	var from, subject string
	var to []string
	for _, header := range message.Payload.Headers {
		switch header.Name {
		case "From":
			from = header.Value
		case "To":
			to = strings.Split(header.Value, ",")
		case "Subject":
			subject = header.Value
		}
	}

	// 获取邮件内容
	content := c.extractEmailContent(message.Payload)

	// 解析时间戳
	timestamp := time.Now()
	if message.InternalDate > 0 {
		timestamp = time.Unix(message.InternalDate/1000, 0)
	}

	return &EmailMessage{
		ID:        message.Id,
		ThreadID:  message.ThreadId,
		From:      from,
		To:        to,
		Subject:   subject,
		Content:   content,
		Timestamp: timestamp,
		Labels:    message.LabelIds,
		IsRead:    !contains(message.LabelIds, "UNREAD"),
		IsReplied: contains(message.LabelIds, "SENT"),
	}, nil
}

// GetThread 获取邮件会话
func (c *GmailClient) GetThread(threadID string) (*EmailThread, error) {
	thread, err := c.service.Users.Threads.Get(c.userID, threadID).Do()
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %v", err)
	}

	var messages []EmailMessage
	for _, msg := range thread.Messages {
		email, err := c.GetEmail(msg.Id)
		if err != nil {
			log.Printf("获取会话中的邮件失败 %s: %v", msg.Id, err)
			continue
		}
		messages = append(messages, *email)
	}

	// 确定会话状态
	status := ThreadStatusPending
	if len(messages) > 0 {
		lastMessage := messages[len(messages)-1]
		if lastMessage.IsReplied {
			status = ThreadStatusReplied
		}
	}

	return &EmailThread{
		ID:        thread.Id,
		Subject:   messages[0].Subject,
		Messages:  messages,
		Status:    status,
		CreatedAt: messages[0].Timestamp,
		UpdatedAt: messages[len(messages)-1].Timestamp,
	}, nil
}

// WatchInbox 监听收件箱变化
func (c *GmailClient) WatchInbox(topicName string) error {
	request := &gmail.WatchRequest{
		TopicName: topicName,
		LabelIds:  []string{"INBOX"},
	}

	_, err := c.service.Users.Watch(c.userID, request).Do()
	if err != nil {
		return fmt.Errorf("设置监听失败: %v", err)
	}

	return nil
}

// StopWatching 停止监听
func (c *GmailClient) StopWatching() error {
	err := c.service.Users.Stop(c.userID).Do()
	if err != nil {
		return fmt.Errorf("停止监听失败: %v", err)
	}
	return nil
}

// extractEmailContent 提取邮件内容
func (c *GmailClient) extractEmailContent(payload *gmail.MessagePart) string {
	if payload.Body.Data != "" {
		data, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			return string(data)
		}
	}

	if payload.Parts != nil {
		for _, part := range payload.Parts {
			if part.MimeType == "text/plain" {
				data, err := base64.URLEncoding.DecodeString(part.Body.Data)
				if err == nil {
					return string(data)
				}
			}
		}
	}

	return ""
}

// contains 检查切片是否包含指定元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
