# OpenList-STRM 项目需求文档 (PRD)

## 1. 项目概述

### 1.1 项目简介
OpenList-STRM 是一个基于 Go 语言开发的 STRM 文件生成工具，用于将 Alist 网盘中的媒体文件批量生成为 STRM 格式，供 Emby、Jellyfin、Plex 等流媒体服务器使用。

### 1.2 核心价值
- **免挂载**：无需挂载网盘，通过 STRM 文件直接播放
- **节省空间**：本地只存储小体积的 STRM 文件
- **自动同步**：支持定时任务和 API 触发，自动更新媒体库
- **高性能**：Go 语言实现，并发处理，性能优于 Python 实现

### 1.3 技术栈
- **后端**：Go 1.23+
- **Web 框架**：Gin
- **数据库**：SQLite (GORM)
- **前端**：Vue.js 3 + Element Plus + Vite
- **部署**：单二进制文件（前端嵌入）/ Docker（待实现）

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
  - 支持常见视频格式：mp4, mkv, avi, mov, flv, wmv 等
  - 文件名处理：支持中文、特殊字符、空格
  - STRM 文件内容：包含 Alist 直链 URL
  - 支持自定义文件过滤规则

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
- **配置示例**：
  ```yaml
  mappings:
    - name: "电影库"
      source: "/media/movies"        # Alist 路径
      target: "/mnt/strm/movies"     # 本地 STRM 路径
      mode: "incremental"            # 更新模式
    - name: "电视剧"
      source: "/media/tv"
      target: "/mnt/strm/tv"
      mode: "incremental"
  ```

### 2.2 任务调度

#### 2.2.1 定时任务（Cron）
- **功能描述**：按计划自动执行 STRM 生成任务
- **详细需求**：
  - 支持标准 Cron 表达式
  - 每个路径可独立配置定时任务
  - 全局定时任务开关
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
POST /api/webhook/alist
Request:
{
  "event": "file.uploaded",
  "path": "/media/movies/new-movie.mp4"
}

# 5. 健康检查
GET /api/health
Response:
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": 3600
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

## 3. 技术架构

### 3.1 项目结构
```
openlist-strm/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口
├── internal/
│   ├── config/                  # 配置管理
│   │   ├── config.go
│   │   └── loader.go
│   ├── alist/                   # Alist API 客户端
│   │   ├── client.go
│   │   └── types.go
│   ├── strm/                    # STRM 生成器
│   │   ├── generator.go
│   │   └── validator.go
│   ├── scheduler/               # 任务调度器
│   │   ├── cron.go
│   │   └── task.go
│   ├── storage/                 # 数据存储
│   │   ├── sqlite.go
│   │   └── models.go
│   ├── api/                     # HTTP API
│   │   ├── handlers.go
│   │   ├── middleware.go
│   │   └── routes.go
│   └── web/                     # Web UI
│       ├── handler.go
│       └── static/
├── web/                         # 前端代码（Vue 3 + Vite）
│   ├── src/
│   │   ├── views/               # 页面组件
│   │   ├── router/              # 路由配置
│   │   ├── api/                 # API 封装
│   │   └── App.vue
│   ├── dist/                    # 构建产物
│   ├── public/
│   ├── package.json
│   └── vite.config.js
├── configs/
│   └── config.example.yaml      # 配置示例
├── deployments/
│   ├── Dockerfile
│   └── docker-compose.yml
├── docs/
│   ├── README.md
│   ├── API.md
│   └── FAQ.md
├── go.mod
├── go.sum
└── Makefile
```

### 3.2 核心模块

#### 3.2.1 配置管理模块
- **职责**：加载、验证、管理配置
- **技术**：viper（YAML 解析）
- **配置文件结构**：
  ```yaml
  # 全局配置
  server:
    host: "0.0.0.0"
    port: 8080

  # Alist 配置
  alist:
    url: "http://localhost:5244"
    token: "your-alist-token"
    sign_enabled: false
    timeout: 30

  # STRM 配置
  strm:
    output_dir: "/mnt/strm"
    concurrent: 10                    # 并发数
    extensions:                        # 视频格式
      - mp4
      - mkv
      - avi
    download_metadata: true            # 下载元数据

  # 路径映射
  mappings:
    - name: "电影"
      source: "/media/movies"
      target: "/mnt/strm/movies"
      mode: "incremental"
      enabled: true

  # 定时任务
  schedule:
    enabled: true
    cron: "0 2 * * *"                  # 每天凌晨2点

  # API 配置
  api:
    enabled: true
    token: "your-api-token"            # 可选
    timeout: 300

  # Web UI
  web:
    enabled: true
    username: "admin"
    password: "admin123"

  # 日志
  log:
    level: "info"                      # debug/info/warn/error
    file: "/var/log/openlist-strm.log"
    max_size: 100                      # MB
    max_backups: 3
  ```

#### 3.2.2 Alist 客户端模块
- **职责**：与 Alist API 交互
- **主要功能**：
  - 获取文件列表（递归）
  - 生成文件直链
  - 签名支持
  - 请求重试

#### 3.2.3 STRM 生成器模块
- **职责**：生成和管理 STRM 文件
- **主要功能**：
  - 创建 STRM 文件
  - 目录结构同步
  - 文件去重和校验
  - 元数据下载

#### 3.2.4 任务调度模块
- **职责**：管理任务执行
- **技术**：robfig/cron（定时任务）
- **主要功能**：
  - Cron 定时任务
  - 手动任务触发
  - 任务队列管理
  - 并发控制

