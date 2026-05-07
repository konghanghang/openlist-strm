# OpenList-STRM 项目需求文档 (PRD)

## 1. 项目概述

> 当前代码实现和模块边界请看 [系统架构](./system-architecture.md)。
> 当前阶段目标和后续版本方向请看 [路线图](./roadmap.md)。
> 测试策略和覆盖进度请看 [测试计划](./TESTING.md)。

### 1.1 项目简介
OpenList-STRM 是一个基于 Go 语言开发的 STRM 文件生成工具，用于将 Alist 网盘中的媒体文件批量生成为 STRM 格式，供 Emby、Jellyfin、Plex 等流媒体服务器使用。

### 1.2 核心价值
- **免挂载**：无需挂载网盘，通过 STRM 文件直接播放
- **节省空间**：本地只存储小体积的 STRM 文件
- **灵活调度**：每个配置独立定时任务，可视化 Cron 编辑器
- **高性能**：Go 语言实现，每配置独立并发控制
- **MediaWarp 集成**：支持 302 重定向代理，优化播放体验
- **Web UI 管理**：无需编辑配置文件，所有配置通过 Web UI 管理

### 1.3 技术栈
- **后端**：Go 1.23+
- **Web 框架**：Gin
- **数据库**：SQLite (GORM)
- **前端**：Vue.js 3 + Element Plus + Vite
- **部署**：单二进制文件（前端嵌入）/ Docker

---

## 2. 功能需求

### 2.1 核心功能

#### 2.1.1 Alist 集成
- **功能描述**：与 Alist API 集成，获取文件列表和生成直链
- **详细需求**：
  - 支持 Alist v3 API
  - 支持多个 Alist 实例配置
  - 支持 Alist 签名功能（安全增强）
  - 支持自定义请求头和认证
  - 错误处理和重试机制

#### 2.1.2 STRM 文件生成
- **功能描述**：根据 Alist 文件列表生成对应的 STRM 文件
- **详细需求**：
  - 保持原始目录结构
  - 支持常见视频格式：mp4, mkv, avi, mov, flv, wmv, ts, m4v 等
  - 文件名处理：支持中文、特殊字符、空格
  - **双模式支持**：
    - **alist_path 模式**：STRM 内容为 Alist 路径（配合 MediaWarp）
    - **http_url 模式**：STRM 内容为完整 URL（直接播放）
  - 支持自定义文件过滤规则
  - 每配置独立并发控制（默认 3，推荐 1-5，防网盘风控）

#### 2.1.3 更新模式
- **增量更新**（默认）
  - 只处理新增或修改的文件
  - 通过文件 hash 或修改时间判断
  - 删除已失效的 STRM 文件

- **全量更新**
  - 清空目标目录重新生成
  - 适用于初次同步或数据修复

#### 2.1.4 目录映射
- **功能描述**：将 Alist 路径映射到本地 STRM 目录
- **管理方式**：通过 Web UI 管理，存储在 SQLite 数据库
- **配置参数**：
  - `name`：配置名称
  - `source`：Alist 源路径
  - `target`：本地 STRM 目标路径
  - `extensions`：视频扩展名列表（如：mp4, mkv, avi）
  - `concurrent`：并发数（默认 3，推荐 1-5）
  - `mode`：更新模式（incremental / full）
  - `strm_mode`：STRM 模式（alist_path / http_url）
  - `cron_expr`：Cron 表达式（可选，为空则不启用定时）
  - `enabled`：是否启用

### 2.2 任务调度

#### 2.2.1 定时任务（Cron）✨ **已优化**
- **功能描述**：按计划自动执行 STRM 生成任务
- **详细需求**：
  - **每配置独立定时**：每个映射配置有独立的 Cron 表达式
  - **可视化编辑器**：Web UI 提供可视化 Cron 配置
  - **6 种预设模式**：
    - 每隔 N 分钟（5/10/15/20/30）
    - 每小时（可指定分钟数）
    - 每天（时间选择器）
    - 每周（选择星期 + 时间）
    - 每月（选择日期 + 时间）
    - 自定义表达式（高级用户）
  - **执行预览**：显示最近三次执行时间
  - **动态管理**：配置增删改时自动更新定时任务
  - 任务执行日志记录

