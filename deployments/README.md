# Docker 部署指南

本目录包含 OpenList-STRM 的 Docker 部署文档。Docker 配置文件位于项目根目录。

## 快速开始

### 1. 使用 Docker Compose（推荐）

```bash
# 返回项目根目录
cd ..

# 复制配置文件
cp configs/config.example.yaml config.yaml

# 编辑配置文件，填入你的 Alist URL 和 Token
vim config.yaml

# 启动服务（使用根目录的 docker-compose.yml）
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 2. 使用 Docker（不使用 Compose）

```bash
# 运行容器（使用预构建镜像）
docker run -d \
  --name openlist-strm \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v /path/to/your/strm:/mnt/strm \
  -e TZ=Asia/Shanghai \
  konghanghang/openlist-strm:master
```

## 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `TZ` | 时区 | `Asia/Shanghai` |

### 挂载卷

| 主机路径 | 容器路径 | 说明 |
|----------|----------|------|
| `./config.yaml` | `/app/configs/config.yaml` | 配置文件（只读） |
| `./data` | `/app/data` | 数据库文件 |
| `/your/strm/path` | `/mnt/strm` | STRM 输出目录 |
| `./logs`（可选） | `/app/logs` | 日志文件（仅在配置 `log.file` 时需要） |

### 端口

- `8080`: Web UI 和 API 端口

## 常用命令

```bash
# 启动服务
docker-compose up -d

# 停止服务
docker-compose down

# 查看日志
docker-compose logs -f

# 重启服务
docker-compose restart

# 重新构建并启动
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 进入容器
docker-compose exec openlist-strm sh
```

## 与 Alist 集成

### 方案 1: Alist 在宿主机运行

配置文件中 Alist URL 使用宿主机 IP：

```yaml
alist:
  url: "http://192.168.1.100:5244"
  token: "your-token"
```

### 方案 2: Alist 也在 Docker 中运行

在根目录的 `docker-compose.yml` 中添加 Alist 服务，然后使用服务名访问：

```yaml
alist:
  url: "http://alist:5244"
  token: "your-token"
```

示例 docker-compose.yml 配置：
```yaml
services:
  openlist-strm:
    # ... openlist-strm 配置

  alist:
    image: xhofe/alist:latest
    container_name: alist
    volumes:
      - ./alist/data:/opt/alist/data
    ports:
      - "5244:5244"
    networks:
      - openlist-network
```

## 持久化数据

确保以下目录被正确挂载以持久化数据：

- `./data` — 数据库文件
- `/your/strm/path` — STRM 文件输出
- `./logs`（可选）— 日志文件，仅在配置 `log.file` 时需要，默认输出到 stdout

## 健康检查

容器内置健康检查，每 30 秒检查一次 `/health` 端点。

查看健康状态：
```bash
docker inspect --format='{{.State.Health.Status}}' openlist-strm
```

## 故障排查

### 容器无法启动

```bash
# 查看容器日志
docker-compose logs openlist-strm

# 检查配置文件是否正确挂载
docker-compose exec openlist-strm cat /app/configs/config.yaml
```

### 无法访问 Alist

```bash
# 进入容器测试网络
docker-compose exec openlist-strm sh
wget -O- http://your-alist-url:5244

# 检查是否在同一网络中
docker network inspect openlist-strm-network
```

### 权限问题

容器以 `app` 用户（UID:1000, GID:1000）运行，确保挂载的目录权限正确：

```bash
# 修改目录权限
sudo chown -R 1000:1000 ./data ./logs
```

## 媒体服务器通知

任务完成后可自动通知 Emby / Jellyfin 扫描媒体库。在 `config.yaml` 中配置：

```yaml
media_server:
  enabled: true
  type: "emby"  # emby / jellyfin / both

  emby:
    url: "http://emby:8096"
    api_key: "your-emby-api-key"
    scan_mode: "full"  # full=全库扫描, path=路径扫描
    path_mapping:      # 仅 path 模式需要
      "/data/strm": "/media"
```

### 扫描模式

| 模式 | 说明 | 优点 | 缺点 |
|------|------|------|------|
| `full` | 触发完整媒体库扫描 | 配置简单，无需路径映射 | 扫描时间较长 |
| `path` | 仅扫描 STRM 文件所在路径 | 速度快，资源占用小 | 需要配置路径映射 |

### 路径映射

当 OpenList-STRM 和媒体服务器的容器路径不一致时，需要配置路径映射：

```
OpenList-STRM 容器: /data/strm/Movies
Emby 容器:         /media/Movies
→ path_mapping: "/data/strm": "/media"
```

Docker Compose 示例：

```yaml
services:
  openlist-strm:
    image: konghanghang/openlist-strm:master
    volumes:
      - ./strm:/data/strm

  emby:
    image: emby/embyserver
    volumes:
      - ./strm:/media  # 路径要与 path_mapping 对应
```

### 获取 API Key

- **Emby**：设置 → 高级 → API 密钥 → 新建应用程序
- **Jellyfin**：设置 → API 密钥 → 添加 API 密钥

### 触发条件

- 仅在有文件创建或删除时触发通知
- 无变更文件时跳过通知
- 通知失败不影响任务完成状态，仅记录日志

## Trace ID 日志追踪

每个任务生成唯一 Trace ID（Task ID 前 8 位），用于关联所有相关日志。

任务级日志：
```
[TraceID: abc12345] Task started: mapping=Movies, mode=incremental, source=/media/movies
[TraceID: abc12345] Found 150 video files to process
[TraceID: abc12345] Task COMPLETED: created=10, deleted=2, skipped=140, errors=0, duration=3.5s
```

文件级日志：
```
[TraceID: abc12345] ✅ CREATED: /media/movies/Movie1.mp4
[TraceID: abc12345] ⏭️  SKIPPED: /media/movies/Movie2.mp4 (already exists)
[TraceID: abc12345] ❌ ERROR: /media/movies/Movie3.mp4 -> failed to get URL: timeout
```

### 日志过滤

```bash
# 通过 Trace ID 过滤所有日志
grep "TraceID: abc12345" logs/openlist-strm.log

# 只看创建的文件
grep "TraceID: abc12345" logs/openlist-strm.log | grep "✅ CREATED"

# 只看错误
grep "TraceID: abc12345" logs/openlist-strm.log | grep "❌ ERROR"
```

### Trace ID 来源

| 触发方式 | 获取方式 |
|---------|---------|
| API 调用 | 响应中的 `task_id` |
| Webhook | 响应中的 `task_id` |
| 定时任务 | 日志中查看 |
| Web UI 手动触发 | 任务列表中显示 |

## 更新

```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build
```

## 备份

### 备份数据库

```bash
# 备份数据库文件
cp ./data/openlist-strm.db ./data/openlist-strm.db.backup
```

### 备份配置

```bash
# 备份配置文件
cp ./config.yaml ./config.yaml.backup
```
