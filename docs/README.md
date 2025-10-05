# OpenList-STRM 文档中心

欢迎来到 OpenList-STRM 项目文档中心。

## 📖 文档导航

### 核心文档

#### [项目需求文档 (PRD)](./PRD.md)
- 项目概述和核心价值
- 完整功能需求
- 技术架构设计
- 开发计划和里程碑
- 推荐工具和最佳实践

**适合人群**: 项目贡献者、架构师、产品经理

---

#### [测试计划文档 (TESTING)](./TESTING.md)
- 测试策略和覆盖目标
- 详细测试清单
- 测试进度跟踪
- 测试编写规范
- 测试命令参考

**适合人群**: 开发者、QA 工程师、贡献者

---

### 部署文档

#### [Docker 部署指南](../deployments/README.md)
- 快速开始
- 配置说明
- 常用命令
- 与 Alist 集成
- 故障排查

**适合人群**: 运维工程师、系统管理员、终端用户

---

#### [Webhook 集成指南](../deployments/WEBHOOK.md)
- Webhook 接口说明
- 与下载器集成 (qBittorrent、Transmission)
- 与自动化工具集成 (n8n、Home Assistant)
- 使用场景和最佳实践

**适合人群**: 高级用户、自动化爱好者

---

### 配置文件

#### [配置文件示例](../configs/config.example.yaml)
- 完整的配置文件模板
- 每个配置项的详细说明
- 常用配置示例

**适合人群**: 所有用户

---

## 🚀 快速开始

### 新用户
1. 阅读 [README](../README.md) 了解项目概述
2. 参考 [配置文件示例](../configs/config.example.yaml) 准备配置
3. 按照 [Docker 部署指南](../deployments/README.md) 快速部署

### 开发者
1. 阅读 [PRD](./PRD.md) 了解项目架构
2. 阅读 [TESTING](./TESTING.md) 了解测试规范
3. 克隆代码开始开发

### 高级用户
1. 阅读 [Webhook 集成指南](../deployments/WEBHOOK.md)
2. 根据需求配置自动化流程

---

## 📊 文档统计

| 文档类型 | 数量 | 总字数 |
|---------|------|--------|
| 核心文档 | 2 | ~8000 字 |
| 部署文档 | 2 | ~3000 字 |
| 配置示例 | 1 | ~80 行 |
| **总计** | **5** | **~11000 字** |

---

## 🔄 文档更新记录

| 日期 | 文档 | 更新内容 |
|------|------|---------|
| 2025-10-05 | TESTING.md | 新增测试计划文档 |
| 2025-10-05 | deployments/WEBHOOK.md | 新增 Webhook 集成指南 |
| 2025-10-05 | deployments/README.md | 新增 Docker 部署指南 |
| 2025-10-04 | PRD.md | 更新 Phase 2、3 进度 |
| 2025-10-04 | README.md | 更新功能列表和使用说明 |

---

## 📝 贡献指南

### 改进文档

欢迎提交文档改进建议！

1. Fork 本项目
2. 修改文档
3. 提交 Pull Request

### 文档规范

- 使用 Markdown 格式
- 保持清晰的章节结构
- 添加必要的代码示例
- 使用表格组织信息
- 添加适当的 emoji 提升可读性

---

## ❓ 常见问题

### 在哪里找配置说明？
参考 [配置文件示例](../configs/config.example.yaml)，每个配置项都有详细注释。

### 如何部署到生产环境？
参考 [Docker 部署指南](../deployments/README.md)，推荐使用 Docker Compose。

### 如何集成到自动化流程？
参考 [Webhook 集成指南](../deployments/WEBHOOK.md)，支持多种自动化工具。

### 如何贡献代码？
参考 [PRD](./PRD.md) 了解架构，参考 [TESTING](./TESTING.md) 编写测试。

---

最后更新: 2025-10-05
