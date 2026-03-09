# Webhook 集成指南

OpenList-STRM 提供 Webhook 接口，接收外部系统通知后自动触发 STRM 生成。

## 接口

```
POST /api/webhook
Content-Type: application/json
```

### 请求参数

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `path` | string | 是 | 文件或目录路径 |
| `event` | string | 否 | 事件类型（用于日志记录） |
| `config_name` | string | 否 | 指定配置名称（优先使用，跳过路径匹配） |
| `mode` | string | 否 | 执行模式：`incremental` 或 `full`（覆盖配置默认值） |
| `source` | string | 否 | 来源标识（用于日志记录） |
| `drive_path` | string | 否 | 网盘路径前缀（用于路径映射） |
| `alist_path` | string | 否 | Alist 路径前缀（用于路径映射） |

### 响应

成功触发：
```json
{"success": true, "message": "webhook received, generation triggered", "task_id": "uuid-string"}
```

未匹配到配置（跳过）：
```json
{"success": true, "skipped": true, "message": "no matching mapping found"}
```

错误：
```json
{"success": false, "message": "config not found: Movies"}
```

### 认证

如配置了 API Token，需携带认证头：

```bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{"path": "/aliyun/movies/new.mp4", "event": "file.upload"}'
```

## 使用场景

### 1. 基本用法（路径自动匹配）

```bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/aliyun/movies/new-movie.mp4",
    "event": "file.upload",
    "source": "alist"
  }'
```

### 2. 指定配置名称

跳过路径匹配，直接触发指定配置：

```bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/aliyun/movies/new-movie.mp4",
    "config_name": "Movies",
    "mode": "incremental"
  }'
```

### 3. 网盘路径映射

当 Webhook 通知的路径是网盘原始路径而非 Alist 挂载路径时，使用 `drive_path` 和 `alist_path` 进行转换。

场景：
```
网盘实际路径:  /我的资源/影视/电影
Alist 挂载路径: /aliyun/movies
```

```bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/我的资源/影视/电影/斗罗大陆/S01E01.mp4",
    "event": "file.upload",
    "drive_path": "/我的资源/影视/电影",
    "alist_path": "/aliyun/movies"
  }'
```

转换过程：
1. 原始路径：`/我的资源/影视/电影/斗罗大陆/S01E01.mp4`
2. 去掉 `drive_path` 前缀：`/斗罗大陆/S01E01.mp4`
3. 拼接 `alist_path`：`/aliyun/movies/斗罗大陆/S01E01.mp4`
4. 用转换后的路径匹配配置

### 4. 强制全量扫描

```bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/aliyun/movies",
    "mode": "full",
    "event": "manual.trigger"
  }'
```

## 集成示例

### qBittorrent

设置 → 下载 → "Torrent 完成时运行外部程序"：

```
/path/to/qb-webhook.sh "%F"
```

脚本 `qb-webhook.sh`：

```bash
#!/bin/bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d "{
    \"event\": \"download.completed\",
    \"path\": \"$1\",
    \"source\": \"qbittorrent\"
  }"
```

### Transmission

编辑 `settings.json`：

```json
{
  "script-torrent-done-enabled": true,
  "script-torrent-done-filename": "/path/to/transmission-webhook.sh"
}
```

脚本 `transmission-webhook.sh`：

```bash
#!/bin/bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d "{
    \"event\": \"download.completed\",
    \"path\": \"$TR_TORRENT_DIR/$TR_TORRENT_NAME\",
    \"source\": \"transmission\"
  }"
```

### n8n

HTTP Request 节点配置：

```json
{
  "method": "POST",
  "url": "http://localhost:8080/api/webhook",
  "headers": {"Content-Type": "application/json"},
  "body": {
    "event": "file.upload",
    "path": "{{$json['path']}}",
    "source": "n8n"
  }
}
```

### Home Assistant

```yaml
rest_command:
  openlist_strm_webhook:
    url: "http://localhost:8080/api/webhook"
    method: POST
    content_type: "application/json"
    payload: '{"event": "file.upload", "path": "{{ path }}", "source": "homeassistant"}'

automation:
  - alias: "Notify OpenList-STRM on file upload"
    trigger:
      - platform: event
        event_type: folder_watcher
        event_data:
          event_type: created
    action:
      - service: rest_command.openlist_strm_webhook
        data:
          path: "{{ trigger.event.data.path }}"
```

## 工作原理

1. 外部系统发送 Webhook 通知
2. 如指定 `config_name`，直接使用该配置；否则根据 `path` 匹配配置的 `source` 路径
3. 如提供 `drive_path` + `alist_path`，先进行路径转换再匹配
4. 匹配成功后在后台异步触发 STRM 生成任务
5. 立即返回响应（含 `task_id`），不阻塞调用方

## 故障排查

**未触发生成**：确认路径与配置的 `source` 匹配，且配置已启用（`enabled: true`）。

**认证失败**：确认 `Authorization: Bearer <token>` 格式正确，Token 与配置一致。

**查看任务状态**：通过返回的 `task_id` 调用 `GET /api/tasks/{task_id}` 查询。
