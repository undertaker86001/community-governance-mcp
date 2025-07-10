# Higress社区治理智能助手

## 项目概述

这是一个融合了DeepWiki知识检索能力和Higress社区治理工具的智能助手系统。系统基于MCP（Model Context Protocol）协议，提供智能问答、问题分析、社区统计等功能，专门为Higress社区治理设计。

## 核心功能

### 1. 智能问答系统
- **多源知识检索**: 从DeepWiki、Higress文档、本地知识库等多个来源检索信息
- **知识融合**: 智能融合多源信息，提供准确、全面的回答
- **上下文理解**: 理解用户问题的上下文和意图
- **问题分类**: 自动识别问题类型（Issue、PR、文本问题等）

### 2. 社区治理工具
- **问题分类器**: 自动分类GitHub Issues和Pull Requests
- **Bug分析器**: 分析错误堆栈，提供诊断建议
- **图片分析器**: 分析截图和错误图片
- **GitHub管理器**: 管理GitHub仓库和Issue
- **社区统计**: 生成社区活跃度统计报告
- **知识库管理**: 本地化存储和管理常用问题和解决方案

### 3. 知识库管理
- **知识存储**: 本地化存储常用问题和解决方案
- **知识更新**: 定期更新知识库内容
- **知识检索**: 快速检索相关知识和解决方案
- **知识融合**: 智能融合多源知识

### 4. 多轮会话记忆管理
- **工作记忆**: 存储最近20个重要问题和上下文，生存时间30分钟
- **短期记忆**: 16个记忆槽存储用户交互历史，生存时间2小时
- **记忆检索**: 基于关键词和标签智能检索相关记忆
- **记忆融合**: 将历史记忆融入当前问题的处理中
- **自动清理**: 定期清理过期记忆，优化内存使用

## 系统架构

```
community-governance-mcp-higress/
├── cmd/                    # 命令行工具
│   └── agent/             # 主服务程序
├── internal/              # 内部核心模块
│   ├── agent/            # 智能代理核心
│   │   ├── processor.go  # 问题处理器
│   │   └── types.go      # 核心类型定义
│   ├── openai/           # OpenAI客户端
│   └── memory/           # 记忆组件
│       ├── types.go      # 记忆类型定义
│       ├── manager.go    # 记忆管理器
│       └── handler.go    # 记忆HTTP处理器
├── tools/                # 社区治理工具
│   ├── bug_analyzer.go   # Bug分析器
│   ├── community_stats.go # 社区统计
│   ├── github_manager.go # GitHub管理器
│   ├── image_analyzer.go # 图片分析器
│   ├── issue_classifier.go # 问题分类器
│   ├── knowledge_base.go # 知识库管理
│   └── load_tools.go     # 工具加载器
├── configs/              # 配置文件
│   └── config.yaml       # 主配置文件
├── docs/                 # 文档
├── test/                 # 测试文件
└── utils/                # 工具函数
```

## 快速开始

### 环境要求
- Go 1.21+
- OpenAI API Key
- GitHub Token (可选)

### 安装和运行

1. **克隆项目**
```bash
git clone <repository-url>
cd community-governance-mcp-higress
```

2. **配置环境变量**
```bash
cp .env.example .env
# 编辑.env文件，填入必要的API密钥
```

3. **安装依赖**
```bash
go mod tidy
```

4. **运行服务**
```bash
# 方式1: 直接运行
go run cmd/agent/main.go

# 方式2: 使用Makefile
make run

# 方式3: 使用Docker
docker-compose up
```

服务将在 `http://localhost:8080` 启动

## API接口

### 1. 智能问答接口

**POST** `/api/v1/process`

处理用户问题，返回智能回答。

**请求示例:**
```json
{
  "title": "Higress网关配置问题",
  "content": "如何配置Higress网关的路由规则？",
  "author": "user123",
  "type": "question",
  "priority": "normal",
  "tags": ["gateway", "configuration"]
}
```

