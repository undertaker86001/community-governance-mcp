# Community Governance MCP Higress

一个融合了DeepWiki知识检索和Higress社区治理工具的智能助手项目，基于MCP协议，提供智能问答、问题分析、社区统计等功能，并集成了Google API实现Agent与维护者的邮件交流功能。

## 项目重构记录

### 2025年重构成果
本次重构成功解决了项目中的循环依赖问题，优化了代码架构：

#### 重构内容
1. **类型定义统一**: 将所有类型定义迁移到`internal/model`包，避免循环依赖
2. **工具加载器重构**: 将`tools/load_tools.go`迁移到`internal/agent/tool_loader.go`
3. **依赖关系优化**: 
   - `tools`包依赖`internal/model`包
   - `internal/agent`包依赖`internal/model`包
   - 彻底解决了`tools`和`agent`包之间的循环依赖
4. **类型补全**: 补全了所有缺失的类型定义，包括：
   - `Document`、`GitHubIssue`、`GitHubComment`、`Repository`
   - `SearchResult`、`KnowledgeSearchResult`、`ClassificationStats`
   - `IssueClassification`、`IssueInfo`、`Config`等
5. **函数重复解决**: 解决了`getString`、`getFloat`等函数的重复定义问题
6. **Import清理**: 清理了所有未使用的import语句

#### 架构优化
- **模块化设计**: 清晰的模块分离，每个包职责明确
- **依赖管理**: 统一的依赖管理，避免循环依赖
- **类型安全**: 完整的类型定义，确保编译时类型检查
- **代码质量**: 清理了所有编译错误和警告

#### 编译状态
- ✅ 项目可正常编译 (`go build ./...`)
- ✅ 依赖关系正确 (`go mod tidy`)
- ✅ 无循环依赖问题
- ✅ 无未使用import警告

### 2025年MCP集成重构
本次重构统一了MCP集成架构，消除了代码冗余：

#### 重构成果
1. **统一MCP管理器**: 创建了`internal/mcp/manager.go`，集中管理所有MCP服务器
2. **架构优化**: 
   - 处理器集成MCP管理器
   - 统一配置管理
   - 保留记忆管理功能
3. **功能增强**:
   - 支持多MCP服务器配置
   - 完整的备用方案和错误处理
   - 统一的API接口
4. **代码清理**: 移除了冗余的MCP调用代码，统一响应解析逻辑

#### 新增功能
- **统一MCP查询接口**: 支持多种MCP服务器
- **备用方案支持**: MCP失败时自动切换到HTTP调用
- **配置统一管理**: 所有MCP配置集中在`mcp`节点下
- **API接口增强**: 新增MCP查询、工具列表、工具调用接口

#### 技术改进
- **错误处理**: 统一的错误处理机制和详细日志
- **性能优化**: 连接池管理、超时控制、缓存机制
- **可扩展性**: 模块化设计，支持插件式扩展
- **安全性**: 认证头管理、敏感信息保护

#### 编译状态
- ✅ 项目成功编译 (`go build cmd/agent/main.go`)
- ✅ 所有依赖正确管理
- ✅ 无循环依赖问题
- ✅ 功能完整可用

## 功能特性

### 核心功能
- **智能问答**: 基于OpenAI GPT模型的问题解答
- **问题分析**: Bug分析、图片分析、Issue分类
- **社区统计**: GitHub仓库活跃度统计
- **知识融合**: 多源知识检索和融合
- **记忆组件**: 工作记忆和短期记忆管理
- **MCP集成**: 统一的MCP服务器集成支持

### MCP集成功能
本项目支持集成远程MCP服务器，通过统一的MCP管理器调用外部工具和服务：

#### 支持的MCP服务器
- **DeepWiki**: 查询GitHub仓库信息和文档
- **其他**: 支持自定义MCP服务器

#### 统一MCP架构
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

#### API接口
- `POST /api/v1/mcp/query` - 执行MCP查询
- `POST /api/v1/mcp/tools` - 获取工具列表
- `POST /api/v1/mcp/call` - 调用特定工具

#### 使用示例
```bash
# 查询GitHub仓库信息
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "input": "What transport protocols are supported in the MCP spec?",
    "repo_name": "modelcontextprotocol/modelcontextprotocol"
  }'

# 获取工具列表
curl -X POST http://localhost:8080/api/v1/mcp/tools \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "server_url": "https://mcp.deepwiki.com/mcp"
  }'

# 调用特定工具
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

#### 配置MCP服务器
在 `configs/config.yaml` 中配置MCP服务器：

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
      allowed_tools: ["ask_question", "read_wiki_structure"]
```

