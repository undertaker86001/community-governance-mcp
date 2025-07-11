# MCP集成指南

## 概述

本项目支持集成远程MCP服务器，通过统一的MCP管理器调用外部工具和服务。重构后的架构提供了更好的可维护性和扩展性。

## 架构设计

### 统一MCP架构
```
MCP集成层
├── MCP管理器 (Manager)
│   ├── 客户端管理
│   ├── 配置管理
│   ├── 健康检查
│   └── 缓存管理
├── MCP客户端 (Client)
│   ├── 工具列表获取
│   ├── 工具调用
│   └── 查询执行
└── HTTP处理器
    ├── 查询接口
    ├── 工具列表接口
    └── 工具调用接口
```

### 核心组件

#### 1. MCP管理器 (`internal/mcp/manager.go`)
- **功能**: 统一管理所有MCP服务器连接
- **特性**: 
  - 支持多服务器配置
  - 内置健康检查
  - 备用方案支持
  - 错误处理和重试机制

#### 2. MCP客户端 (`internal/mcp/client.go`)
- **功能**: 处理与MCP服务器的通信
- **特性**:
  - 支持JSON-RPC 2.0协议
  - 超时控制
  - 请求头管理
  - 响应解析

#### 3. HTTP处理器 (`cmd/agent/main.go`)
- **功能**: 提供RESTful API接口
- **端点**:
  - `POST /api/v1/mcp/query` - 执行MCP查询
  - `POST /api/v1/mcp/tools` - 获取工具列表
  - `POST /api/v1/mcp/call` - 调用特定工具

## 支持的MCP服务器

### 1. DeepWiki MCP服务器
- **服务器URL**: `https://mcp.deepwiki.com/mcp`
- **功能**: 查询GitHub仓库信息和文档
- **认证**: 无需认证
- **工具**: `ask_question`, `read_wiki_structure`



## 配置说明

### 基本配置
```yaml
# configs/config.yaml
mcp:
  enabled: true
  timeout: "30s"
  servers:
    deepwiki:
      enabled: true
      server_url: "https://mcp.deepwiki.com/mcp"
      server_label: "deepwiki"
      require_approval: "never"
      allowed_tools: ["ask_question", "read_wiki_structure"]
    
    stripe:
      enabled: false
      server_url: "https://mcp.stripe.com"
      server_label: "stripe"
      require_approval: "always"
      headers:
        Authorization: "${STRIPE_API_KEY}"
```

### 配置字段说明

| 字段 | 类型 | 说明 | 示例 |
|------|------|------|------|
| `enabled` | boolean | 是否启用MCP功能 | `true` |
| `timeout` | string | 请求超时时间 | `"30s"` |
| `servers` | object | 服务器配置映射 | - |
| `server_url` | string | MCP服务器URL | `"https://mcp.deepwiki.com/mcp"` |
| `server_label` | string | 服务器标签 | `"deepwiki"` |
| `require_approval` | string | 审批要求 | `"never"`, `"always"` |
| `allowed_tools` | array | 允许的工具列表 | `["ask_question"]` |
| `headers` | object | 请求头配置 | `{"Authorization": "Bearer token"}` |

### 认证配置
```yaml
mcp:
  servers:
    stripe:
      enabled: true
      server_url: "https://mcp.stripe.com"
      server_label: "stripe"
      headers:
        Authorization: "Bearer ${STRIPE_API_KEY}"
      require_approval:
        never:
          tool_names: ["create_payment_link"]
```

## API使用示例

### 1. 查询GitHub仓库信息
```bash
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "input": "What transport protocols are supported in the 2025-03-26 version of the MCP spec?",
    "repo_name": "modelcontextprotocol/modelcontextprotocol"
  }'
```

**响应示例**:
```json
{
  "output": "The MCP spec supports the following transport protocols...",
  "error": null
}
```

### 2. 调用特定工具
```bash
curl -X POST http://localhost:8080/api/v1/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "server_url": "https://mcp.deepwiki.com/mcp",
    "tool_name": "ask_question",
    "arguments": {
      "question": "What is the MCP protocol?",
      "repo_name": "modelcontextprotocol/modelcontextprotocol"
    }
  }'
```



## 集成到处理器

### 在知识检索中使用MCP
```go
// 使用MCP管理器进行查询
items, err := p.mcpManager.QueryWithFallback(
    ctx,
    "deepwiki",
    question.Title+" "+question.Content,
    "modelcontextprotocol/modelcontextprotocol",
    func() ([]model.KnowledgeItem, error) {
        // 备用方案：直接HTTP调用
        return p.retrieveFromDeepWikiHTTP(ctx, question)
    },
)
```

### 备用方案支持
- **MCP查询失败**: 自动切换到HTTP调用
- **HTTP调用失败**: 使用备用数据
- **完整错误处理**: 详细的日志记录

## 安全注意事项

