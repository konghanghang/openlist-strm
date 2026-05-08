# OpenList-STRM

English | [中文](./README.md)

Batch generate STRM files from media files in Alist cloud storage for direct playback in Emby / Jellyfin / Plex. Built with Go, single binary deployment, Web UI management.

## Features

- **No Mount Required** — Play media directly via STRM files, no cloud drive mounting needed
- **High Performance** — Go concurrency with per-config concurrency control (rate limit friendly)
- **Incremental Sync** — Smart incremental updates, only processes changed files
- **Independent Scheduling** — Per-config Cron jobs with visual editor
- **Dual Mode** — Supports `alist_path` (with MediaWarp) and `http_url` (direct link)
- **Web UI** — Modern Vue 3 interface with dashboard, config management, and task monitoring
- **Webhook** — Auto-trigger via external notifications with path mapping support
- **Media Notifications** — Auto-notify Emby / Jellyfin to scan library after task completion

## Quick Start

### Docker Compose (Recommended)

```bash
mkdir openlist-strm && cd openlist-strm

# Download example config
wget https://raw.githubusercontent.com/konghanghang/openlist-strm/master/configs/config.example.yaml -O config.yaml

# Edit config (fill in your Alist URL and Token)
vim config.yaml
```

Create `docker-compose.yml`:

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

Open `http://localhost:8080` to access the Web UI.

### Build from Source

```bash
git clone https://github.com/konghang/openlist-strm.git
cd openlist-strm

# Build frontend
cd web && npm install && npm run build && cd ..

# Build backend (frontend assets embedded automatically)
make build

# Run
./bin/openlist-strm
```

The backend now searches for `config.yaml` / `config.yml` in this order when `-config` is not provided:

1. The current working directory
2. `configs/` under the current working directory
3. The parent directory of the current working directory and its `configs/`
4. The executable directory and its parent directory, plus their `configs/`
5. Docker default path `/app/configs/`
6. System path `/etc/openlist-strm/`

You can still override this behavior with `-config /path/to/config.yaml`.

## Minimal Configuration

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

Path mappings are managed via the Web UI, not in the config file. See [config.example.yaml](./configs/config.example.yaml) for the full configuration reference.

## STRM Modes

| Mode | STRM Content | Use Case |
|------|-------------|----------|
| `alist_path` | Alist path, e.g. `/media/movies/movie.mp4` | With [MediaWarp](https://github.com/AkimioJR/MediaWarp) 302 redirect, recommended |
| `http_url` | Full URL, e.g. `http://alist.example.com/d/...` | Direct playback, simple setup |

## API

```bash
# Generate STRM (incremental)
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{"mode": "incremental"}'

# List tasks
curl http://localhost:8080/api/tasks

# Webhook trigger
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{"path": "/aliyun/movies/new-movie.mp4", "event": "file.upload"}'
```

If API Token is configured, requests must include the `Authorization: Bearer <token>` header.

For detailed API and Webhook usage, see the [Webhook Integration Guide](./deployments/WEBHOOK.md).

## Documentation

| Document | Description |
|----------|-------------|
| [Configuration Example](./configs/config.example.yaml) | Full configuration reference |
| [Docker Deployment Guide](./deployments/README.md) | Docker deployment, Alist integration, troubleshooting |
| [Webhook Integration Guide](./deployments/WEBHOOK.md) | Webhook API, downloader integration, automation |
| [Product Requirements](./docs/PRD.md) | Architecture design, feature planning |
| [Testing Plan](./docs/TESTING.md) | Testing strategy, progress tracking |

## Recommended Tools

- [MediaWarp](https://github.com/AkimioJR/MediaWarp) — Emby/Jellyfin STRM 302 redirect proxy
- [Strm Assistant](https://github.com/sjtuross/StrmAssistant) — Emby STRM optimization plugin
- [ChineseSubFinder](https://github.com/ChineseSubFinder/ChineseSubFinder) — Auto Chinese subtitle downloader

## Contributing

Issues and Pull Requests are welcome.

## License

[MIT License](./LICENSE)

## Acknowledgements

- [Alist](https://alist.nn.ci/) — File listing program
- [tefuirZ/alist-strm](https://github.com/tefuirZ/alist-strm) — Project inspiration
