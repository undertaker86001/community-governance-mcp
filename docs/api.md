# Higress社区治理智能助手 API 文档

## 概述

Higress社区治理智能助手提供RESTful API接口，支持智能问答、问题分析、社区统计等功能。

## 基础信息

- **基础URL**: `http://localhost:8080`
- **API版本**: `v1`
- **内容类型**: `application/json`
- **认证**: 目前无需认证

## 通用响应格式

所有API响应都遵循以下格式：

```json
{
  "status": "success|error",
  "data": {},
  "message": "响应消息",
  "timestamp": 1640995200
}
```

## API端点

### 1. 智能问答

#### POST /api/v1/process

处理用户问题，返回智能回答。

**请求参数:**

```json
{
  "title": "Higress网关配置问题",
  "content": "如何配置Higress网关的路由规则？",
  "author": "user123",
  "type": "question",
  "priority": "normal",
  "tags": ["gateway", "configuration"],
  "metadata": {
    "source": "web",
    "session_id": "abc123"
  }
}
```

**响应示例:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "question_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "基于Higress官方文档，配置路由规则的步骤如下：\n\n1. 创建Ingress资源\n2. 配置路由规则\n3. 设置后端服务\n\n详细配置请参考官方文档...",
  "summary": "Higress网关路由配置方法",
  "sources": [
    {
      "id": "source-1",
      "title": "Higress官方文档",
      "url": "https://higress.io/docs",
      "source": "official",
      "relevance": 0.95,
      "confidence": 0.9,
      "last_updated": "2024-01-15T10:30:00Z"
    }
  ],
  "confidence": 0.92,
  "processing_time": "1.2s",
  "fusion_score": 0.88,
  "recommendations": [
    "建议查看官方文档获取详细配置",
    "可以尝试使用Higress控制台进行可视化配置"
  ]
}
```

### 2. 问题分析

#### POST /api/v1/analyze

分析错误堆栈和问题信息。

**请求参数:**

```json
{
  "stack_trace": "panic: runtime error: invalid memory address or nil pointer dereference\n\tat main.main()\n\t\t./main.go:10 +0x1a",
  "environment": "Go 1.21, Linux, Kubernetes 1.24",
  "version": "Higress 1.0.0",
  "image_url": "https://example.com/error-screenshot.png",
  "issue_type": "bug",
  "metadata": {
    "component": "gateway",
    "severity": "high"
  }
}
```

**响应示例:**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440002",
  "problem_type": "bug",
  "severity": "critical",
  "diagnosis": "空指针异常，变量未正确初始化",
  "solutions": [
    "检查变量初始化，确保在使用前已正确赋值",
    "添加空值检查，避免直接访问可能为空的变量",
    "使用安全的访问方法，如可选链操作符"
  ],
  "confidence": 0.85,
  "processing_time": "0.8s",
  "related_issues": [
    "https://github.com/alibaba/higress/issues/123",
    "https://github.com/alibaba/higress/issues/456"
  ]
}
```

### 3. 社区统计

#### GET /api/v1/stats

获取社区活跃度统计。

**响应示例:**

```json
{
  "total_issues": 1250,
  "open_issues": 89,
  "closed_issues": 1161,
  "total_prs": 456,
  "open_prs": 23,
  "merged_prs": 433,
  "contributors": 156,
  "active_users": 89,
  "top_contributors": [
    {
      "username": "higress-maintainer",
      "avatar_url": "https://github.com/higress-maintainer.png",
      "contributions": 45,
      "issues": 12,
      "prs": 33
    }
  ],
  "issue_trends": [
    {
      "date": "2024-01-15",
      "opened": 5,
      "closed": 8
    }
  ],
  "pr_trends": [
    {
      "date": "2024-01-15",
      "opened": 2,
      "merged": 3
    }
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

### 4. 健康检查

#### GET /api/v1/health

检查服务状态。

**响应示例:**

```json
{
  "status": "healthy",
  "timestamp": 1640995200,
  "version": "1.0.0",
  "services": {
    "deepwiki": true,
    "knowledge": true,
    "fusion": true
  }
}
```

### 5. 配置信息

#### GET /api/v1/config

获取服务配置信息（不包含敏感数据）。

**响应示例:**

```json
{
  "name": "higress-community-agent",
  "version": "1.0.0",
  "port": 8080,
  "debug": true,
  "features": {
    "deepwiki_enabled": true,
    "knowledge_enabled": true,
    "fusion_enabled": true
  }
}
```

## 错误处理

### 错误响应格式

```json
{
  "error": "错误类型",
  "message": "详细错误信息",
  "timestamp": 1640995200,
  "request_id": "req-123"
}
```

### 常见错误码

| 状态码 | 错误类型 | 描述 |
|--------|----------|------|
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 未授权访问 |
| 403 | Forbidden | 禁止访问 |
| 404 | Not Found | 资源未找到 |
| 429 | Too Many Requests | 请求频率过高 |
| 500 | Internal Server Error | 服务器内部错误 |

## 使用示例

### cURL示例

#### 智能问答

```bash
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Higress配置问题",
    "content": "如何配置Higress的SSL证书？",
    "author": "developer",
    "type": "question"
  }'
```

#### 问题分析

```bash
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "stack_trace": "panic: runtime error: invalid memory address",
    "environment": "Go 1.21, Linux",
    "issue_type": "bug"
  }'
```

#### 健康检查

```bash
curl http://localhost:8080/api/v1/health
```

### JavaScript示例

```javascript
// 智能问答
const response = await fetch('http://localhost:8080/api/v1/process', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    title: 'Higress配置问题',
    content: '如何配置Higress的SSL证书？',
    author: 'developer',
    type: 'question'
  })
});

const result = await response.json();
console.log(result);
```

### Python示例

```python
import requests

# 智能问答
response = requests.post('http://localhost:8080/api/v1/process', json={
    'title': 'Higress配置问题',
    'content': '如何配置Higress的SSL证书？',
    'author': 'developer',
    'type': 'question'
})

result = response.json()
print(result)
```

## 速率限制

- 每个IP每分钟最多100个请求
- 每个用户每分钟最多50个请求
- 超过限制将返回429状态码

## 数据格式

### 时间格式

所有时间字段使用ISO 8601格式：`YYYY-MM-DDTHH:mm:ssZ`

### 优先级枚举

- `low`: 低优先级
- `normal`: 普通优先级
- `high`: 高优先级
- `urgent`: 紧急优先级
- `critical`: 严重优先级

### 问题类型枚举

- `text`: 文本问题
- `issue`: GitHub Issue
- `pr`: Pull Request
- `bug`: Bug报告
- `config`: 配置问题

## 更新日志

### v1.0.0 (2024-01-15)

- 初始版本发布
- 支持智能问答功能
- 支持问题分析功能
- 支持社区统计功能
- 支持健康检查

## 支持

如有问题或建议，请通过以下方式联系：

- 提交GitHub Issue
- 发送邮件至项目维护者
- 查看项目文档 