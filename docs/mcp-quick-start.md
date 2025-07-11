# MCP快速开始指南

## 快速配置

### 1. 启用MCP功能
在 `configs/config.yaml` 中启用MCP：

```yaml
mcp:
  enabled: true
  timeout: "30s"
  servers:
    deepwiki:
      enabled: true
      server_url: "https://mcp.deepwiki.com/mcp"
      server_label: "deepwiki"
      require_approval: "never"
```

### 2. 启动服务
```bash
go build -o bin/agent cmd/agent/main.go
./bin/agent
```

## 基本使用

### 查询DeepWiki
```bash
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "input": "What is the MCP protocol?",
    "repo_name": "modelcontextprotocol/modelcontextprotocol"
  }'
```

### 获取工具列表
```bash
curl -X POST http://localhost:8080/api/v1/mcp/tools \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "server_url": "https://mcp.deepwiki.com/mcp"
  }'
```

### 调用特定工具
```bash
curl -X POST http://localhost:8080/api/v1/mcp/call \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "tool_name": "ask_question",
    "arguments": {
      "question": "What is the MCP protocol?",
      "repo_name": "modelcontextprotocol/modelcontextprotocol"
    }
  }'
```

## 配置代理

如果遇到网络访问问题，可以配置代理：

```yaml
network:
  proxy_enabled: true
  proxy_url: "http://proxy.example.com:8080"
  proxy_type: "http"
```

## 健康检查

```bash
curl http://localhost:8080/api/v1/health
```

## 常见问题

### 1. 连接超时
- 检查网络连接
- 配置代理设置
- 增加超时时间

### 2. 认证失败
- 检查API密钥
- 验证服务器URL
- 确认权限设置

### 3. 工具调用失败
- 检查工具名称
- 验证参数格式
- 查看错误日志

## 更多信息

详细文档请参考：[MCP集成指南](mcp-integration.md) 