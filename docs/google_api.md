# Google API 集成文档

## 概述

本项目集成了Google API，实现了Agent作为Google账号加入私有邮件组，分析GitHub Issue，无法解决的问题通过邮件组与维护者交流，维护者回复后Agent理解并回复Issue的功能。

## 功能特性

### 核心功能
- **GitHub Issue分析**: 自动分析GitHub Issue内容，判断是否可以自动解决
- **邮件组集成**: Agent作为成员加入私有邮件组，与维护者进行交流
- **多问题并行处理**: 支持同时处理多个Issue，每个Issue创建独立的邮件会话
- **邮件会话管理**: 维护Issue与邮件会话的映射关系，跟踪处理状态
- **自动回复生成**: 基于维护者回复自动生成Issue回复内容

### 技术特性
- **Google Gmail API**: 发送和接收邮件，管理邮件会话
- **Google Groups API**: 管理邮件组成员，获取组信息
- **实时监听**: 监听邮件变化，及时处理新邮件
- **状态跟踪**: 完整的Issue处理状态跟踪和统计
- **数据备份**: 支持数据导出、导入和备份功能

## 架构设计

### 组件结构
```
Google API 集成
├── internal/google/
│   ├── types.go          # 类型定义
│   ├── gmail_client.go   # Gmail API客户端
│   ├── groups_client.go  # Groups API客户端
│   ├── manager.go        # 管理器
│   └── handler.go        # HTTP处理器
├── tools/
│   └── google_tools.go   # 工具函数
├── configs/
│   └── google_config.yaml # 配置文件
└── test/
    └── google_api_test.go # 测试文件
```

### 数据流
1. **Issue接收** → 2. **内容分析** → 3. **邮件发送** → 4. **回复监听** → 5. **回复处理** → 6. **Issue回复**

## API接口

### Issue管理

#### 处理GitHub Issue
```http
POST /api/google/issues
Content-Type: application/json

{
  "issue_id": "123",
  "issue_url": "https://github.com/test/repo/issues/123",
  "issue_title": "Bug Report",
  "issue_content": "This is a bug description"
}
```

#### 获取Issue列表
```http
GET /api/google/issues?status=pending
```

#### 获取单个Issue
```http
GET /api/google/issues/{id}
```

#### 更新Issue状态
```http
PUT /api/google/issues/{id}/status
Content-Type: application/json

{
  "status": "resolved"
}
```

### 邮件管理

#### 发送邮件
```http
POST /api/google/emails/send
Content-Type: application/json

{
  "to": ["maintainers@example.com"],
  "subject": "Issue Report",
  "content": "Issue content",
  "thread_id": "optional_thread_id"
}
```

#### 获取邮件列表
```http
GET /api/google/emails?query=is:unread&max_results=50
```

#### 同步邮件
```http
POST /api/google/emails/sync
```

#### 处理邮件回复
```http
POST /api/google/emails/reply
Content-Type: application/json

{
  "thread_id": "thread_id",
  "reply": {
    "from": "maintainer@example.com",
    "content": "Reply content",
    "timestamp": "2024-01-01T00:00:00Z"
  }
}
```

### 会话管理

#### 获取会话列表
```http
GET /api/google/threads
```

#### 获取单个会话
```http
GET /api/google/threads/{id}
```

### 统计信息

#### 获取统计信息
```http
GET /api/google/stats
```

### 监听管理

#### 开始监听
```http
POST /api/google/watch
Content-Type: application/json

{
  "topic_name": "projects/your-project/topics/gmail-notifications"
}
```

#### 停止监听
```http
DELETE /api/google/watch
```

## 配置说明

### Google API配置
```yaml
google:
  gmail:
    credentials_file: "credentials.json"  # 服务账号凭证文件
    token_file: "token.json"             # 访问令牌文件
    group_email: "maintainers@example.com" # 邮件组地址
    scopes:                              # 权限范围
      - "https://www.googleapis.com/auth/gmail.send"
      - "https://www.googleapis.com/auth/gmail.readonly"
      - "https://www.googleapis.com/auth/gmail.modify"
  
  groups:
    admin_email: "admin@example.com"     # 管理员邮箱
    group_key: "maintainers@example.com" # 邮件组标识
    domain: "example.com"                # 域名
```

### 邮件处理配置
```yaml
email:
  maintainers:                           # 维护者邮箱列表
    - "maintainer1@example.com"
    - "maintainer2@example.com"
  
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
```

### Issue处理配置
```yaml
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

## 使用指南

### 1. 设置Google API凭证

1. 创建Google Cloud项目
2. 启用Gmail API和Admin Directory API
3. 创建服务账号并下载凭证文件
4. 配置必要的权限范围

### 2. 配置邮件组

1. 创建私有邮件组
2. 将Agent的服务账号添加为组成员
3. 配置邮件组权限

### 3. 启动服务

```bash
# 启动主服务
go run main.go

# 或者使用Docker
docker-compose up
```

### 4. 处理Issue

```bash
# 发送Issue处理请求
curl -X POST http://localhost:8080/api/google/issues \
  -H "Content-Type: application/json" \
  -d '{
    "issue_id": "123",
    "issue_url": "https://github.com/test/repo/issues/123",
    "issue_title": "Bug Report",
    "issue_content": "This is a bug description"
  }'
```

### 5. 监控状态

```bash
# 获取统计信息
curl http://localhost:8080/api/google/stats

# 获取待处理Issue
curl http://localhost:8080/api/google/issues?status=pending
```

## 状态管理

### Issue状态
- `new`: 新建
- `analyzing`: 分析中
- `waiting`: 等待回复
- `replied`: 已回复
- `resolved`: 已解决
- `closed`: 已关闭

### 会话状态
- `pending`: 等待回复
- `replied`: 已回复
- `resolved`: 已解决
- `closed`: 已关闭

## 错误处理

### 常见错误
1. **凭证错误**: 检查服务账号凭证文件
2. **权限不足**: 确保配置了正确的权限范围
3. **邮件组不存在**: 检查邮件组配置
4. **网络连接**: 检查网络连接和防火墙设置

### 调试方法
1. 查看日志文件: `logs/google_api.log`
2. 检查API响应: 使用API测试工具
3. 验证配置: 检查配置文件格式

## 性能优化

### 建议配置
- 设置合适的邮件同步间隔
- 配置邮件查询过滤器
- 启用数据压缩
- 使用连接池

### 监控指标
- 邮件处理延迟
- API调用成功率
- 内存使用情况
- 并发处理数量

## 安全考虑

### 认证授权
- 使用服务账号进行API认证
- 配置最小权限原则
- 定期轮换凭证

### 数据保护
- 加密敏感数据
- 限制API访问范围
- 记录审计日志

## 扩展功能

### 计划功能
- 支持更多邮件服务商
- 集成更多Issue平台
- 添加机器学习分析
- 支持多语言处理

### 自定义开发
- 添加自定义邮件模板
- 实现自定义分析规则
- 集成第三方服务

## 故障排除

### 常见问题
1. **邮件发送失败**: 检查Gmail API配置
2. **无法接收邮件**: 验证邮件组权限
3. **状态同步错误**: 检查网络连接
4. **性能问题**: 优化查询和缓存

### 日志分析
```bash
# 查看错误日志
grep "ERROR" logs/google_api.log

# 查看性能日志
grep "latency" logs/google_api.log

# 查看API调用日志
grep "API" logs/google_api.log
```

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 支持基本的Issue处理和邮件集成
- 实现邮件会话管理
- 添加统计和监控功能

### 计划更新
- 支持更多邮件格式
- 优化AI分析算法
- 添加Web界面
- 支持集群部署 