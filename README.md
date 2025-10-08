# OpenList-STRM

OpenList-STRM 是一个基于 Go 语言开发的 STRM 文件生成工具，用于将 Alist 网盘中的媒体文件批量生成为 STRM 格式，供 Emby、Jellyfin、Plex 等流媒体服务器使用。

## ✨ 特性

- 🚀 **高性能**：Go 语言实现，并发处理，性能优于 Python 实现
- 💾 **免挂载**：无需挂载网盘，通过 STRM 文件直接播放
- 📦 **节省空间**：本地只存储小体积的 STRM 文件
- ⏰ **灵活调度**：每个配置独立的定时任务，可视化 Cron 编辑器
- 🔄 **增量更新**：智能增量同步，只处理新增和修改的文件
- 🎯 **简单易用**：单二进制文件部署，Web UI 可视化配置
- 🌐 **Web UI**：现代化 Vue 3 界面，无需编辑配置文件
- 🔌 **API 接口**：RESTful API，支持外部程序调用
- 🎬 **MediaWarp 支持**：支持 302 重定向代理，优化播放体验
- 🔔 **自动通知**：支持 Emby/Jellyfin 自动扫描，无缝更新媒体库

## 📋 当前版本

**v1.0.0** - 完整功能版本

已实现功能：
- ✅ Alist API 集成
- ✅ STRM 文件生成（支持 alist_path 和 http_url 两种模式）
- ✅ 增量/全量更新模式
- ✅ **每配置独立定时任务**（可视化 Cron 编辑器）
- ✅ SQLite 数据存储
- ✅ 每配置独立并发控制（防风控）
- ✅ 日志系统
- ✅ **RESTful API 接口**（完整的 CRUD 操作）
- ✅ **Vue 3 Web UI 管理界面**（数据库配置管理）
- ✅ **Webhook 支持**（自动触发任务）
- ✅ **Docker 部署**
- ✅ **MediaWarp 集成支持**
- ✅ **媒体服务器通知**（支持 Emby/Jellyfin 自动扫描）

待实现功能（后续版本）：
- ⏳ 元数据下载
- ⏳ 文件有效性检测

## 🚀 快速开始

### 1. 下载

```bash
# 下载二进制文件（编译后）
# 或者从源码编译
git clone https://github.com/konghang/openlist-strm.git
cd openlist-strm
make build
```

### 2. 配置

```bash
# 复制配置文件示例
cp configs/config.example.yaml config.yaml

# 编辑配置文件
vim config.yaml
```

**最小配置**：

```yaml
# 服务器配置
server:
  host: "0.0.0.0"
  port: 8080

# Alist 配置（必需）
alist:
  url: "http://your-alist-url:5244"
  token: "your-alist-token"
  sign_enabled: false
  timeout: 30

# 数据库配置
database:
  path: "./data/openlist-strm.db"

# 日志配置
log:
  level: "info"
  file: ""  # 留空则输出到 stdout（Docker 推荐）
            # 设置路径则输出到文件，如 "./logs/openlist-strm.log"
```

**注意**：路径映射（mappings）现在通过 Web UI 管理，不再在配置文件中设置。

### 3. 运行

```bash
./bin/openlist-strm -config config.yaml
```

### 4. 访问 Web UI

服务启动后，访问 Web 管理界面：

```
http://localhost:8080
```

Web UI 提供以下功能：
- 📊 **仪表盘**：查看系统状态、快速操作、最近任务
- 📋 **任务管理**：查看任务列表、执行状态、详细信息
- ⚙️ **配置管理**：
  - 创建/编辑/删除路径映射
  - 配置视频扩展名、并发数、更新模式
  - 选择 STRM 模式（Alist 路径或直链 URL）
  - 可视化 Cron 编辑器，设置独立定时任务
  - 查看最近三次执行时间预览
  - 手动触发生成任务

## 📖 配置说明

### Alist 配置

```yaml
alist:
  url: "http://localhost:5244"  # Alist 服务地址
  token: "your-alist-token"     # Alist API Token
  sign_enabled: false           # 是否启用签名
  timeout: 30                   # 请求超时（秒）
```

### 路径映射配置

**路径映射现在通过 Web UI 管理**，不再在配置文件中设置。

在 Web UI 中创建配置时，需要设置以下参数：

