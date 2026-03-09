# OpenList-STRM

[English](./README_EN.md) | 中文

将 Alist 网盘中的媒体文件批量生成 STRM 文件，供 Emby / Jellyfin / Plex 直接播放。Go 语言实现，单二进制部署，Web UI 管理。

## 特性

- **免挂载** — 无需挂载网盘，STRM 文件直接播放
- **高性能** — Go 并发处理，每配置独立并发控制（防风控）
- **增量更新** — 智能增量同步，只处理变更文件
- **独立调度** — 每个配置独立 Cron 定时任务，可视化编辑器
- **双模式** — 支持 `alist_path`（配合 MediaWarp）和 `http_url`（直链）
- **Web UI** — Vue 3 现代界面，仪表盘 + 配置管理 + 任务监控
- **Webhook** — 接收外部通知自动触发，支持路径映射
- **媒体通知** — 任务完成后自动通知 Emby / Jellyfin 扫描媒体库

## 快速开始

### Docker Compose（推荐）

```bash
mkdir openlist-strm && cd openlist-strm

# 下载示例配置
wget https://raw.githubusercontent.com/konghanghang/openlist-strm/master/configs/config.example.yaml -O config.yaml

# 编辑配置（填入 Alist URL 和 Token）
vim config.yaml
```

创建 `docker-compose.yml`：

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

```bash
docker-compose up -d
```

访问 `http://localhost:8080` 进入 Web UI。

### 从源码构建

```bash
git clone https://github.com/konghang/openlist-strm.git
cd openlist-strm

# 构建前端
cd web && npm install && npm run build && cd ..

# 构建后端（前端资源自动嵌入）
make build

# 运行
./bin/openlist-strm -config config.yaml
```

## 最小配置

```yaml
server:
  host: "0.0.0.0"
  port: 8080

alist:
  url: "http://your-alist-url:5244"
  token: "your-alist-token"

database:
  path: "./data/openlist-strm.db"
```

路径映射通过 Web UI 管理，不在配置文件中设置。完整配置参考 [config.example.yaml](./configs/config.example.yaml)。

## STRM 模式

| 模式 | STRM 内容 | 适用场景 |
|------|-----------|---------|
| `alist_path` | Alist 路径，如 `/media/movies/movie.mp4` | 配合 [MediaWarp](https://github.com/AkimioJR/MediaWarp) 302 重定向，推荐 |
| `http_url` | 完整 URL，如 `http://alist.example.com/d/...` | 直接播放，简单场景 |

## API

```bash
# 生成 STRM（增量）
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"mode": "incremental"}'

# 查询任务
curl http://localhost:8080/api/tasks

# Webhook 触发
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{"path": "/aliyun/movies/new-movie.mp4", "event": "file.upload"}'
```

如配置了 API Token，请求需携带 `Authorization: Bearer <token>` 头。

详细 API 和 Webhook 用法见 [Webhook 集成指南](./deployments/WEBHOOK.md)。

## 文档

| 文档 | 说明 |
|------|------|
| [配置文件示例](./configs/config.example.yaml) | 完整配置项说明 |
| [Docker 部署指南](./deployments/README.md) | Docker 部署、Alist 集成、故障排查 |
| [Webhook 集成指南](./deployments/WEBHOOK.md) | Webhook API、下载器集成、自动化 |
| [项目需求文档](./docs/PRD.md) | 架构设计、功能规划 |
| [测试计划](./docs/TESTING.md) | 测试策略、进度跟踪 |

## 推荐配套工具

- [MediaWarp](https://github.com/AkimioJR/MediaWarp) — Emby/Jellyfin STRM 302 重定向代理
- [Strm Assistant](https://github.com/sjtuross/StrmAssistant) — Emby STRM 优化插件
- [ChineseSubFinder](https://github.com/ChineseSubFinder/ChineseSubFinder) — 中文字幕自动下载

## 贡献

欢迎提交 Issue 和 Pull Request。

## 许可证

MIT License

## 致谢

- [Alist](https://alist.nn.ci/) — 文件列表程序
- [tefuirZ/alist-strm](https://github.com/tefuirZ/alist-strm) — 项目灵感来源
