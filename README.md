# Community Governance MCP Higress

一个融合了DeepWiki知识检索和Higress社区治理工具的智能助手项目，基于MCP协议，提供智能问答、问题分析、社区统计等功能，并集成了Google API实现Agent与维护者的邮件交流功能。

## 功能特性

### 核心功能
- **智能问答**: 基于DeepWiki知识库的智能问答系统
- **问题分析**: 自动分析GitHub Issue，识别问题类型和优先级
- **社区统计**: 提供社区活跃度、贡献者统计等数据分析
- **Google API集成**: Agent作为Google账号加入私有邮件组，与维护者交流
- **多问题并行处理**: 支持同时处理多个Issue，每个Issue创建独立的邮件会话
- **邮件会话管理**: 维护Issue与邮件会话的映射关系，跟踪处理状态
- **自动回复生成**: 基于维护者回复自动生成Issue回复内容

### 技术特性
- **MCP协议**: 基于Model Context Protocol的标准化接口
- **模块化架构**: 清晰的模块分离和依赖管理
- **容器化部署**: 支持Docker容器化部署
- **Google Gmail API**: 发送和接收邮件，管理邮件会话
- **Google Groups API**: 管理邮件组成员，获取组信息
- **实时监听**: 监听邮件变化，及时处理新邮件
- **状态跟踪**: 完整的Issue处理状态跟踪和统计
- **数据备份**: 支持数据导出、导入和备份功能

## 项目架构

### 整体架构
```
Community Governance MCP Higress
├── 服务层 (Service Layer)
│   ├── Agent服务
│   ├── 知识检索服务
│   ├── 社区治理服务
│   └── Google API服务
├── 处理器层 (Processor Layer)
│   ├── 请求处理器
│   ├── 响应处理器
│   ├── 错误处理器
│   └── 日志处理器
├── 工具层 (Tools Layer)
│   ├── 知识库工具
│   ├── 社区统计工具
│   ├── GitHub管理工具
│   └── Google API工具
├── 数据层 (Data Layer)
│   ├── 内存存储
│   ├── 文件存储
│   └── 外部API
└── 配置层 (Config Layer)
    ├── 应用配置
    ├── API配置
    └── 环境配置
```

### 数据流
1. **请求接收** → 2. **请求解析** → 3. **工具选择** → 4. **执行处理** → 5. **响应生成** → 6. **返回结果**

### 知识融合流程
1. **知识检索** → 2. **内容分析** → 3. **知识整合** → 4. **结果优化** → 5. **输出展示**

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
└── Google API工具
    ├── Gmail客户端
    ├── Groups客户端
    └── 邮件管理器
```

### 配置管理架构
```
配置系统
├── 应用配置
│   ├── 服务端口
│   ├── 日志级别
│   └── 超时设置
├── API配置
│   ├── OpenAI配置
│   ├── GitHub配置
│   └── Google API配置
├── 工具配置
│   ├── 知识库配置
│   ├── 社区配置
│   └── 邮件配置
└── 环境配置
    ├── 开发环境
    ├── 测试环境
    └── 生产环境
```

### 部署架构
```
容器化部署
├── 应用容器
│   ├── 主服务
│   ├── API服务
│   └── 监控服务
├── 数据容器
│   ├── 配置存储
│   ├── 日志存储
│   └── 缓存存储
├── 网络配置
│   ├── 服务发现
│   ├── 负载均衡
│   └── 安全策略
└── 监控系统
    ├── 健康检查
    ├── 性能监控
    └── 告警通知
```

## 核心组件

### Agent服务
- **智能问答**: 基于OpenAI的智能问答能力
- **工具调用**: 动态加载和执行各种工具
- **会话管理**: 支持多轮对话和上下文记忆
- **错误处理**: 完善的错误处理和恢复机制

### 知识检索服务
- **DeepWiki集成**: 集成DeepWiki知识库
- **内容检索**: 高效的文档检索和内容匹配
- **知识更新**: 支持知识库的动态更新
- **缓存优化**: 智能缓存机制提升性能

### 社区治理服务
- **问题分析**: 自动分析GitHub Issue和PR
- **统计分析**: 社区活跃度和贡献者统计
- **报告生成**: 自动生成社区治理报告
- **趋势预测**: 基于历史数据的趋势分析

### Google API服务
- **Gmail集成**: 发送和接收邮件
- **Groups管理**: 管理邮件组成员
- **会话跟踪**: 维护Issue与邮件的映射关系
- **状态管理**: 完整的处理状态跟踪

## 快速开始

### 环境要求
- Go 1.21+
- Docker (可选)
- Google Cloud项目 (用于Google API)

### 安装依赖
```bash
go mod download
```

### 配置设置
1. 复制配置文件模板
```bash
cp configs/config.yaml.example configs/config.yaml
```

2. 配置Google API凭证
```bash
# 下载Google服务账号凭证文件
# 重命名为 credentials.json 并放置在项目根目录
```

3. 更新配置文件
```yaml
# configs/config.yaml
google:
  gmail:
    credentials_file: "credentials.json"
    group_email: "your-maintainers@example.com"
  groups:
    admin_email: "your-admin@example.com"
    group_key: "your-group@example.com"