#### 特性
- **统一管理**: 通过MCP管理器统一管理所有服务器
- **备用方案**: MCP失败时自动切换到HTTP调用
- **网络代理**: 支持代理配置解决网络访问问题
- **健康检查**: 内置服务器健康状态监控
- **错误处理**: 完整的错误处理和重试机制

### 技术特性
- **MCP协议**: 基于Model Context Protocol的标准化接口
- **模块化架构**: 清晰的模块分离和依赖管理
- **容器化部署**: 支持Docker容器化部署
- **Google Gmail API**: 发送和接收邮件，管理邮件会话
- **Google Groups API**: 管理邮件组成员，获取组信息
- **实时监听**: 监听邮件变化，及时处理新邮件
- **状态跟踪**: 完整的Issue处理状态跟踪和统计
- **数据备份**: 支持数据导出、导入和备份功能
- **网络代理**: 支持代理配置，解决网络访问问题
- **统一MCP管理**: 集中管理所有MCP服务器连接和配置

### Google API集成功能
本项目集成了Google API，实现了Agent作为Google账号加入私有邮件组，分析GitHub Issue，无法解决的问题通过邮件组与维护者交流，维护者回复后Agent理解并回复Issue的功能。

#### 完整Issue处理流程
```
用户提出Issue → Agent分析 → 无法解决 → 发送邮件给维护者 → 维护者回复 → Agent理解回复 → 生成Issue回复 → 更新Issue状态
```

#### 核心功能
- **GitHub Issue分析**: 自动分析GitHub Issue内容，判断是否可以自动解决
- **邮件组集成**: Agent作为成员加入私有邮件组，与维护者进行交流
- **多问题并行处理**: 支持同时处理多个Issue，每个Issue创建独立的邮件会话
- **邮件会话管理**: 维护Issue与邮件会话的映射关系，跟踪处理状态
- **自动回复生成**: 基于维护者回复自动生成Issue回复内容
- **LLM/DeepWiki集成**: 结合LLM和DeepWiki进行深度分析和知识检索

#### API接口
- `POST /api/google/issues` - 处理GitHub Issue
- `GET /api/google/issues` - 获取Issue列表
- `POST /api/google/emails/send` - 发送邮件给维护者
- `POST /api/google/emails/reply` - 处理维护者邮件回复
- `GET /api/google/threads` - 获取邮件会话列表
- `GET /api/google/stats` - 获取处理统计信息

#### 完整流程示例

##### 1. 用户提出Issue
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

##### 2. Agent分析Issue
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
  "analysis": {
    "can_resolve": false,
    "priority": "high",
    "tags": ["bug", "ai-gateway", "timeout"],
    "summary": "需要维护者协助处理AI Gateway连接超时问题",
    "requires_maintainer": true
  }
}
```

##### 3. Agent发送邮件给维护者
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

##### 4. 维护者回复邮件
```bash
# 维护者通过邮件回复
# 邮件内容: "这个问题是由于AI Gateway的默认超时设置过短导致的。建议将timeout配置从30秒增加到120秒，并在配置文件中添加retry机制。"
```

##### 5. Agent监听并处理邮件回复
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

##### 6. Agent结合LLM和DeepWiki进行深度分析
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

##### 7. Agent生成完整的Issue回复
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

##### 8. 更新Issue状态
```bash
# 更新Issue状态为已解决
curl -X PUT http://localhost:8080/api/google/issues/123/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "resolved"
  }'
```

#### 配置Google API
在 `configs/google_config.yaml` 中配置Google API：

```yaml
google:
  gmail:
    credentials_file: "credentials.json"  # 服务账号凭证文件
    token_file: "token.json"             # 访问令牌文件
    group_email: "maintainers@higress.io" # 维护者邮件组
    scopes:                              # 权限范围
      - "https://www.googleapis.com/auth/gmail.send"
      - "https://www.googleapis.com/auth/gmail.readonly"
      - "https://www.googleapis.com/auth/gmail.modify"
  
  groups:
    admin_email: "admin@higress.io"      # 管理员邮箱
    group_key: "maintainers@higress.io"  # 邮件组标识
    domain: "higress.io"                 # 域名

email:
  maintainers:                           # 维护者邮箱列表
    - "maintainer1@higress.io"
    - "maintainer2@higress.io"
    - "maintainer3@higress.io"
