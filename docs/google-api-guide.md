# Google API使用指南

## 概述

本指南详细说明如何使用Google API集成功能，实现Agent作为Google账号加入私有邮件组，分析GitHub Issue，无法解决的问题通过邮件组与维护者交流，维护者回复后Agent理解并回复Issue的完整流程。

## 功能特性

### 核心功能
- **GitHub Issue分析**: 自动分析GitHub Issue内容，判断是否可以自动解决
- **邮件组集成**: Agent作为成员加入私有邮件组，与维护者进行交流
- **多问题并行处理**: 支持同时处理多个Issue，每个Issue创建独立的邮件会话
- **邮件会话管理**: 维护Issue与邮件会话的映射关系，跟踪处理状态
- **自动回复生成**: 基于维护者回复自动生成Issue回复内容
- **LLM/DeepWiki集成**: 结合LLM和DeepWiki进行深度分析和知识检索

### 技术特性
- **Google Gmail API**: 发送和接收邮件，管理邮件会话
- **Google Groups API**: 管理邮件组成员，获取组信息
- **实时监听**: 监听邮件变化，及时处理新邮件
- **状态跟踪**: 完整的Issue处理状态跟踪和统计
- **数据备份**: 支持数据导出、导入和备份功能

## 完整流程示例

### 场景：AI Gateway连接超时问题

#### 1. 用户提出Issue
用户创建了一个关于AI Gateway连接超时的GitHub Issue：

```bash
# 用户创建GitHub Issue
curl -X POST https://api.github.com/repos/alibaba/higress/issues \
  -H "Authorization: token $GITHUB_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bug: Connection timeout in AI Gateway",
    "body": "When using AI Gateway with large models, connections frequently timeout after 30 seconds. This happens especially with GPT-4 models."
  }'
```

#### 2. Agent接收并分析Issue
Agent接收到Issue后，进行智能分析：

```bash
# Agent接收并分析Issue
curl -X POST http://localhost:8080/api/google/issues \
  -H "Content-Type: application/json" \
  -d '{
    "issue_id": "123",
    "issue_url": "https://github.com/alibaba/higress/issues/123",
    "issue_title": "Bug: Connection timeout in AI Gateway",
    "issue_content": "When using AI Gateway with large models, connections frequently timeout after 30 seconds. This happens especially with GPT-4 models."
  }'
```

**Agent分析结果**:
```json
{
  "success": true,
  "analysis": {
    "can_resolve": false,
    "priority": "high",
    "tags": ["bug", "ai-gateway", "timeout"],
    "summary": "需要维护者协助处理AI Gateway连接超时问题",
    "requires_maintainer": true
  },
  "issue_id": "123",
  "status": "waiting"
}
```

#### 3. Agent发送邮件给维护者
由于无法自动解决，Agent自动发送邮件给维护者邮件组：

```bash
# Agent自动发送邮件给维护者邮件组
curl -X POST http://localhost:8080/api/google/emails/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["maintainers@higress.io"],
    "subject": "[Issue #123] Bug: Connection timeout in AI Gateway",
    "content": "Issue详情:\n- URL: https://github.com/alibaba/higress/issues/123\n- 标题: Bug: Connection timeout in AI Gateway\n- 内容: When using AI Gateway with large models, connections frequently timeout after 30 seconds.\n- 优先级: high\n- 标签: [bug, ai-gateway, timeout]\n\n分析结果: 需要维护者协助处理AI Gateway连接超时问题\n\n请协助处理此Issue。",
    "thread_id": ""
  }'
```

**邮件发送响应**:
```json
{
  "success": true,
  "message_id": "msg_123456",
  "thread_id": "thread_123",
  "message": "邮件发送成功"
}
```

#### 4. 维护者回复邮件
维护者通过邮件回复：

```
发件人: maintainer@higress.io
主题: Re: [Issue #123] Bug: Connection timeout in AI Gateway
内容: 这个问题是由于AI Gateway的默认超时设置过短导致的。建议将timeout配置从30秒增加到120秒，并在配置文件中添加retry机制。
```

#### 5. Agent监听并处理邮件回复
Agent监听邮件变化并处理回复：

```bash
# Agent监听邮件变化并处理回复
curl -X POST http://localhost:8080/api/google/emails/reply \
  -H "Content-Type: application/json" \
  -d '{
    "thread_id": "thread_123",
    "reply": {
      "from": "maintainer@higress.io",
      "content": "这个问题是由于AI Gateway的默认超时设置过短导致的。建议将timeout配置从30秒增加到120秒，并在配置文件中添加retry机制。",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  }'
```