```

### 启动服务
```bash
# 开发模式
go run main.go

# 生产模式
make build
./bin/community-governance-mcp-higress

# Docker模式
docker-compose up -d
```

### 测试功能
```bash
# 测试Google API集成
curl -X POST http://localhost:8080/api/google/issues \
  -H "Content-Type: application/json" \
  -d '{
    "issue_id": "123",
    "issue_url": "https://github.com/test/repo/issues/123",
    "issue_title": "Bug Report",
    "issue_content": "This is a bug description"
  }'

# 获取统计信息
curl http://localhost:8080/api/google/stats
```

## API接口

### 核心API
- `POST /api/chat` - 智能问答
- `GET /api/tools` - 获取可用工具列表
- `POST /api/tools/{tool}` - 执行特定工具

### Google API接口
- `POST /api/google/issues` - 处理GitHub Issue
- `GET /api/google/issues` - 获取Issue列表
- `POST /api/google/emails/send` - 发送邮件
- `POST /api/google/emails/sync` - 同步邮件
- `GET /api/google/stats` - 获取统计信息

### 监控接口
- `GET /health` - 健康检查
- `GET /metrics` - 性能指标

## 配置说明

### 主要配置项
```yaml
# 服务配置
server:
  port: 8080
  host: "0.0.0.0"

# OpenAI配置
openai:
  api_key: "your-openai-api-key"
  model: "gpt-4"
  max_tokens: 4096

# Google API配置
google:
  gmail:
    credentials_file: "credentials.json"
    group_email: "maintainers@example.com"
  groups:
    admin_email: "admin@example.com"
    group_key: "maintainers@example.com"

# 工具配置
tools:
  knowledge_base:
    enabled: true
    cache_size: 1000
  community_stats:
    enabled: true
    update_interval: 3600
```

## 开发指南

### 项目结构
```
community-governance-mcp-higress/
├── cmd/                    # 命令行工具
├── config/                 # 配置管理
├── configs/                # 配置文件
├── internal/               # 内部包
│   ├── agent/             # Agent核心
│   ├── google/            # Google API集成
│   ├── knowledge/         # 知识库
│   └── openai/            # OpenAI客户端
├── tools/                  # 工具集合
├── test/                   # 测试文件
├── docs/                   # 文档
└── examples/               # 示例配置
```

### 添加新工具
1. 在 `tools/` 目录下创建新工具文件
2. 实现工具接口
3. 在 `tools/load_tools.go` 中注册工具
4. 添加测试用例

### 扩展Google API功能
1. 在 `internal/google/` 目录下添加新功能
2. 更新类型定义和处理器
3. 添加配置项
4. 编写测试用例

## 部署指南

### Docker部署
```bash
# 构建镜像
docker build -t community-governance-mcp-higress .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  -v $(pwd)/credentials.json:/app/credentials.json \
  community-governance-mcp-higress
```

### Kubernetes部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: community-governance-mcp-higress
spec:
  replicas: 3
  selector:
    matchLabels:
      app: community-governance-mcp-higress
  template:
    metadata:
      labels:
        app: community-governance-mcp-higress
    spec:
      containers:
      - name: app
        image: community-governance-mcp-higress:latest
        ports:
        - containerPort: 8080
        env:
        - name: CONFIG_FILE
          value: "/app/configs/config.yaml"
        volumeMounts:
        - name: config
          mountPath: /app/configs
        - name: credentials
          mountPath: /app/credentials.json
          subPath: credentials.json
      volumes:
      - name: config
        configMap:
          name: app-config
      - name: credentials
        secret:
          secretName: google-credentials
```

## 监控和维护

### 健康检查
```bash
curl http://localhost:8080/health
```

### 性能监控
```bash
curl http://localhost:8080/metrics
```

### 日志查看
```bash
# 查看应用日志
docker logs community-governance-mcp-higress

# 查看Google API日志
tail -f logs/google_api.log
```

## 故障排除

### 常见问题
1. **Google API认证失败**: 检查凭证文件和服务账号权限
2. **邮件发送失败**: 验证Gmail API配置和权限
3. **工具加载失败**: 检查工具配置和依赖
4. **性能问题**: 优化缓存和并发设置

### 调试方法
1. 启用详细日志
2. 使用API测试工具
3. 检查网络连接
4. 验证配置文件格式

## 贡献指南

### 开发流程
1. Fork项目
2. 创建功能分支
3. 编写代码和测试
4. 提交Pull Request
5. 代码审查和合并

### 代码规范
- 使用Go标准格式化
- 编写完整的注释
- 添加单元测试
- 遵循项目结构约定

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 集成DeepWiki知识检索
- 实现社区治理工具
- 添加Google API集成
- 支持邮件组交流功能
- 实现多问题并行处理
- 添加完整的监控和日志系统

### 计划功能
- 支持更多邮件服务商
- 集成更多Issue平台
- 添加机器学习分析
- 支持多语言处理
- 实现Web管理界面
- 支持集群部署  