| 参数 | 说明 | 示例 |
|------|------|------|
| 配置名称 | 映射的标识名称 | `Movies` |
| 源路径 | Alist 中的路径 | `/media/movies` |
| 目标路径 | 本地 STRM 文件保存路径 | `/mnt/strm/movies` |
| 视频扩展名 | 需要处理的视频格式 | `mp4, mkv, avi, mov` |
| 并发数 | 同时处理的文件数量 | `3`（推荐 1-5，防风控） |
| 更新模式 | 增量或全量 | `incremental` / `full` |
| STRM 模式 | 路径或直链 | `alist_path` / `http_url` |
| 定时任务 | Cron 表达式（可选） | `0 2 * * *` |
| 启用状态 | 是否启用此配置 | `true` / `false` |

### STRM 模式说明

**alist_path 模式**（推荐搭配 MediaWarp）：
- STRM 文件内容为 Alist 路径：`/media/movies/movie.mp4`
- 需要配合 [MediaWarp](https://github.com/AkimioJR/MediaWarp) 使用
- MediaWarp 负责 302 重定向获取实际播放链接
- 优点：支持 Alist 签名、CDN 切换等高级功能

**http_url 模式**（直接播放）：
- STRM 文件内容为完整 URL：`http://alist.example.com/d/media/movies/movie.mp4`
- 直接播放，无需额外组件
- 适合简单场景

### 定时任务配置

**每个配置可以有独立的定时任务**，通过 Web UI 的可视化编辑器设置：

**预设模式**：
- 每隔 N 分钟（5/10/15/20/30 分钟）
- 每小时（可指定分钟数）
- 每天（时间选择器）
- 每周（选择星期 + 时间）
- 每月（选择日期 + 时间）
- 自定义表达式

**Cron 表达式示例**：
- `*/30 * * * *` - 每 30 分钟
- `0 * * * *` - 每小时整点
- `0 2 * * *` - 每天凌晨 2 点
- `0 2 * * 0` - 每周日凌晨 2 点
- `0 2 1 * *` - 每月 1 号凌晨 2 点

**执行时间预览**：
- 编辑器会实时显示最近三次执行时间
- 帮助验证 Cron 表达式是否正确

### API 配置

```yaml
api:
  enabled: true
  token: ""  # 可选，留空则不需要认证
  timeout: 300
```

### Web UI 配置

```yaml
web:
  enabled: true
  username: "admin"      # 保留字段，当前版本未使用
  password: "admin123"   # 保留字段，当前版本未使用
```

### 媒体服务器通知配置

OpenList-STRM 支持在生成 STRM 文件后自动通知 Emby 或 Jellyfin 扫描媒体库，实现自动更新媒体库内容。

```yaml
media_server:
  enabled: false  # 是否启用媒体服务器通知
  type: "emby"    # 媒体服务器类型: emby, jellyfin, both

  # Emby 配置
  emby:
    url: "http://emby:8096"        # Emby 服务器地址
    api_key: "your-emby-api-key"   # Emby API Key
    scan_mode: "full"              # 扫描模式: full=全库扫描, path=路径扫描
    # 路径映射配置（仅在 scan_mode=path 时需要）
    path_mapping:
      # OpenList-STRM 容器路径 -> Emby 容器路径
      # "/data/strm": "/media/movies"

  # Jellyfin 配置
  jellyfin:
    url: "http://jellyfin:8096"        # Jellyfin 服务器地址
    api_key: "your-jellyfin-api-key"   # Jellyfin API Key
    scan_mode: "full"                  # 扫描模式: full=全库扫描, path=路径扫描
    # 路径映射配置（仅在 scan_mode=path 时需要）
    path_mapping:
      # OpenList-STRM 容器路径 -> Jellyfin 容器路径
      # "/data/strm": "/media/movies"
```

**配置说明**：

| 参数 | 说明 | 可选值 |
|------|------|--------|
| `enabled` | 是否启用通知功能 | `true` / `false` |
| `type` | 媒体服务器类型 | `emby` / `jellyfin` / `both` |
| `url` | 媒体服务器地址 | 如：`http://emby:8096` |
| `api_key` | API 密钥 | 在媒体服务器设置中获取 |
| `scan_mode` | 扫描模式 | `full` / `path` |
| `path_mapping` | 路径映射（可选） | 仅在 `path` 模式时需要 |

**扫描模式说明**：

1. **全局扫描模式（full）** - 推荐
   - 触发完整媒体库扫描
   - 配置简单，无需路径映射
   - 适合大多数场景
   - 缺点：扫描时间较长

2. **路径扫描模式（path）** - 高级
   - 仅扫描 STRM 文件所在路径
   - 需要配置路径映射
   - 扫描速度快，资源占用小
   - 要求：OpenList-STRM 和媒体服务器的路径映射必须一致

**路径映射示例**：

```yaml
# 场景：Docker 容器间路径不一致
# OpenList-STRM 容器:  /data/strm/Movies
# Emby 容器:          /media/Movies

media_server:
  enabled: true
  type: "emby"
  emby:
    url: "http://emby:8096"
    api_key: "your-api-key"
    scan_mode: "path"
    path_mapping:
      "/data/strm": "/media"  # 将 /data/strm 映射为 /media
```

**获取 API Key**：

- **Emby**：设置 → 高级 → API 密钥 → 新建应用程序
- **Jellyfin**：设置 → API 密钥 → 添加 API 密钥

**通知触发条件**：

- 仅在有文件创建或删除时触发通知
- 如果任务没有变更文件，则跳过通知
- 通知失败不影响任务完成状态，仅记录日志

**Docker Compose 配置示例**：

```yaml
services:
  openlist-strm:
    image: konghanghang/openlist-strm:master
    volumes:
      - ./strm:/data/strm
    environment:
      - TZ=Asia/Shanghai

  emby:
    image: emby/embyserver
    volumes:
      - ./strm:/media  # 注意：路径要与 path_mapping 对应
    environment:
      - TZ=Asia/Shanghai
```

## 🛠️ 构建

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/konghang/openlist-strm.git
cd openlist-strm

# 构建前端
cd web
npm install
npm run build
cd ..

# 构建后端（前端已自动嵌入）
make build

# 或者手动编译
go build -o bin/openlist-strm ./cmd/server
```

构建后的二进制文件包含：
- Go 后端服务
- 嵌入的 Vue 3 前端资源
- 单文件部署，无需额外依赖

## 🔌 API 接口

OpenList-STRM 提供 RESTful API 供外部程序调用。

### 生成 STRM 文件

```bash
# 生成所有映射
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"mode": "incremental"}'