#### 6. Agent结合LLM和DeepWiki进行深度分析
Agent使用LLM分析维护者回复：

```bash
# Agent使用LLM分析维护者回复
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "type": "question",
    "content": "分析维护者回复：这个问题是由于AI Gateway的默认超时设置过短导致的。建议将timeout配置从30秒增加到120秒，并在配置文件中添加retry机制。请提供具体的配置修改建议和代码示例。",
    "author": "agent"
  }'
```

Agent使用DeepWiki检索相关知识：

```bash
# Agent使用DeepWiki检索相关知识
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "input": "AI Gateway timeout configuration retry mechanism",
    "repo_name": "alibaba/higress"
  }'
```

#### 7. Agent生成完整的Issue回复
基于维护者回复和DeepWiki检索结果，Agent生成完整的Issue回复：

```bash
# Agent生成Issue回复
curl -X POST http://localhost:8080/api/google/issues/123/reply \
  -H "Content-Type: application/json" \
  -d '{
    "content": "感谢维护者的回复！\n\n根据维护者的建议和DeepWiki检索结果，这个问题可以通过以下方式解决：\n\n## 解决方案\n\n### 1. 修改超时配置\n在 `higress-config.yaml` 中修改AI Gateway配置：\n\n```yaml\nai:\n  gateway:\n    timeout: 120s  # 从30s增加到120s\n    retry:\n      max_attempts: 3\n      backoff:\n        initial_interval: 1s\n        max_interval: 10s\n```\n\n### 2. 添加重试机制\n在AI Gateway插件配置中添加重试逻辑：\n\n```yaml\nplugins:\n  - name: ai-gateway\n    config:\n      timeout: 120s\n      retry:\n        enabled: true\n        max_attempts: 3\n```\n\n### 3. 监控和日志\n建议添加以下监控指标：\n- AI Gateway请求超时次数\n- 重试成功率\n- 平均响应时间\n\n## 测试建议\n1. 使用GPT-4模型进行压力测试\n2. 验证超时配置是否生效\n3. 检查重试机制是否正常工作\n\n这个问题将在配置更新后得到解决。",
    "analysis": {
      "is_resolved": true,
      "action": "close",
      "summary": "问题已解决，提供了完整的配置修改方案"
    }
  }'
```

#### 8. 更新Issue状态
最后更新Issue状态为已解决：

```bash
# 更新Issue状态为已解决
curl -X PUT http://localhost:8080/api/google/issues/123/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "resolved"
  }'
```

## API接口详解

### Issue管理

#### 处理GitHub Issue
```http
POST /api/google/issues
Content-Type: application/json

