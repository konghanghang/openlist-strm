# 系统架构

本文档描述 OpenList-STRM 当前代码实现的真实结构。它回答 4 件事：

1. 程序如何启动
2. 核心模块分别负责什么
3. 关键数据和执行链路是什么
4. 哪些行为是当前实现已经确定的

如果本文档与代码冲突，以代码为准，并在完成改动后同步更新本文档。

## 1. 总体结构

OpenList-STRM 当前是一个单仓库、单后端进程 + 单前端构建产物的应用：

- `backend/`：Go 后端，负责配置加载、数据库、Alist 访问、STRM 生成、调度、API、媒体库通知
- `web/`：Vue 3 前端，构建后由 Go 服务托管
- `configs/`：配置模板
- `deployments/`：部署相关文档和文件
- `docs/`：项目文档

## 2. 启动流程

当前启动入口是 [backend/cmd/server/main.go](/Users/konghang/data/github/openlist-strm/backend/cmd/server/main.go:1)。

启动顺序：

1. 解析命令行参数，支持 `-config` 和 `-version`
2. 加载配置文件
3. 初始化日志
4. 打开 SQLite 数据库并执行 GORM `AutoMigrate`
5. 创建 Alist 客户端并执行 `Ping`
6. 创建 STRM 生成器
7. 创建并启动调度器，调度器会从数据库读取映射配置和 Cron 表达式
8. 创建 Gin API 服务
9. 如果启用了 Web UI，则注册前端静态资源路由
10. 启动 HTTP 服务并等待退出信号

## 3. 核心模块

### 3.1 配置模块

文件：

- [backend/internal/config/config.go](/Users/konghang/data/github/openlist-strm/backend/internal/config/config.go:1)
- [backend/internal/config/loader.go](/Users/konghang/data/github/openlist-strm/backend/internal/config/loader.go:1)

职责：

- 读取 YAML 配置
- 校验基础配置
- 提供服务监听地址

当前稳定事实：

- 路径映射不再从 YAML 读取，而是由 Web UI 写入 SQLite
- `web.username` 和 `web.password` 目前仍存在于配置中，但当前版本未形成真正的前端登录边界
- `media_server` 配置用于任务完成后的 Emby/Jellyfin 通知

### 3.2 存储模块

文件：

- [backend/internal/storage/models.go](/Users/konghang/data/github/openlist-strm/backend/internal/storage/models.go:1)
- [backend/internal/storage/sqlite.go](/Users/konghang/data/github/openlist-strm/backend/internal/storage/sqlite.go:1)

职责：

- 管理 SQLite 连接
- 自动迁移表结构
- 提供 `files`、`tasks`、`mappings`、`users` 四类数据访问

当前稳定事实：

- 映射配置以 `mappings` 表为准
- 任务历史以 `tasks` 表为准
- 数据库 schema 由启动时 `AutoMigrate` 维护，没有独立 SQL migration 体系

### 3.3 Alist 客户端

文件：

- [backend/internal/alist/client.go](/Users/konghang/data/github/openlist-strm/backend/internal/alist/client.go:1)
- [backend/internal/alist/types.go](/Users/konghang/data/github/openlist-strm/backend/internal/alist/types.go:1)

职责：

- 访问 Alist API
- 递归列出文件
- 获取可播放 URL
- 支持签名模式

### 3.4 STRM 生成器

文件：

- [backend/internal/strm/generator.go](/Users/konghang/data/github/openlist-strm/backend/internal/strm/generator.go:1)

职责：

- 根据源路径扫描结果生成 `.strm` 文件
- 支持 `incremental` / `full` 两种模式
- 支持 `alist_path` / `http_url` 两种 STRM 内容模式
- 以配置粒度控制并发

当前稳定事实：

- `full` 模式会先清空目标目录
- `incremental` 模式下，目标 `.strm` 已存在则直接跳过
- 同名不同扩展名文件会按扩展优先级去重

### 3.5 调度器

文件：

- [backend/internal/scheduler/scheduler.go](/Users/konghang/data/github/openlist-strm/backend/internal/scheduler/scheduler.go:1)