```

#### 特性
- **智能分析**: 结合LLM和DeepWiki进行深度分析
- **自动处理**: 自动识别可解决的问题并处理
- **邮件集成**: 无缝集成邮件组通信
- **状态跟踪**: 完整的Issue处理状态跟踪
- **多源知识**: 结合本地知识库、DeepWiki和LLM
- **实时监听**: 实时监听邮件变化并处理

## 项目架构

### 整体架构
```
Community Governance MCP Higress
├── 服务层 (Service Layer)
│   ├── Agent服务
│   ├── 知识检索服务
│   ├── 社区治理服务
│   ├── Google API服务
│   └── MCP集成服务
├── 处理器层 (Processor Layer)
│   ├── 请求处理器
│   ├── 响应处理器
│   ├── 错误处理器
│   ├── 日志处理器
│   ├── MCP管理器
│   └── Google API管理器
├── 工具层 (Tools Layer)
│   ├── 知识库工具
│   ├── 社区统计工具
│   ├── GitHub管理工具
│   ├── Google API工具
│   └── MCP客户端
├── 数据层 (Data Layer)
│   ├── 内存存储
│   ├── 文件存储
│   ├── 外部API
│   ├── MCP服务器
│   └── Google API服务
└── 配置层 (Config Layer)
    ├── 应用配置
    ├── API配置
    ├── MCP配置
    ├── Google API配置
    └── 环境配置
```

### 数据流
1. **请求接收** → 2. **请求解析** → 3. **工具选择** → 4. **执行处理** → 5. **响应生成** → 6. **返回结果**

### 知识融合流程
1. **知识检索** → 2. **内容分析** → 3. **知识整合** → 4. **结果优化** → 5. **输出展示**

### MCP集成流程
1. **MCP查询** → 2. **服务器连接** → 3. **工具调用** → 4. **响应解析** → 5. **结果返回**

### Google API集成流程
1. **Issue接收** → 2. **内容分析** → 3. **邮件发送** → 4. **回复监听** → 5. **回复处理** → 6. **Issue回复**

### Issue处理完整流程
```
用户提出Issue → Agent分析 → 无法解决 → 发送邮件给维护者 → 维护者回复 → Agent理解回复 → LLM/DeepWiki分析 → 生成Issue回复 → 更新Issue状态
```

### 工具集成架构
```
工具管理器
├── 知识库工具
│   ├── 文档检索
│   ├── 内容分析
│   └── 知识更新
├── 社区治理工具
│   ├── 问题分类
│   ├── 统计分析
│   └── 报告生成
├── GitHub管理工具
│   ├── Issue管理
│   ├── PR处理
│   └── 仓库监控
├── Google API工具
│   ├── Gmail客户端
│   ├── Groups客户端
│   └── 邮件管理器
└── MCP集成工具
    ├── MCP管理器
    ├── 客户端管理
    └── 配置管理
```

### 配置管理架构
```
配置系统
├── 应用配置
│   ├── Agent配置
│   ├── 服务器配置
│   └── 日志配置
├── API配置
│   ├── OpenAI配置
│   ├── GitHub配置
│   ├── Google API配置
│   └── MCP配置
├── 服务配置
│   ├── 邮件组配置
│   ├── 维护者配置
│   └── 模板配置
└── 环境配置
    ├── 网络代理
    ├── 超时设置
    └── 重试策略
```

## 快速开始

### 环境要求
- Go 1.21+
- Docker (可选)
- OpenAI API Key
- GitHub Token (可选)
- Google API Credentials (可选)

### 安装和运行

1. **克隆项目**
```bash
git clone https://github.com/your-org/community-governance-mcp-higress.git
cd community-governance-mcp-higress
```

2. **配置环境变量**
```bash
cp env.example .env
# 编辑.env文件，设置必要的API密钥
```

3. **安装依赖**
```bash
go mod tidy
```

4. **编译项目**
```bash
go build -o bin/agent cmd/agent/main.go
```

5. **运行服务**
```bash
./bin/agent
```

### Docker部署

1. **构建镜像**
```bash
docker build -t community-governance-mcp-higress .
```

2. **运行容器**
```bash
docker run -p 8080:8080 --env-file .env community-governance-mcp-higress
```

### 使用示例

#### 智能问答
```bash
curl -X POST http://localhost:8080/api/v1/process \
  -H "Content-Type: application/json" \
  -d '{
    "type": "question",
    "content": "如何配置Higress的AI Gateway？",
    "author": "user123"
  }'
