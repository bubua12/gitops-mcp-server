# GitOps MCP Server — 产品设计文档

> 面向开发者和 DevOps 工程师的全功能 GitHub 运维中枢，通过 MCP 协议让 AI Agent 具备完整的 Git 平台操作能力。

## 文档目录

| 文档 | 内容 | 状态 |
|------|------|------|
| [01-产品概述](01-product-overview.md) | 产品定位、愿景、核心价值、竞品对比 | ✅ |
| [02-用户画像与场景](02-user-scenarios.md) | 3 类用户画像、6 大核心使用场景 | ✅ |
| [03-功能模块设计](03-feature-modules.md) | 8 大功能模块（54 个 MCP Tool）详细设计 | ✅ |
| [04-系统架构](04-architecture.md) | 技术选型、分层架构、目录结构、核心流程 | ✅ |
| [05-API/Tool 规范](05-api-spec.md) | 所有 MCP Tool 的完整 Schema 定义 | ✅ |
| [06-安全设计](06-security.md) | 认证、操作安全、传输安全、数据安全 | ✅ |
| [07-里程碑规划](07-roadmap.md) | 5 阶段交付计划、优先级裁剪、演进方向 | ✅ |

## 快速概览

- **协议：** MCP（Model Context Protocol）
- **语言：** Go
- **传输：** stdio / SSE / Streamable HTTP
- **平台：** GitHub（后续扩展 GitLab / Gitea）
- **工具总数：** 54 个 MCP Tool
- **预计工期：** 10 周（MVP 3-4 周）
