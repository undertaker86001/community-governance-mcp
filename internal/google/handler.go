package google

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// GoogleHandler Google API HTTP处理器
type GoogleHandler struct {
	manager *GoogleManager
}

// NewGoogleHandler 创建Google API处理器
func NewGoogleHandler(manager *GoogleManager) *GoogleHandler {
	return &GoogleHandler{
		manager: manager,
	}
}

// RegisterRoutes 注册路由
func (h *GoogleHandler) RegisterRoutes(router *mux.Router) {
	// Issue相关路由
	router.HandleFunc("/api/google/issues", h.ProcessIssue).Methods("POST")
	router.HandleFunc("/api/google/issues", h.GetIssues).Methods("GET")
	router.HandleFunc("/api/google/issues/{id}", h.GetIssue).Methods("GET")
	router.HandleFunc("/api/google/issues/{id}/status", h.UpdateIssueStatus).Methods("PUT")

	// 邮件相关路由
	router.HandleFunc("/api/google/emails", h.GetEmails).Methods("GET")
	router.HandleFunc("/api/google/emails/send", h.SendEmail).Methods("POST")
	router.HandleFunc("/api/google/emails/sync", h.SyncEmails).Methods("POST")
	router.HandleFunc("/api/google/emails/reply", h.HandleEmailReply).Methods("POST")

	// 会话相关路由
	router.HandleFunc("/api/google/threads", h.GetThreads).Methods("GET")
	router.HandleFunc("/api/google/threads/{id}", h.GetThread).Methods("GET")

	// 统计相关路由
	router.HandleFunc("/api/google/stats", h.GetStats).Methods("GET")

	// 监听相关路由
	router.HandleFunc("/api/google/watch", h.StartWatching).Methods("POST")
	router.HandleFunc("/api/google/watch", h.StopWatching).Methods("DELETE")
}

// ProcessIssue 处理Issue请求
func (h *GoogleHandler) ProcessIssue(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IssueID      string `json:"issue_id"`
		IssueURL     string `json:"issue_url"`
		IssueTitle   string `json:"issue_title"`
		IssueContent string `json:"issue_content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if req.IssueID == "" || req.IssueURL == "" || req.IssueTitle == "" {
		http.Error(w, "缺少必要参数", http.StatusBadRequest)
		return
	}

	err := h.manager.ProcessGitHubIssue(req.IssueID, req.IssueURL, req.IssueTitle, req.IssueContent)
	if err != nil {
		http.Error(w, fmt.Sprintf("处理Issue失败: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":  true,
		"message":  "Issue处理成功",
		"issue_id": req.IssueID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIssues 获取Issue列表
func (h *GoogleHandler) GetIssues(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var issues []*IssueTracking
	if status == "pending" {
		issues = h.manager.GetPendingIssues()
	} else {
		// 获取所有Issue
		// 这里需要添加获取所有Issue的方法
		issues = []*IssueTracking{}
	}

	response := map[string]interface{}{
		"success": true,
		"issues":  issues,
		"count":   len(issues),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetIssue 获取单个Issue
func (h *GoogleHandler) GetIssue(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// issueID := vars["id"]

	// 这里需要添加获取单个Issue的方法
	// 暂时返回空数据
	response := map[string]interface{}{
		"success": true,
		"issue":   nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateIssueStatus 更新Issue状态
func (h *GoogleHandler) UpdateIssueStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	issueID := vars["id"]

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	// 这里需要添加更新Issue状态的方法
	response := map[string]interface{}{
		"success":  true,
		"message":  "状态更新成功",
		"issue_id": issueID,
		"status":   req.Status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEmails 获取邮件列表
func (h *GoogleHandler) GetEmails(w http.ResponseWriter, r *http.Request) {
	maxResults, _ := strconv.ParseInt(r.URL.Query().Get("max_results"), 10, 64)
	if maxResults == 0 {
		maxResults = 50
	}

	// 这里需要添加获取邮件列表的方法
	emails := []*EmailMessage{}

	response := map[string]interface{}{
		"success": true,
		"emails":  emails,
		"count":   len(emails),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendEmail 发送邮件
func (h *GoogleHandler) SendEmail(w http.ResponseWriter, r *http.Request) {
	var req GmailRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if len(req.To) == 0 || req.Subject == "" || req.Content == "" {
		http.Error(w, "缺少必要参数", http.StatusBadRequest)
		return
	}

	response, err := h.manager.SendEmailToGroup(req.Subject, req.Content, req.ThreadID)
	if err != nil {
		http.Error(w, fmt.Sprintf("发送邮件失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SyncEmails 同步邮件
func (h *GoogleHandler) SyncEmails(w http.ResponseWriter, r *http.Request) {
	err := h.manager.SyncEmails()
	if err != nil {
		http.Error(w, fmt.Sprintf("同步邮件失败: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   "邮件同步成功",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HandleEmailReply 处理邮件回复
func (h *GoogleHandler) HandleEmailReply(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ThreadID string     `json:"thread_id"`
		Reply    EmailReply `json:"reply"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if req.ThreadID == "" {
		http.Error(w, "缺少会话ID", http.StatusBadRequest)
		return
	}

	err := h.manager.HandleEmailReply(req.ThreadID, &req.Reply)
	if err != nil {
		http.Error(w, fmt.Sprintf("处理邮件回复失败: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   "邮件回复处理成功",
		"thread_id": req.ThreadID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetThreads 获取会话列表
func (h *GoogleHandler) GetThreads(w http.ResponseWriter, r *http.Request) {
	threads := h.manager.GetEmailThreads()

	response := map[string]interface{}{
		"success": true,
		"threads": threads,
		"count":   len(threads),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetThread 获取单个会话
func (h *GoogleHandler) GetThread(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	// threadID := vars["id"]

	// 这里需要添加获取单个会话的方法
	response := map[string]interface{}{
		"success": true,
		"thread":  nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetStats 获取统计信息
func (h *GoogleHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := h.manager.GetStats()

	response := map[string]interface{}{
		"success": true,
		"stats":   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StartWatching 开始监听
func (h *GoogleHandler) StartWatching(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TopicName string `json:"topic_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if req.TopicName == "" {
		http.Error(w, "缺少主题名称", http.StatusBadRequest)
		return
	}

	err := h.manager.WatchForChanges(req.TopicName)
	if err != nil {
		http.Error(w, fmt.Sprintf("开始监听失败: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    "开始监听成功",
		"topic_name": req.TopicName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// StopWatching 停止监听
func (h *GoogleHandler) StopWatching(w http.ResponseWriter, r *http.Request) {
	err := h.manager.StopWatching()
	if err != nil {
		http.Error(w, fmt.Sprintf("停止监听失败: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "停止监听成功",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