```

#### MCP查询
```bash
curl -X POST http://localhost:8080/api/v1/mcp/query \
  -H "Content-Type: application/json" \
  -d '{
    "server_label": "deepwiki",
    "input": "What is the MCP protocol?",
    "repo_name": "modelcontextprotocol/modelcontextprotocol"
  }'
```

#### 问题分析
```bash
curl -X POST http://localhost:8080/api/v1/analyze \
  -H "Content-Type: application/json" \
  -d '{
    "type": "bug",
    "content": "Error: connection timeout",
    "stack_trace": "..."
  }'
```

#### Google API集成
```bash
# 处理GitHub Issue
curl -X POST http://localhost:8080/api/google/issues \
  -H "Content-Type: application/json" \
  -d '{
    "issue_id": "123",
    "issue_url": "https://github.com/alibaba/higress/issues/123",
    "issue_title": "Bug: Connection timeout in AI Gateway",
    "issue_content": "When using AI Gateway with large models, connections frequently timeout after 30 seconds."
  }'

# 获取待处理Issue列表
curl -X GET http://localhost:8080/api/google/issues?status=pending

# 同步邮件
curl -X POST http://localhost:8080/api/google/emails/sync

# 获取统计信息
curl -X GET http://localhost:8080/api/google/stats
```

## 配置说明

### 基础配置
```yaml
# configs/config.yaml
agent:
  name: "higress-community-agent"
  version: "1.0.0"
  port: 8080
  debug: true

openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o"
  max_tokens: 4000
  temperature: 0.7
```

### MCP配置
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
      allowed_tools: ["ask_question", "read_wiki_structure"]
```

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
  
  groups:
    admin_email: "admin@higress.io"      # 管理员邮箱
    group_key: "maintainers@higress.io"  # 邮件组标识
    domain: "higress.io"                 # 域名

email:
  maintainers:                           # 维护者邮箱列表
    - "maintainer1@higress.io"
    - "maintainer2@higress.io"
    - "maintainer3@higress.io"
```

### 网络代理配置
```yaml
network:
  proxy_enabled: true
  proxy_url: "http://proxy.example.com:8080"
  timeout: "30s"
  max_retries: 3
```

## 开发指南

### 项目结构
```
community-governance-mcp-higress/
├── cmd/agent/           # 主程序入口
├── config/              # 配置管理
├── configs/             # 配置文件
├── docs/                # 文档
├── examples/            # 示例配置
├── internal/            # 内部包
│   ├── agent/          # 处理器
│   ├── model/          # 数据模型
│   ├── mcp/            # MCP集成
│   ├── openai/         # OpenAI客户端
│   └── google/         # Google API
├── tools/               # 工具包
├── utils/               # 工具函数
├── test/                # 测试文件
└── main.go             # 主程序
```

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

### 扩展Google API功能

1. **添加新的邮件模板**
```yaml
email:
  templates:
    custom_notification:
      subject: "[Custom] {issue_title}"
      content: |
        自定义邮件模板内容
        - Issue ID: {issue_id}
        - 标题: {issue_title}
        - 内容: {issue_content}
```

2. **添加新的维护者**
```yaml
email:
  maintainers:
    - "new_maintainer@higress.io"
```

3. **自定义Issue分析规则**
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

### 扩展功能

1. **添加新工具**
   - 在`tools/`目录下创建新工具
   - 实现工具接口
   - 注册到工具管理器

2. **添加新的知识源**
   - 在`processor.go`中添加检索方法
   - 更新知识融合逻辑
   - 添加配置支持

3. **添加新的API端点**
   - 在`cmd/agent/main.go`中注册路由
   - 实现处理器方法
   - 添加文档和测试

4. **添加新的Google API功能**
   - 在`internal/google/`目录下添加新客户端
   - 在`manager.go`中添加管理逻辑
   - 在`handler.go`中添加HTTP处理器
   - 更新配置和文档

## 测试

### 运行测试
```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test ./test/

# 运行集成测试
go test ./test/integration_test.go
```

### 测试覆盖率
```bash
go test -cover ./...
```

## 监控和日志

### 健康检查
```bash
curl http://localhost:8080/api/v1/health
```

### 日志配置
```yaml
logging:
  level: "info"
  format: "json"
  output: "stdout"
  file_path: "./logs/agent.log"
```

## 贡献指南

1. Fork项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建Pull Request

## 许可证

本项目采用MIT许可证。详见LICENSE文件。

## 联系方式

- 项目主页: https://github.com/your-org/community-governance-mcp-higress
- 问题反馈: https://github.com/your-org/community-governance-mcp-higress/issues
- 文档: https://github.com/your-org/community-governance-mcp-higress/docs  