#### 2.2.2 手动触发
- **功能描述**：通过 Web UI 手动执行任务
- **详细需求**：
  - 选择特定配置执行
  - 选择更新模式（增量/全量）
  - 实时查看执行进度
  - 支持任务取消

#### 2.2.3 API 触发（差异化功能 🌟）
- **功能描述**：提供 RESTful API 供外部系统调用
- **应用场景**：
  - Alist Webhook 通知
  - 下载器完成后自动触发
  - 自动化工作流集成
- **接口设计**：
  ```
  POST /api/generate
  {
    "path": "/media/movies",
    "mode": "incremental"
  }
  ```

### 2.3 扩展功能

#### 2.3.1 元数据下载
- **功能描述**：下载媒体元数据文件
- **支持文件类型**：
  - `.nfo` - 媒体信息文件
  - `.jpg`/`.png` - 封面图片
  - `.srt`/`.ass` - 字幕文件
- **配置选项**：可选启用/禁用

#### 2.3.2 文件有效性检测
- **功能描述**：检测 STRM 文件链接是否有效
- **检测模式**：
  - 快速扫描：仅检查文件是否存在
  - 完整扫描：验证链接可访问性
- **失效处理**：
  - 标记失效文件
  - 可选自动删除失效 STRM

#### 2.3.3 并发处理
- **功能描述**：多任务并行执行，提高处理速度
- **详细需求**：
  - 多配置并行执行（多个路径同时处理）
  - 单配置内文件并发处理
  - 可配置 goroutine 数量
  - 并发安全和资源控制

### 2.4 Web UI

#### 2.4.1 管理界面
- **技术选型**：Gin + Vue.js 3（或简单的 HTML/JS）
- **核心页面**：
  1. **仪表盘**：系统状态、任务统计
  2. **配置管理**：添加/编辑/删除路径配置
  3. **任务中心**：手动触发、任务列表、执行历史
  4. **日志查看**：实时日志、日志搜索
  5. **系统设置**：全局参数、用户管理

#### 2.4.2 用户认证
- **认证方式**：简单的用户名密码
- **功能**：
  - 登录/登出
  - Session 管理
  - 可选：支持多用户（管理员/普通用户）

### 2.5 API 接口（差异化 🌟）

#### 2.5.1 核心接口
```
# 1. 生成 STRM（主要功能）
POST /api/generate
Request:
{
  "path": "/media/movies",      # 可选，不传则处理所有配置
  "mode": "incremental"         # 可选，默认 incremental
}
Response:
{
  "task_id": "uuid",
  "status": "running"
}

# 2. 查询任务状态
GET /api/tasks/{task_id}
Response:
{
  "task_id": "uuid",
  "status": "completed",        # running/completed/failed
  "files_created": 123,
  "files_deleted": 5,
  "errors": [],
  "started_at": "2025-01-01T00:00:00Z",
  "completed_at": "2025-01-01T00:05:00Z"
}

# 3. 获取配置列表
GET /api/configs
Response:
{
  "configs": [
    {
      "name": "电影库",
      "source": "/media/movies",
      "target": "/mnt/strm/movies",
      "enabled": true
    }
  ]
}

# 4. Webhook 接收（可选）
POST /api/webhook
Request:
{
  "event": "file.uploaded",
  "path": "/media/movies/new-movie.mp4"
}

# 5. 健康检查
GET /health
Response:
{
  "status": "ok",
  "version": "1.0.0"
}
```

#### 2.5.2 API 认证
- **认证方式**：简单 Token 认证
- **使用方式**：
  - 配置文件中设置固定 Token
  - 请求头携带：`X-API-Token: your-token`
  - Token 不匹配返回 401