# 生成指定路径
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"path": "Movies", "mode": "full"}'
```

### 查询任务状态

```bash
# 获取任务列表
curl http://localhost:8080/api/tasks

# 获取指定任务
curl http://localhost:8080/api/tasks/{task_id}
```

### 获取配置

```bash
# 获取所有路径映射配置
curl http://localhost:8080/api/configs
```

### Webhook 接口

接收外部系统（如 Alist、下载器）的通知，自动触发 STRM 生成：

```bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "file.upload",
    "path": "/media/movies/new-movie.mp4",
    "action": "add"
  }'
```

**响应示例**：
```json
{
  "success": true,
  "message": "webhook received, generation triggered",
  "task_id": "uuid-string"
}
```

**使用场景**：
- Alist Webhook 通知文件上传
- 下载器完成后自动触发
- 自动化工作流集成

### API 认证

如果在配置文件中设置了 API Token：

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Authorization: Bearer your-api-token" \
  -H "Content-Type: application/json" \
  -d '{"mode": "incremental"}'
```

### Trace ID 日志追踪

每个任务执行都会生成唯一的 Trace ID（取 Task ID 前 8 位），用于关联所有相关日志：

**任务级日志**：
```
[TraceID: abc12345] Task started: mapping=Movies, mode=incremental, source=/media/movies
[TraceID: abc12345] Scanning source directory: /media/movies
[TraceID: abc12345] Found 150 video files to process
[TraceID: abc12345] Task COMPLETED: created=10, deleted=2, skipped=140, errors=0, duration=3.5s
```

**文件级日志**（每个文件处理状态）：
```
[TraceID: abc12345] ✅ CREATED: /media/movies/Movie1.mp4
[TraceID: abc12345] ⏭️  SKIPPED: /media/movies/Movie2.mp4 (already exists)
[TraceID: abc12345] ❌ ERROR: /media/movies/Movie3.mp4 -> failed to get URL: timeout
```