### 1. 服务器信任
- 优先使用官方MCP服务器
- 避免使用第三方代理服务器
- 仔细审查服务器提供商的信誉

### 2. 数据保护
- 启用审批机制审查数据共享
- 记录所有MCP服务器交互
- 定期审查数据使用情况

### 3. 认证管理
- 使用环境变量存储敏感密钥
- 定期轮换API密钥
- 监控异常访问模式

### 4. 网络安全
- 支持代理配置解决网络访问问题
- 超时控制防止长时间等待
- 重试机制处理临时网络问题

## 错误处理

### 常见错误
1. **连接超时**: 检查网络连接和服务器可用性
2. **认证失败**: 验证API密钥和权限
3. **工具调用失败**: 检查工具参数和服务器响应
4. **JSON解析错误**: 检查响应格式

### 调试方法
```bash
# 启用详细日志
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -H "X-Debug: true" \
  -d '{
    "server_label": "deepwiki",
    "input": "test query"
  }'
```

### 错误响应格式
```json
{
  "error": "错误描述",
  "message": "详细错误信息"
}
```

## 性能优化

### 1. 工具过滤
```yaml
mcp:
  servers:
    deepwiki:
      allowed_tools: ["ask_question"]  # 只允许特定工具
```

### 2. 超时控制
```yaml
mcp:
  timeout: "30s"  # 全局超时设置
```

### 3. 并发控制
- 内置连接池管理
- 请求限流
- 错误重试机制

## 监控和日志

### 1. 指标监控
- MCP服务器响应时间
- 工具调用成功率
- 错误率和类型统计

### 2. 日志记录
```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "./logs/agent.log"
```

### 3. 健康检查
```bash
# 检查MCP服务器健康状态
curl -X GET http://localhost:8080/api/v1/health
```

## 最佳实践

### 1. 渐进式采用
- 从可信的MCP服务器开始
- 逐步启用更多功能
- 充分测试后再部署到生产环境

### 2. 审批流程
- 为敏感操作启用审批机制
- 记录所有MCP服务器交互
- 定期审查使用情况

### 3. 监控告警
- 设置异常行为告警
- 监控响应时间和错误率
- 定期检查服务器可用性

### 4. 备份方案
- 为关键功能提供备用方案
- 实现优雅的降级策略
- 保持数据一致性

### 5. 网络配置
```yaml
network:
  proxy_enabled: true
  proxy_url: "http://proxy.example.com:8080"
  proxy_type: "http"
```

## 故障排除

### 1. 连接问题
```bash
# 测试MCP服务器连接
curl -X GET https://mcp.deepwiki.com/mcp/health

# 检查网络代理
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "input": "test"
  }'
```

### 2. 认证问题
```bash
# 验证API密钥
curl -H "Authorization: Bearer $API_KEY" \
  https://mcp.stripe.com/health

# 检查环境变量
echo $STRIPE_API_KEY
```

### 3. 工具调用问题
```bash
# 获取工具列表
curl -X POST http://localhost:8080/api/v1/mcp/tools \
  -H "Content-Type: application/json" \
  -d '{"server_label": "deepwiki"}'

# 测试工具调用
curl -X POST http://localhost:8080/api/v1/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "tool_name": "ask_question",
    "arguments": {"question": "test"}
  }'
```

### 4. 配置问题
```bash
# 检查配置加载
curl -X GET http://localhost:8080/api/v1/config

# 验证MCP配置
grep -A 10 "mcp:" configs/config.yaml
```

## 开发指南

### 添加新的MCP服务器

1. **更新配置**
```yaml
mcp:
  servers:
    new_server:
      enabled: true
      server_url: "https://mcp.newserver.com"
      server_label: "new_server"
      require_approval: "always"
      headers:
        Authorization: "${NEW_SERVER_API_KEY}"
```

2. **测试连接**
```bash
curl -X POST http://localhost:8080/api/v1/mcp/tools \
  -H "Content-Type: application/json" \
  -d '{"server_label": "new_server"}'
```

3. **使用API**
```bash
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "new_server",
    "input": "your query here"
  }'
```

### 扩展MCP管理器

1. **添加新方法**
```go
// 在 internal/mcp/manager.go 中添加新方法
func (m *Manager) CustomQuery(ctx context.Context, serverLabel, input string) (*QueryResponse, error) {
    // 实现自定义查询逻辑
}
```

2. **更新处理器**
```go
// 在 internal/agent/processor.go 中使用新方法
result, err := p.mcpManager.CustomQuery(ctx, "deepwiki", question.Content)
```

## 版本历史

### v1.0.0 (当前版本)
- ✅ 统一MCP管理器架构
- ✅ 支持多服务器配置
- ✅ 完整的备用方案
- ✅ 统一的API接口
- ✅ 网络代理支持
- ✅ 健康检查和监控

### 计划功能
- 🔄 缓存机制优化
- 🔄 更多MCP服务器支持
- �� 高级审批流程
- 🔄 性能指标监控 