# Higress社区治理Agent API使用示例

## 快速开始

### 1. 启动Agent
```bash
# 设置环境变量
export OPENAI_API_KEY="your-openai-api-key"

# 启动Agent
go run cmd/agent/main.go
```

### 2. 测试健康检查
```bash
curl http://localhost:8080/api/v1/health
```

响应：
```json
{
  "status": "healthy",
  "timestamp": 1703123456,
  "version": "1.0.0"
}
```

## API接口示例

### 处理Issue问题

**请求：**
```bash
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "type": "issue",
    "title": "Gateway配置问题",
    "content": "我在配置Higress Gateway时遇到了路由问题，具体错误是：404 Not Found。我的配置如下：...",
    "author": "developer123",
    "priority": "medium",
    "tags": ["gateway", "routing", "404"]
  }'
```

**响应：**
```json
{
  "id": "resp-123456",
  "question_id": "q-789012",
  "content": "## 问题分析\n\n您的问题是关于 **Gateway配置问题** 的 issue 类型问题。\n\n## 解决方案\n\n### 来源 1: DeepWiki回答\n\n根据Higress官方文档，404错误通常由以下原因引起：\n\n1. **路由配置错误**：检查路由规则是否正确配置\n2. **服务发现问题**：确保后端服务正常运行\n3. **路径匹配问题**：验证请求路径是否与路由规则匹配\n\n### 来源 2: 本地知识库\n\n常见解决方案：\n- 检查Gateway配置文件\n- 验证后端服务状态\n- 查看Gateway日志\n\n## 建议\n\n- 请根据您的具体环境调整配置\n- 建议查看官方文档获取最新信息\n- 如有疑问，欢迎在社区讨论\n",
  "summary": "## 问题分析\n\n您的问题是关于 **Gateway配置问题** 的 issue 类型问题。\n\n## 解决方案\n\n### 来源 1: DeepWiki回答\n\n根据Higress官方文档，404错误通常由以下原因引起：\n\n1. **路由配置错误**：检查路由规则是否正确配置\n2. **服务发现问题**：确保后端服务正常运行\n3. **路径匹配问题**：验证请求路径是否与路由规则匹配\n\n### 来源 2: 本地知识库\n\n常见解决方案：\n- 检查Gateway配置文件\n- 验证后端服务状态\n- 查看Gateway日志\n\n## 建议\n\n- 请根据您的具体环境调整配置\n- 建议查看官方文档获取最新信息\n- 如有疑问，欢迎在社区讨论\n",
  "sources": [
    {
      "id": "source-1",
      "source": "deepwiki",
      "title": "DeepWiki回答",
      "content": "根据Higress官方文档，404错误通常由以下原因引起：\n\n1. **路由配置错误**：检查路由规则是否正确配置\n2. **服务发现问题**：确保后端服务正常运行\n3. **路径匹配问题**：验证请求路径是否与路由规则匹配",
      "url": "https://github.com/alibaba/higress",
      "relevance": 0.9,
      "tags": ["gateway", "routing", "404"],
      "created_at": "2024-01-01T12:00:00Z",
      "metadata": {
        "repo_name": "alibaba/higress",
        "question": "Gateway配置问题: 我在配置Higress Gateway时遇到了路由问题，具体错误是：404 Not Found。我的配置如下：..."
      }
    }
  ],
  "confidence": 0.92,
  "processing_time": "2.5s",
  "fusion_score": 0.85,
  "recommendations": [
    "建议提供详细的错误日志和复现步骤",
    "检查是否是最新版本的问题",
    "考虑在GitHub上创建Issue"
  ]
}
```

### 处理PR问题

**请求：**
```bash
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "type": "pr",
    "title": "Add new plugin feature",
    "content": "我添加了一个新的插件功能，包括配置验证和错误处理。主要变更包括：1. 新增插件接口 2. 添加配置验证 3. 完善错误处理",
    "author": "contributor456",
    "priority": "medium",
    "tags": ["plugin", "feature", "validation"]
  }'
```

### 处理图文问题

**请求：**
```bash
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text",
    "title": "Kubernetes部署问题",
    "content": "我想了解如何在Kubernetes中部署Higress，需要哪些配置文件和步骤？",
    "author": "k8s-user",
    "priority": "low",
    "tags": ["kubernetes", "deployment", "helm"]
  }'
```

## 错误处理

### 请求格式错误
```bash
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试问题"
  }'
```

**响应：**
```json
{
  "error": "请求验证失败",
  "message": "内容不能为空"
}
```

### 服务器错误
```json
{
  "error": "问题处理失败",
  "message": "DeepWiki查询失败: API request failed with status: 401"
}
```

## 配置信息

### 获取配置
```bash
curl http://localhost:8080/api/v1/config
```

**响应：**
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

## 使用场景示例

### 1. GitHub Issue自动处理
```python
import requests
import json

def process_github_issue(issue_data):
    """处理GitHub Issue"""
    payload = {
        "type": "issue",
        "title": issue_data["title"],
        "content": issue_data["body"],
        "author": issue_data["user"]["login"],
        "priority": "medium",
        "tags": issue_data.get("labels", [])
    }
    
    response = requests.post(
        "http://localhost:8080/api/v1/process",
        json=payload
    )
    
    return response.json()
```

### 2. 社区论坛集成
```python
def process_community_question(question_data):
    """处理社区问题"""
    payload = {
        "type": "text",
        "title": question_data["title"],
        "content": question_data["content"],
        "author": question_data["author"],
        "priority": "low",
        "tags": question_data.get("tags", [])
    }
    
    response = requests.post(
        "http://localhost:8080/api/v1/process",
        json=payload
    )
    
    return response.json()
```

### 3. 批量处理
```python
def batch_process_questions(questions):
    """批量处理问题"""
    results = []
    
    for question in questions:
        try:
            response = requests.post(
                "http://localhost:8080/api/v1/process",
                json=question,
                timeout=30
            )
            results.append(response.json())
        except Exception as e:
            results.append({"error": str(e)})
    
    return results
```

## 监控和日志

### 查看日志
```bash
# 如果配置了文件日志
tail -f ./logs/agent.log

# 查看JSON格式日志
tail -f ./logs/agent.log | jq
```

### 性能监控
```bash
# 检查处理时间
curl -w "@curl-format.txt" -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{"type":"text","title":"test","content":"test"}'
```

## 最佳实践

1. **设置合理的超时时间**：建议30秒
2. **处理错误响应**：检查HTTP状态码和错误信息
3. **缓存结果**：对于重复问题可以缓存回答
4. **监控性能**：关注处理时间和成功率
5. **安全考虑**：在生产环境中使用HTTPS 