**查询任务日志**：
```bash
# 通过 Trace ID 过滤所有日志
grep "TraceID: abc12345" logs/openlist-strm.log

# 只看任务级日志（排除文件级）
grep "TraceID: abc12345" logs/openlist-strm.log | grep -v "CREATED\|SKIPPED\|ERROR:"

# 只看创建的文件
grep "TraceID: abc12345" logs/openlist-strm.log | grep "✅ CREATED"

# 只看错误的文件
grep "TraceID: abc12345" logs/openlist-strm.log | grep "❌ ERROR"

# 统计各状态文件数量
grep "TraceID: abc12345" logs/openlist-strm.log | grep -E "✅|⏭️|❌" | wc -l
```

**Trace ID 来源**：
- API 调用：返回的 `task_id` 即为完整 Trace ID
- Webhook 触发：响应中的 `task_id` 即为完整 Trace ID
- 定时任务：自动生成，从日志中查看
- 手动触发：Web UI 任务列表中显示

## 🐳 Docker 部署

### 拉取预构建镜像

```bash
# 从 Docker Hub 拉取（国内推荐）
docker pull konghanghang/openlist-strm:master

# 或从 GitHub Container Registry 拉取
docker pull ghcr.io/konghanghang/openlist-strm:master
```

### 使用 Docker Compose（推荐）

```bash
# 创建工作目录
mkdir openlist-strm && cd openlist-strm

# 下载示例配置
wget https://raw.githubusercontent.com/konghanghang/openlist-strm/master/configs/config.example.yaml -O config.yaml

# 编辑配置（主要配置 Alist URL 和 Token）
vim config.yaml

# 创建 docker-compose.yml 并启动
docker-compose up -d
```

**docker-compose.yml 示例：**

```yaml
services:
  openlist-strm:
    image: konghanghang/openlist-strm:master
    container_name: openlist-strm
    restart: unless-stopped
    ports:
      - 8080:8080
    volumes:
      - ./config.yaml:/app/configs/config.yaml:ro
      - ./data:/app/data
      - ./strm:/mnt/strm
    environment:
      - TZ=Asia/Shanghai
```

### 查看日志

```bash
# 实时查看日志
docker logs -f openlist-strm

# 查看最近 100 行日志
docker logs --tail 100 openlist-strm

# 查看带时间戳的日志
docker logs -t openlist-strm
```

**注意**：默认配置日志输出到 stdout，通过 `docker logs` 查看即可。如需文件日志，请在配置中设置 `log.file` 并挂载 `/app/logs` 目录。

### 使用 Docker

```bash
# 运行容器
docker run -d \
  --name openlist-strm \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v /path/to/strm:/mnt/strm \
  -e TZ=Asia/Shanghai \
  konghanghang/openlist-strm:master
```

详细部署文档请查看：[deployments/README.md](./deployments/README.md)

## 📦 推荐配套工具

### 302 重定向代理

- **[MediaWarp](https://github.com/AkimioJR/MediaWarp)** ⭐⭐⭐⭐⭐
  - 支持 Emby/Jellyfin STRM 302 重定向
  - 支持 Alist 签名、CDN 切换
  - 完美配合 alist_path 模式使用
  - 推荐搭配使用

### Emby 插件

- **[Strm Assistant (神医助手)](https://github.com/sjtuross/StrmAssistant)** ⭐⭐⭐⭐⭐
  - 优化 STRM 播放速度
  - 中文搜索和排序
  - STRM 用户必备

### 刮削工具

- **[MediaElch](https://www.mediaelch.de/)** - 轻量级，免费
- **[tinyMediaManager](https://www.tinymediamanager.org/)** - 专业，准确率高

### 字幕工具

- **[ChineseSubFinder](https://github.com/ChineseSubFinder/ChineseSubFinder)** - 中文字幕自动下载

## 📚 文档

详细文档请查看：
- [项目需求文档 (PRD)](./docs/PRD.md)
- [测试计划文档 (TESTING)](./docs/TESTING.md)
- [配置文件示例](./configs/config.example.yaml)
- [Docker 部署指南](./deployments/README.md)
- [Webhook 集成指南](./deployments/WEBHOOK.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

- [Alist](https://alist.nn.ci/) - 优秀的文件列表程序
- [tefuirZ/alist-strm](https://github.com/tefuirZ/alist-strm) - 项目灵感来源

---

**🤖 Powered by Go | Made with ❤️**