- **可选功能**：IP 白名单

---

## 3. 实现说明

当前仓库已经有独立的现状文档，本 PRD 不再重复维护具体目录树、模块文件名和运行链路，避免需求文档与实现文档互相污染。

需要看当前真实实现时，请直接阅读：

- [系统架构](./system-architecture.md)
- [配置示例](../configs/config.example.yaml)
- [部署文档](../deployments/README.md)

---

## 4. 规划说明

当前阶段、里程碑、待完成项和后续版本方向已经单独整理到 [路线图](./roadmap.md)。

PRD 只保留需求本身，不再把阶段状态和实现进度混写在一起。

---

## 5. 非功能需求

### 5.1 性能要求
- 10,000 个文件的处理时间 < 5 分钟
- Web UI 响应时间 < 500ms
- API 响应时间 < 200ms
- 内存占用 < 200MB（正常运行）

### 5.2 可靠性要求
- 支持断点续传（任务失败后可恢复）
- 网络异常自动重试（最多 3 次）
- 数据一致性保证（SQLite 事务）

### 5.3 可用性要求
- 7x24 小时运行
- 优雅停机（保存任务状态）
- 自动恢复（重启后继续未完成任务）

### 5.4 安全要求
- API Token 认证
- 密码加密存储（bcrypt）
- 防止路径穿越攻击
- 输入验证和过滤

### 5.5 可维护性
- 代码注释覆盖率 > 30%
- 模块化设计，低耦合
- 完善的错误日志
- 易于调试和排查

---

## 6. 测试要求

### 6.1 单元测试
- 配置加载和验证
- Alist API 客户端
- STRM 生成逻辑
- 文件对比和去重

### 6.2 集成测试
- 端到端流程测试
- 定时任务执行
- API 接口测试
- Web UI 功能测试

### 6.3 性能测试
- 大量文件处理性能
- 并发任务执行
- 内存和 CPU 占用

详细测试清单、已完成项和覆盖进度请看 [测试计划](./TESTING.md)。

---

## 7. 部署方案

具体部署命令、Docker 使用方式、卷挂载、Webhook 集成和故障排查请看：

- [Docker 部署指南](../deployments/README.md)
- [Webhook 集成指南](../deployments/WEBHOOK.md)

---

## 8. 风险和挑战

### 8.1 技术风险
- **Alist API 变更**：可能需要适配新版本
- **网盘风控**：频繁调用可能触发限制
- **并发问题**：goroutine 泄漏和竞态条件

### 8.2 解决方案
- 版本兼容性测试，支持多个 Alist 版本
- 请求频率控制，合理设置定时任务间隔
- 使用 context 控制 goroutine 生命周期，加锁保护共享资源

---

## 9. 推荐配套工具

OpenList-STRM 专注于 STRM 文件生成，以下工具可以与其配合使用，构建完整的媒体库管理方案。

### 9.1 Emby/Jellyfin 插件

#### 🌟 Strm Assistant（神医助手）- 强烈推荐