**响应示例:**
```json
{
  "id": "uuid",
  "question_id": "uuid",
  "content": "详细的回答内容...",
  "summary": "回答摘要",
  "sources": [
    {
      "title": "Higress官方文档",
      "url": "https://higress.io/docs",
      "relevance": 0.95
    }
  ],
  "confidence": 0.92,
  "processing_time": "1.2s",
  "fusion_score": 0.88,
  "recommendations": ["建议1", "建议2"]
}
```

### 2. 问题分析接口

**POST** `/api/v1/analyze`

分析错误堆栈和问题信息。

**请求示例:**
```json
{
  "stack_trace": "错误堆栈信息...",
  "environment": "Kubernetes 1.24",
  "version": "Higress 1.0.0"
}
```

### 3. 社区统计接口

**GET** `/api/v1/stats`

获取社区活跃度统计。

### 4. 健康检查

**GET** `/api/v1/health`

检查服务状态。

### 5. 记忆管理接口

#### 存储记忆
**POST** `/api/v1/memory/store`

存储用户交互记忆。

**请求示例:**
```json
{
  "session_id": "session_123",
  "user_id": "user_456",
  "type": "working",
  "content": "用户询问了关于Higress网关配置的问题",
  "context": "网关配置相关",
  "tags": ["gateway", "config"],
  "metadata": {"priority": "high"}
}
```

#### 检索记忆
**POST** `/api/v1/memory/retrieve`

检索相关记忆。

**请求示例:**
```json
{
  "session_id": "session_123",
  "user_id": "user_456",
  "type": "working",
  "keywords": ["gateway", "config"],
  "limit": 5
}
```

#### 获取记忆统计
**GET** `/api/v1/memory/stats/:session_id`

获取记忆使用统计。

#### 清除记忆
**DELETE** `/api/v1/memory/clear/:session_id`

清除指定会话的记忆。

## 配置说明

### 主要配置项

```yaml
# Agent基础配置
agent:
  name: "higress-community-agent"
  version: "1.0.0"
  port: 8080
  debug: true

# OpenAI配置
openai:
  api_key: "${OPENAI_API_KEY}"
  model: "gpt-4o"
  max_tokens: 4000
  temperature: 0.7

# DeepWiki配置
deepwiki:
  enabled: true
  endpoint: "https://mcp.deepwiki.com/mcp"
  timeout: "30s"

# 知识库配置
knowledge:
  enabled: true
  storage_path: "./data/knowledge"
  max_size: "1GB"

# 记忆组件配置
memory:
  working_memory_max_items: 20
  working_memory_ttl: "30m"
  short_term_memory_slots: 16
  short_term_memory_ttl: "2h"
  cleanup_interval: "5m"
  importance_threshold: 0.3
```

## 工具说明

### 1. Bug分析器 (bug_analyzer.go)
- 分析错误堆栈信息
- 提供诊断建议
- 生成修复方案

### 2. 问题分类器 (issue_classifier.go)
- 自动分类GitHub Issues
- 识别问题类型和优先级
- 分配处理人员

### 3. 图片分析器 (image_analyzer.go)
- 分析错误截图
- 识别界面问题
- 提供可视化建议

### 4. GitHub管理器 (github_manager.go)
- 管理GitHub仓库
- 自动化Issue处理
- 社区协作支持

### 5. 社区统计 (community_stats.go)
- 生成社区活跃度报告
- 分析贡献者数据
- 监控项目健康度

### 6. 知识库管理 (knowledge_base.go)
- 本地知识存储
- 知识检索和更新
- 知识融合支持

## 开发指南

### 添加新工具

1. 在 `tools/` 目录下创建新的工具文件
2. 实现工具的核心功能
3. 在 `tools/load_tools.go` 中注册工具
4. 更新配置文件和文档

### 测试

```bash
# 运行所有测试
make test

# 运行特定测试
go test ./test/...

# 运行集成测试
go test ./test/integration_test.go
```

### 构建

```bash
# 构建二进制文件
make build

# 构建Docker镜像
make docker-build

# 发布
make release
```

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 联系方式

- 项目主页: [GitHub Repository](https://github.com/your-org/community-governance-mcp-higress)
- 问题反馈: [Issues](https://github.com/your-org/community-governance-mcp-higress/issues)
- 讨论区: [Discussions](https://github.com/your-org/community-governance-mcp-higress/discussions)  
