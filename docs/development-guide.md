# 开发指南

本文件只回答 3 件事：

1. 开发 OpenList-STRM 时先看什么
2. 当前哪些文档算现行依据
3. 改动完成后最小要验证什么

## 最短阅读路径

1. [README.md](../README.md)
2. [系统架构](./system-architecture.md)
3. [PRD.md](./PRD.md)
4. [路线图](./roadmap.md)
5. [TESTING.md](./TESTING.md)
6. [配置示例](../configs/config.example.yaml)
7. 按任务进入部署文档或代码目录：
   - 部署与运行：[`deployments/README.md`](../deployments/README.md)
   - Webhook 集成：[`deployments/WEBHOOK.md`](../deployments/WEBHOOK.md)
   - 后端实现：`backend/`
   - 前端实现：`web/`，前端动手前先读 [前端设计规范](./design-system/index.md)

## 文档判断规则

- 项目入口、功能概览、使用方式：看 [README.md](../README.md)。
- 当前代码实现、模块边界、关键执行链路：看 [system-architecture.md](./system-architecture.md)。
- 产品需求、功能范围、较稳定的架构说明：看 [PRD.md](./PRD.md)。
- 当前阶段目标和后续方向：看 [roadmap.md](./roadmap.md)。
- 测试策略、测试清单、当前覆盖进度：看 [TESTING.md](./TESTING.md)。
- 前端样式、色彩、组件、微交互、反模式：看 [前端设计规范](./design-system/index.md)。
- 运行配置和可选项：看 [配置示例](../configs/config.example.yaml)。
- 部署方式、容器使用、故障排查：看 [`deployments/README.md`](../deployments/README.md)。
- Webhook 用法和外部集成：看 [`deployments/WEBHOOK.md`](../deployments/WEBHOOK.md)。

如果文档与代码冲突，以代码当前实现为准，并在完成修改后同步回文档。

## 目录边界

- `backend/`：Go 后端、API、调度、文件生成、存储逻辑
- `web/`：Vue 前端页面和 API 调用
- `configs/`：配置模板
- `deployments/`：部署说明和部署相关文件
- `docs/`：文档中心

## 开发约束

- 改动前先确认影响范围，不要在未理解链路时直接重写实现。
- 涉及配置、调度、Webhook、STRM 生成链路时，要检查用户可见行为是否改变。
- 新增稳定规则时，优先更新现有文档，不新增“临时说明”文件。
- `AGENTS.md` 与 `CLAUDE.md` 是协作规范，内容必须保持一致。

## 最小验证

### 后端改动

在 `backend/` 目录下按需执行：

```bash
gofmt -w .
go test ./...
go build ./...
```

如果仓库已安装 `golangci-lint`，再补：

```bash
golangci-lint run --timeout=5m
```

### 前端改动

在 `web/` 目录下至少执行：

```bash
npm run build
```

### 文档改动

- 检查被引用的文件是否真实存在
- 检查相对链接是否可达
- 确认没有把同一规则写进多份文档

## 文档维护

- 修改 `README.md`、`PRD.md`、`TESTING.md` 任一文件时，顺手检查另外两者是否需要同步。
- 修改 `system-architecture.md` 或 `roadmap.md` 时，顺手检查 `PRD.md` 和 `README.md` 是否也需要同步。
- 失效链接和失效路径要立即修正，不要继续保留占位引用。
- 如果某段内容已经不再准确，优先删掉或改写，不要保留“以后再整理”的残留。
