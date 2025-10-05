# Webhook 集成指南

OpenList-STRM 提供 Webhook 接口，支持外部系统通知后自动触发 STRM 生成。

## Webhook 接口

### 接口地址

```
POST /api/webhook
```

### 请求格式

```json
{
  "event": "file.upload",           // 事件类型
  "path": "/media/movies/new.mp4",  // 文件路径
  "action": "add"                    // 操作类型
}
```

### 响应格式

```json
{
  "success": true,
  "message": "webhook received, generation triggered",
  "task_id": "uuid-string"
}
```

## 集成示例

### 1. Alist Webhook

Alist 目前不直接支持 Webhook，但可以通过以下方式实现：

#### 方案 A: 定时任务（推荐）

在 OpenList-STRM 配置文件中设置定时任务：

```yaml
schedule:
  enabled: true
  cron: "0 */2 * * *"  # 每 2 小时检查一次
```

#### 方案 B: 脚本监听

编写监听脚本，检测到文件变化后调用 Webhook：

```bash
#!/bin/bash
# watch-and-notify.sh

# 监听目录
WATCH_DIR="/path/to/media"

# Webhook URL
WEBHOOK_URL="http://localhost:8080/api/webhook"

# 使用 inotifywait 监听文件变化
inotifywait -m -r -e create,delete,move "$WATCH_DIR" |
while read path action file; do
    echo "Detected: $action on $path$file"

    # 发送 Webhook 通知
    curl -X POST "$WEBHOOK_URL" \
      -H "Content-Type: application/json" \
      -d "{
        \"event\": \"file.$action\",
        \"path\": \"$path$file\",
        \"action\": \"$action\"
      }"
done
```

### 2. 下载器集成

#### qBittorrent

在 qBittorrent 设置中添加下载完成脚本：

**Linux/macOS:**
```bash
#!/bin/bash
# qb-webhook.sh

WEBHOOK_URL="http://localhost:8080/api/webhook"
FILE_PATH="$1"

curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"event\": \"download.completed\",
    \"path\": \"$FILE_PATH\",
    \"action\": \"add\"
  }"
```

**配置方法**：
1. 工具 → 选项 → 下载
2. "Torrent 完成时运行外部程序"
3. 输入：`/path/to/qb-webhook.sh "%F"`

#### Transmission

编辑 `settings.json`：

```json
{
  "script-torrent-done-enabled": true,
  "script-torrent-done-filename": "/path/to/transmission-webhook.sh"
}
```

脚本内容：
```bash
#!/bin/bash
# transmission-webhook.sh

WEBHOOK_URL="http://localhost:8080/api/webhook"
FILE_PATH="$TR_TORRENT_DIR/$TR_TORRENT_NAME"

curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"event\": \"download.completed\",
    \"path\": \"$FILE_PATH\",
    \"action\": \"add\"
  }"
```

### 3. 自动化工具集成

#### n8n

创建 Workflow：

1. **Trigger**: Webhook / File Watcher
2. **HTTP Request**: 调用 OpenList-STRM Webhook

```json
{
  "method": "POST",
  "url": "http://localhost:8080/api/webhook",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "event": "file.upload",
    "path": "{{$json['path']}}",
    "action": "add"
  }
}
```

#### Home Assistant

创建 Automation：

```yaml
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

rest_command:
  openlist_strm_webhook:
    url: "http://localhost:8080/api/webhook"
    method: POST
    content_type: "application/json"
    payload: >
      {
        "event": "file.upload",
        "path": "{{ path }}",
        "action": "add"
      }
```

## 认证

如果在配置文件中设置了 API Token：

```yaml
api:
  token: "your-secret-token"
```

调用 Webhook 时需要添加认证头：

```bash
curl -X POST http://localhost:8080/api/webhook \
  -H "Authorization: Bearer your-secret-token" \
  -H "Content-Type: application/json" \
  -d '{
    "event": "file.upload",
    "path": "/media/movies/new.mp4",
    "action": "add"
  }'
```

## 工作原理

1. 外部系统发送 Webhook 通知到 OpenList-STRM
2. OpenList-STRM 根据 `path` 匹配对应的映射配置
3. 如果匹配成功，在后台异步触发 STRM 生成任务
4. 立即返回响应，不阻塞外部系统

## 注意事项

### 频率限制

- Webhook 没有内置频率限制
- 建议外部系统自行控制调用频率
- 避免短时间内大量调用导致系统负载过高

### 路径匹配

- Webhook 会自动匹配配置文件中的映射
- 只有匹配到 `enabled: true` 的映射才会触发
- 路径需要与配置中的 `source` 路径前缀匹配

### 异步执行

- Webhook 接口立即返回，任务在后台执行
- 通过返回的 `task_id` 查询任务状态
- 使用 `/api/tasks/{task_id}` 获取任务详情

## 测试

使用 curl 测试 Webhook：

```bash
# 测试基本功能
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "file.upload",
    "path": "/media/movies/test.mp4",
    "action": "add"
  }'

# 测试无效路径
curl -X POST http://localhost:8080/api/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "file.upload",
    "path": "/invalid/path/test.mp4",
    "action": "add"
  }'
```

## 故障排查

### Webhook 调用失败

- 检查网络连接
- 确认 OpenList-STRM 服务正在运行
- 查看日志文件：`./logs/openlist-strm.log`

### 未触发生成

- 确认路径与映射配置匹配
- 检查映射是否启用（`enabled: true`）
- 查看任务列表确认是否创建任务

### 认证失败

- 确认 API Token 配置正确
- 检查 Authorization 头格式：`Bearer {token}`