#### 3.2.5 数据存储模块
- **职责**：持久化数据存储
- **技术**：gorm + SQLite
- **数据表设计**：
  ```sql
  -- 文件状态表
  CREATE TABLE files (
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE NOT NULL,
    size INTEGER,
    modified_at DATETIME,
    hash TEXT,
    strm_path TEXT,
    created_at DATETIME,
    updated_at DATETIME
  );

  -- 任务历史表
  CREATE TABLE tasks (
    id INTEGER PRIMARY KEY,
    task_id TEXT UNIQUE NOT NULL,
    config_name TEXT,
    mode TEXT,
    status TEXT,
    files_created INTEGER,
    files_deleted INTEGER,
    errors TEXT,
    started_at DATETIME,
    completed_at DATETIME
  );

  -- 用户表
  CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT,
    created_at DATETIME
  );
  ```

#### 3.2.6 HTTP API 模块
- **职责**：提供 RESTful API
- **技术**：Gin 框架
- **中间件**：
  - 日志中间件
  - 认证中间件（Token/Session）
  - CORS 中间件
  - 限流中间件

#### 3.2.7 Web UI 模块
- **职责**：提供管理界面
- **技术方案**：Vue.js 3 + Element Plus + Vite
- **页面结构**：
  - 仪表盘：系统状态、快速操作、最近任务
  - 任务管理：任务列表、执行状态、详情查看
  - 配置管理：路径映射配置、手动触发
- **部署方式**：构建后嵌入 Go 二进制文件（embed.FS）

---

## 4. 开发计划

### 4.1 阶段划分

#### Phase 1: MVP（已完成 ✅）
**目标**：实现核心功能，CLI 可用

- [x] 项目初始化
- [x] 配置管理模块
- [x] Alist 客户端实现
- [x] STRM 生成器（基础版）
- [x] 增量/全量更新逻辑
- [x] 定时任务（Cron）
- [x] SQLite 存储
- [x] 基础日志

**交付物**：
- ✅ 可执行的 CLI 工具
- ✅ 配置文件示例
- ✅ 基础 README

#### Phase 2: Web UI + API（已完成 ✅）
**目标**：添加 Web UI 和 RESTful API

- [x] RESTful API 实现
- [x] API Token 认证
- [x] 任务状态查询接口
- [x] Vue 3 项目结构
- [x] Web UI 基础框架
- [x] 配置管理界面
- [x] 任务管理界面
- [x] 仪表盘界面
- [x] 前端构建和嵌入

**交付物**：
- ✅ 带 Web UI 的完整应用
- ✅ RESTful API 接口
- ✅ 单二进制文件部署
- ✅ 更新文档

#### Phase 3: 扩展功能（部分完成 ✅）
**目标**：添加高级功能和优化

- [x] Docker 打包
- [x] Docker Compose 配置
- [x] Webhook 支持
- [ ] 元数据下载
- [ ] 文件有效性检测
- [ ] UI 优化和完善
- [ ] ~~实时文件监控~~（已取消，避免触发风控）

**交付物**：
- ✅ Docker 镜像和 Dockerfile
- ✅ Docker Compose 配置
- ✅ 部署文档
- ✅ Webhook API 接口

#### Phase 4: 优化和发布（1-2 天）
**目标**：性能优化和文档完善

- [ ] 性能测试和优化
- [ ] 错误处理完善
- [ ] 完整的中英文文档
- [ ] 单元测试（核心模块）
- [ ] CI/CD 配置
- [ ] 发布 v1.0.0

**交付物**：
- 生产级应用
- 完整文档
- GitHub Release

### 4.2 里程碑

| 里程碑 | 状态 | 目标 |
|--------|------|------|
| M1: MVP 完成 | ✅ 已完成 | CLI 工具可用，核心功能完成 |
| M2: Web UI + API 完成 | ✅ 已完成 | 带管理界面和 API 的完整应用 |
| M3: Docker + Webhook | ✅ 已完成 | Docker 部署、Webhook 集成 |
| M4: v1.0.0 发布 | ✅ 当前版本 | 生产可用，功能完整 |
| M5: 扩展功能 | 🔄 规划中 | 元数据下载、文件检测等 |

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

## 6. 测试计划

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

---

## 7. 部署方案

### 7.1 Docker 部署（推荐）
```bash
docker run -d \
  --name openlist-strm \
  -p 8080:8080 \
  -v /path/to/config:/config \
  -v /path/to/strm:/strm \
  -v /path/to/data:/data \
  openlist-strm:latest
```

### 7.2 Docker Compose 部署
```yaml
version: '3.8'
services:
  openlist-strm:
    image: openlist-strm:latest
    container_name: openlist-strm
    ports:
      - "8080:8080"
    volumes:
      - ./config:/config
      - ./strm:/strm
      - ./data:/data
    environment:
      - TZ=Asia/Shanghai
    restart: unless-stopped
```

### 7.3 二进制部署
```bash
# 下载二进制文件
wget https://github.com/xxx/openlist-strm/releases/download/v1.0.0/openlist-strm-linux-amd64

# 赋予执行权限
chmod +x openlist-strm-linux-amd64

# 运行
./openlist-strm-linux-amd64 --config config.yaml
```

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

### 10.1 v1.1 版本
- [ ] 支持更多网盘（WebDAV、OneDrive 等）
- [ ] 智能分类（电影/电视剧自动识别）
- [ ] 简单刮削功能（可选，基于 TMDB API）

### 10.2 v1.2 版本
- [ ] 图形化安装向导
- [ ] 多语言支持（i18n）
- [ ] 插件系统（可扩展）

### 10.3 v2.0 版本
- [ ] 分布式部署（多节点）
- [ ] 消息队列（RabbitMQ/Redis）
- [ ] 监控告警（Prometheus）

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