- **GitHub**: [sjtuross/StrmAssistant](https://github.com/sjtuross/StrmAssistant)
- **类型**: Emby 官方插件
- **适用**: Emby 用户

**核心功能**：
- ✅ **提升 STRM 首播速度**：优化 STRM 文件首次加载性能
- ✅ **智能合并多版本**：同目录多个清晰度版本自动合并（1080p/4K）
- ✅ **片头片尾识别**：智能跳过片头片尾
- ✅ **中文搜索优化**：支持拼音、中文搜索
- ✅ **拼音首字母排序**：中文标题按拼音排序
- ✅ **外挂字幕扫描**：自动扫描外挂字幕文件
- ✅ **代理配置**：支持代理访问 TMDB
- ✅ **元数据增强**：TMDB 剧集组刮削、多语言海报

**安装方式**：
```
Emby 控制台 → 插件 → 目录 → 搜索 "Strm Assistant" → 安装
```

**版本选择**：
- 完整版：所有功能
- 精简版（StrmAssistant_less）：仅保留"中文搜索增强"和"代理配置"

**为什么推荐**：
- STRM 用户必备，显著提升播放体验
- 完美解决 STRM 文件加载慢的问题
- 中文媒体库友好

---

### 9.2 刮削工具（元数据管理）

OpenList-STRM v1.0 不包含刮削功能，建议使用以下专业工具：

#### MediaElch（推荐：轻量用户）⭐⭐⭐⭐⭐

- **官网**: [mediaelch.de](https://www.mediaelch.de/mediaelch/)
- **开发语言**: C++ (Qt)
- **平台**: Windows / macOS / Linux
- **特点**: 轻量、免费、界面现代化

**优势**：
- ✅ 轻量快速（内存占用 ~100-200MB）
- ✅ 完全免费开源
- ✅ 界面美观，易于使用
- ✅ 灵活配置：可为每条信息选择不同数据源
- ✅ 支持多源刮削（TMDB/IMDB/Fanart.tv）
- ✅ 支持音乐刮削

**适用场景**：
- 追求轻量和美观界面
- 不需要太复杂功能
- 免费开源爱好者

#### tinyMediaManager（推荐：专业用户）⭐⭐⭐⭐

- **官网**: [tinymediamanager.org](https://www.tinymediamanager.org/)
- **开发语言**: Java
- **平台**: Windows / macOS / Linux
- **特点**: 功能全面、准确率高

**优势**：
- ✅ 成熟稳定，社区活跃
- ✅ 刮削准确率高达 99%
- ✅ 本地化好，支持中文
- ✅ 批量重命名、NFO 生成
- ✅ 支持电影/电视剧/音乐
- ✅ 媒体库检查和修复

**注意**：
- ⚠️ 需要 Java 运行环境
- ⚠️ v4+ 高级功能需付费（€19.99）
- ⚠️ 内存占用较高（~300-500MB）

**适用场景**：
- 大型媒体库管理
- 需要精细控制刮削结果
- 愿意付费购买专业工具

#### Ember Media Manager（Windows 专用）

- **平台**: Windows Only
- **开发语言**: .NET
- **特点**: Windows 深度集成

**适用场景**：仅限 Windows 用户

---

### 9.3 字幕工具

#### ChineseSubFinder（中文字幕下载）⭐⭐⭐⭐⭐

- **GitHub**: [ChineseSubFinder/ChineseSubFinder](https://github.com/ChineseSubFinder/ChineseSubFinder)
- **开发语言**: Go
- **类型**: 独立工具

**核心功能**：
- ✅ 自动下载中文字幕
- ✅ 支持多个字幕网站（shooter、xunlei、zimuku、subhd 等）
- ✅ 集成 Emby/Jellyfin/Plex API
- ✅ 基于 IMDB ID 精准匹配
- ✅ 定时扫描媒体库
- ✅ 支持跳过已观看视频

**部署方式**：
- Docker 容器部署（推荐）
- 二进制文件部署

**配置示例**：
```yaml
emby:
  enabled: true
  url: "http://localhost:8096"
  api_key: "your-emby-api-key"
  skip_watched: true

subtitle_sources:
  - shooter
  - xunlei
  - zimuku

media_paths:
  - /mnt/strm/movies
  - /mnt/strm/tv
```

---

### 9.4 完整工作流推荐

#### 方案 A：轻量化方案（推荐新手）

```
1. Alist (云端存储)
   ↓
2. OpenList-STRM (生成 STRM 文件)
   ↓
3. MediaElch (刮削元数据 - 轻量免费)
   ↓
4. ChineseSubFinder (下载中文字幕 - 可选)
   ↓
5. Emby + Strm Assistant (播放和优化)
```

**优势**：
- 所有工具免费
- 资源占用低
- 适合入门用户

#### 方案 B：专业方案（推荐重度用户）

```
1. Alist (云端存储)
   ↓
2. OpenList-STRM (生成 STRM 文件)
   ↓
3. tinyMediaManager (刮削元数据 - 专业工具)
   ↓
4. ChineseSubFinder (下载中文字幕)
   ↓
5. Emby + Strm Assistant (播放和优化)
```

**优势**：
- 刮削准确率最高
- 功能最全面
- 适合大型媒体库

#### 方案 C：全自动化方案（推荐 NAS 用户）

```
1. Alist (云端存储)
   ↓
2. MoviePilot (订阅+下载+刮削+整理 - 一条龙)
   ↓
3. OpenList-STRM (通过 API 触发生成 STRM)
   ↓
4. ChineseSubFinder (自动下载字幕)
   ↓
5. Emby + Strm Assistant (播放和优化)
```

**优势**：
- 完全自动化
- 从订阅到播放无需人工干预
- 适合追剧用户

**注意**：
- MoviePilot 资源占用高（~500MB-1GB 内存）
- 配置较复杂

---

### 9.5 工具对比总结

| 工具类别 | 工具名称 | 轻量级 | 免费 | 推荐度 | 适用场景 |
|---------|---------|--------|------|--------|---------|
| **Emby 插件** | Strm Assistant | ✅ | ✅ | ⭐⭐⭐⭐⭐ | STRM 用户必装 |
| **刮削工具** | MediaElch | ✅ | ✅ | ⭐⭐⭐⭐⭐ | 轻量用户 |
| **刮削工具** | tinyMediaManager | ⚠️ | ⚠️ | ⭐⭐⭐⭐ | 专业用户 |
| **刮削工具** | MoviePilot | ❌ | ✅ | ⭐⭐⭐ | 全自动化 |
| **字幕工具** | ChineseSubFinder | ✅ | ✅ | ⭐⭐⭐⭐⭐ | 中文字幕需求 |

**图例**：
- ✅ 是 / ⚠️ 中等 / ❌ 否

---

## 10. 后续规划

后续版本方向已迁移到 [路线图](./roadmap.md)，本节不再重复维护。

---

## 11. 参考资料

### 11.1 核心技术文档
- [Alist API 文档](https://alist.nn.ci/guide/api/)
- [STRM 文件格式说明](https://kodi.wiki/view/Internet_video_and_audio_streams)
- [Gin 框架文档](https://gin-gonic.com/docs/)
- [Cron 表达式语法](https://crontab.guru/)

### 11.2 竞品项目
- [tefuirZ/alist-strm](https://github.com/tefuirZ/alist-strm) - Python 实现的 STRM 生成工具
- [imshuai/AlistAutoStrm](https://github.com/imshuai/AlistAutoStrm) - 另一个 STRM 生成工具
- [suxss/AList-STRM](https://github.com/suxss/AList-STRM) - AList STRM 生成器

### 11.3 配套工具
- [Strm Assistant (神医助手)](https://github.com/sjtuross/StrmAssistant) - Emby STRM 优化插件
- [MediaElch](https://www.mediaelch.de/mediaelch/) - 轻量级媒体刮削工具
- [tinyMediaManager](https://www.tinymediamanager.org/) - 专业媒体管理工具
- [ChineseSubFinder](https://github.com/ChineseSubFinder/ChineseSubFinder) - 中文字幕自动下载
- [MoviePilot](https://github.com/jxxghp/MoviePilot) - NAS 媒体库自动化管理

### 11.4 媒体服务器
- [Emby](https://emby.media/) - 媒体服务器
- [Jellyfin](https://jellyfin.org/) - 开源媒体服务器
- [Plex](https://www.plex.tv/) - 媒体服务器

---

**文档版本**：v1.1
**最后更新**：2025-01-04
**维护者**：OpenList-STRM Team