职责：

- 从数据库读取启用的映射配置
- 为带 `cron_expr` 的映射注册 Cron 任务
- 记录任务历史
- 在生成结束后触发媒体服务器通知

当前稳定事实：

- Cron 表达式以映射记录中的 `cron_expr` 为准
- 手动触发和 API 触发最终都会走 `RunMapping` / `RunAll`
- 任务执行使用 `task_id` 作为 trace id 写日志

### 3.6 API 服务

文件：

- [backend/internal/api/router.go](/Users/konghang/data/github/openlist-strm/backend/internal/api/router.go:1)
- [backend/internal/api/handlers.go](/Users/konghang/data/github/openlist-strm/backend/internal/api/handlers.go:1)
- [backend/internal/api/middleware.go](/Users/konghang/data/github/openlist-strm/backend/internal/api/middleware.go:1)

职责：

- 提供健康检查、任务查询、映射配置管理、Webhook 触发、状态查询
- 在配置了 `api.token` 时启用 Token 鉴权

当前稳定事实：

- 健康检查路由是 `GET /health`
- 业务 API 路由挂在 `/api` 下
- `POST /api/generate` 支持触发全部映射或指定映射
- `POST /api/webhook` 用于外部事件触发
- `POST /api/cron/preview` 接收 6 字段 cron 表达式，返回未来 N 次（1~10）触发时间，时区与调度器一致；与 createMapping 校验严格对齐，不接受 5 字段

### 3.7 Web UI 托管

文件：

- [backend/internal/web/handler.go](/Users/konghang/data/github/openlist-strm/backend/internal/web/handler.go:1)
- `web/src/**`

职责：

- 托管前端构建产物
- 在开发场景优先使用外部 `web/dist`
- 在生产场景使用 embed 打包的静态资源

当前前端页面：

- `Dashboard`
- `Tasks`
- `Configs`

## 4. 关键数据模型

### 4.1 Mapping

当前映射记录包含：

- 名称
- Alist 源路径
- STRM 目标路径
- 扩展名列表
- 并发数
- 更新模式
- STRM 模式
- 是否强制刷新 Alist 缓存
- 是否启用
- Cron 表达式

它是调度、手动执行、Webhook 命中的核心配置实体。

### 4.2 Task

任务记录包含：

- `task_id`
- `config_name`
- `mode`
- `status`
- `files_created`
- `files_deleted`
- `files_skipped`
- `errors`
- `started_at`
- `completed_at`

它是 Web UI 和 API 查询执行结果的主要数据来源。

## 5. 关键链路

### 5.1 手动/API 触发链路

1. 前端或调用方请求 API
2. API 创建后台任务上下文和 `task_id`
3. 调度器定位目标映射
4. 生成器扫描 Alist 并写入 `.strm`
5. 调度器更新任务记录
6. 如有文件新增或删除，则通知 Emby/Jellyfin 扫描库

### 5.2 Cron 自动执行链路

1. 服务启动时读取数据库映射
2. 为启用且带 `cron_expr` 的映射注册任务
3. 触发时走与手动执行相同的 `RunMapping` 链路

### 5.3 Webhook 链路

1. 外部系统调用 `POST /api/webhook`
2. 服务按请求中的配置名或路径定位映射
3. 调度器执行对应映射
4. 任务结果进入 `tasks` 表，必要时触发媒体库刷新

## 6. 当前实现边界

- 当前项目的“配置中心”是 YAML + SQLite 混合模式：
  - 全局参数来自 YAML
  - 映射配置来自 SQLite
- 当前数据库使用 SQLite，且依赖 GORM 自动迁移
- 当前 API 认证是可选 Token 鉴权，不是完整用户权限体系
- 当前 Web UI 是单页应用，由 Go 进程直接托管
- 当前媒体服务器通知支持 Emby / Jellyfin 两类目标

## 7. 关联文档

- [开发指南](./development-guide.md)
- [路线图](./roadmap.md)
- [测试计划](./TESTING.md)
- [部署文档](../deployments/README.md)