{
  "issue_id": "123",
  "issue_url": "https://github.com/alibaba/higress/issues/123",
  "issue_title": "Bug: Connection timeout in AI Gateway",
  "issue_content": "When using AI Gateway with large models..."
}
```

**响应**:
```json
{
  "success": true,
  "analysis": {
    "can_resolve": false,
    "priority": "high",
    "tags": ["bug", "ai-gateway"],
    "summary": "需要维护者协助处理"
  },
  "issue_id": "123",
  "status": "waiting"
}
```

#### 获取Issue列表
```http
GET /api/google/issues?status=pending
```

**响应**:
```json
{
  "success": true,
  "issues": [
    {
      "issue_id": "123",
      "issue_url": "https://github.com/alibaba/higress/issues/123",
      "issue_title": "Bug: Connection timeout in AI Gateway",
      "status": "waiting",
      "priority": "high",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ],
  "count": 1
}
```

#### 获取单个Issue
```http
GET /api/google/issues/123
```

**响应**:
```json
{
  "success": true,
  "issue": {
    "issue_id": "123",
    "issue_url": "https://github.com/alibaba/higress/issues/123",
    "issue_title": "Bug: Connection timeout in AI Gateway",
    "issue_content": "When using AI Gateway...",
    "status": "waiting",
    "priority": "high",
    "tags": ["bug", "ai-gateway"],
    "email_thread_id": "thread_123",
    "maintainer_replies": [
      {
        "from": "maintainer@higress.io",
        "content": "这个问题是由于...",
        "timestamp": "2024-01-15T10:30:00Z"
      }
    ],
    "created_at": "2024-01-15T10:00:00Z",
    "last_updated": "2024-01-15T10:30:00Z"
  }
}
```

### 邮件管理

#### 发送邮件
```http
POST /api/google/emails/send
Content-Type: application/json

{
  "to": ["maintainers@higress.io"],
  "subject": "[Issue #123] Bug: Connection timeout in AI Gateway",
  "content": "Issue详情:\n- URL: https://github.com/alibaba/higress/issues/123\n...",
  "thread_id": ""
}
```

**响应**:
```json
{
  "success": true,
  "message_id": "msg_123456",
  "thread_id": "thread_123",
  "message": "邮件发送成功"
}
```

#### 处理邮件回复
```http
POST /api/google/emails/reply
Content-Type: application/json

{
  "thread_id": "thread_123",
  "reply": {
    "from": "maintainer@higress.io",
    "content": "这个问题是由于AI Gateway的默认超时设置过短导致的...",
    "timestamp": "2024-01-15T10:30:00Z"
  }
}
```

**响应**:
```json
{
  "success": true,
  "message": "邮件回复处理成功",
  "thread_id": "thread_123",
  "issue_reply": "感谢维护者的回复！\n\n根据维护者的建议..."
}
```

#### 同步邮件
```http
POST /api/google/emails/sync
```

**响应**:
```json
{
  "success": true,
  "message": "邮件同步成功",
  "timestamp": "2024-01-15T10:35:00Z",
  "synced_count": 5
}
```

### 会话管理

#### 获取会话列表
```http
GET /api/google/threads
```

**响应**:
```json
{
  "success": true,
  "threads": [
    {
      "id": "thread_123",
      "subject": "[Issue #123] Bug: Connection timeout in AI Gateway",
      "issue_id": "123",
      "status": "replied",
      "messages": [
        {
          "id": "msg_123456",
          "from": "agent@higress.io",
          "subject": "[Issue #123] Bug: Connection timeout in AI Gateway",
          "content": "Issue详情...",
          "timestamp": "2024-01-15T10:00:00Z"
        },
        {
          "id": "msg_123457",
          "from": "maintainer@higress.io",
          "subject": "Re: [Issue #123] Bug: Connection timeout in AI Gateway",
          "content": "这个问题是由于...",
          "timestamp": "2024-01-15T10:30:00Z"
        }
      ],
      "created_at": "2024-01-15T10:00:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

### 统计信息

#### 获取统计信息
```http
GET /api/google/stats
```

**响应**:
```json
{
  "success": true,
  "stats": {
    "total_issues": 10,
    "pending_issues": 3,
    "active_threads": 2,
    "total_emails": 15,
    "last_sync": "2024-01-15T10:35:00Z",
    "success_rate": 0.95
  }
}
```

## 配置说明

### Google API配置
```yaml
# configs/google_config.yaml
google:
  gmail:
    credentials_file: "credentials.json"  # 服务账号凭证文件
    token_file: "token.json"             # 访问令牌文件
    group_email: "maintainers@higress.io" # 维护者邮件组
    scopes:                              # 权限范围
      - "https://www.googleapis.com/auth/gmail.send"
      - "https://www.googleapis.com/auth/gmail.readonly"
      - "https://www.googleapis.com/auth/gmail.modify"
      - "https://www.googleapis.com/auth/gmail.labels"
      - "https://www.googleapis.com/auth/gmail.compose"
  
  groups:
    admin_email: "admin@higress.io"      # 管理员邮箱
    group_key: "maintainers@higress.io"  # 邮件组标识
    domain: "higress.io"                 # 域名

email:
  maintainers:                           # 维护者邮箱列表
    - "maintainer1@higress.io"
    - "maintainer2@higress.io"
    - "maintainer3@higress.io"
  
  templates:                             # 邮件模板
    issue_notification:
      subject: "[Issue #{issue_id}] {issue_title}"
      content: |
        Issue详情:
        - URL: {issue_url}
        - 标题: {issue_title}
        - 内容: {issue_content}
        - 优先级: {priority}
        - 标签: {tags}
        
        分析结果: {analysis}
        
        请协助处理此Issue。
    
    reply_confirmation:
      subject: "Re: [Issue #{issue_id}] {issue_title}"
      content: |
        感谢维护者的回复:
        
        {maintainer_email}
        
        维护者回复: {reply_content}
        
        状态: {status}

issue:
  auto_resolve_keywords:                 # 自动处理关键词
    - "documentation"
    - "typo"
    - "format"
  
  high_priority_keywords:                # 高优先级关键词
    - "bug"
    - "error"
    - "crash"
    - "security"
  
  feature_keywords:                      # 功能请求关键词
    - "feature"
    - "enhancement"
    - "improvement"
```

## 设置指南

### 1. 设置Google Cloud项目

1. 创建Google Cloud项目
2. 启用Gmail API和Admin Directory API
3. 创建服务账号并下载凭证文件
4. 配置必要的权限范围

### 2. 配置邮件组

1. 创建私有邮件组
2. 将Agent的服务账号添加为组成员
3. 配置邮件组权限

### 3. 配置环境变量

```bash
# 设置Google API凭证
export GOOGLE_APPLICATION_CREDENTIALS="path/to/credentials.json"

# 设置邮件组配置
export MAINTAINER_GROUP_EMAIL="maintainers@higress.io"
export ADMIN_EMAIL="admin@higress.io"
```

### 4. 启动服务

```bash
# 启动主服务
go run main.go

# 或者使用Docker
docker-compose up -d
```

## 监控和调试

### 1. 查看处理状态
```bash
# 获取所有Issue状态
curl -X GET http://localhost:8080/api/google/issues

# 获取统计信息
curl -X GET http://localhost:8080/api/google/stats
```

### 2. 查看邮件会话
```bash
# 获取所有邮件会话
curl -X GET http://localhost:8080/api/google/threads

# 获取特定会话
curl -X GET http://localhost:8080/api/google/threads/thread_123
```

### 3. 手动同步邮件
```bash
# 手动同步邮件
curl -X POST http://localhost:8080/api/google/emails/sync
```

### 4. 查看日志
```bash
# 查看服务日志
docker logs community-governance-agent

# 或者查看本地日志
tail -f logs/agent.log
```

## 最佳实践

### 1. Issue分析
- 使用关键词分析自动识别问题类型
- 结合LLM进行深度语义分析
- 利用DeepWiki检索相关知识

### 2. 邮件处理
- 使用模板确保邮件格式一致
- 设置合适的超时和重试机制
- 定期同步邮件状态

### 3. 状态管理
- 实时跟踪Issue处理状态
- 维护Issue与邮件的映射关系
- 定期清理已完成的会话

### 4. 监控告警
- 监控邮件发送成功率
- 设置Issue处理超时告警
- 跟踪维护者响应时间

## 故障排除

### 1. 认证问题
```bash
# 检查Google API凭证
gcloud auth application-default print-access-token

# 验证邮件组权限
curl -H "Authorization: Bearer $(gcloud auth print-access-token)" \
  https://admin.googleapis.com/admin/directory/v1/groups/maintainers@higress.io
```

### 2. 邮件发送失败
```bash
# 检查邮件组配置
curl -X GET http://localhost:8080/api/google/groups/maintainers@higress.io

# 测试邮件发送
curl -X POST http://localhost:8080/api/google/emails/send \
  -H "Content-Type: application/json" \
  -d '{
    "to": ["test@example.com"],
    "subject": "Test Email",
    "content": "This is a test email"
  }'
```

### 3. 邮件同步失败
```bash
# 检查邮件监听状态
curl -X GET http://localhost:8080/api/google/watch/status

# 重新启动监听
curl -X POST http://localhost:8080/api/google/watch \
  -H "Content-Type: application/json" \
  -d '{"topic_name": "projects/your-project/topics/gmail-notifications"}'
```

## 扩展开发

### 1. 添加新的邮件模板
```yaml
email:
  templates:
    custom_template:
      subject: "[Custom] {issue_title}"
      content: |
        自定义模板内容
        - Issue ID: {issue_id}
        - 标题: {issue_title}
```

### 2. 自定义Issue分析规则
```go
// 在 internal/google/manager.go 中添加自定义分析逻辑
func (m *GoogleManager) analyzeIssue(content string) (*IssueAnalysis, error) {
    // 添加自定义分析规则
    if containsKeywords(content, []string{"custom_keyword"}) {
        analysis.Priority = "custom"
        analysis.Tags = append(analysis.Tags, "custom_tag")
    }
    return analysis, nil
}
```

### 3. 添加新的API端点
```go
// 在 internal/google/handler.go 中添加新的处理器
func (h *GoogleHandler) CustomEndpoint(w http.ResponseWriter, r *http.Request) {
    // 实现自定义逻辑
}
```

这个完整的Google API使用指南提供了详细的流程说明、API接口文档、配置指南和最佳实践，帮助用户理解和使用Google API集成功能。 