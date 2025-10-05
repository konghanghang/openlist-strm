# OpenList-STRM

OpenList-STRM 是一个基于 Go 语言开发的 STRM 文件生成工具，用于将 Alist 网盘中的媒体文件批量生成为 STRM 格式，供 Emby、Jellyfin、Plex 等流媒体服务器使用。

## ✨ 特性

- 🚀 **高性能**：Go 语言实现，并发处理，性能优于 Python 实现
- 💾 **免挂载**：无需挂载网盘，通过 STRM 文件直接播放
- 📦 **节省空间**：本地只存储小体积的 STRM 文件
- ⏰ **自动同步**：支持定时任务，自动更新媒体库
- 🔄 **增量更新**：智能增量同步，只处理新增和修改的文件
- 🎯 **简单易用**：单二进制文件部署，配置简单
- 🌐 **Web UI**：现代化 Vue 3 界面，可视化管理
- 🔌 **API 接口**：RESTful API，支持外部程序调用

## 📋 当前版本

**v1.0.0** - 完整功能版本

已实现功能：
- ✅ Alist API 集成
- ✅ STRM 文件生成
- ✅ 增量/全量更新模式
- ✅ 定时任务调度（Cron）
- ✅ SQLite 数据存储
- ✅ 并发处理
- ✅ 日志系统
- ✅ **RESTful API 接口**
- ✅ **Vue 3 Web UI 管理界面**
- ✅ **Webhook 支持**
- ✅ **Docker 部署**

待实现功能（后续版本）：
- ⏳ 元数据下载
- ⏳ 文件有效性检测
- ⏳ UI 优化和完善

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
alist:
  url: "http://your-alist-url:5244"
  token: "your-alist-token"

mappings:
  - name: "Movies"
    source: "/media/movies"  # Alist 中的路径
    target: "/mnt/strm/movies"  # 本地 STRM 路径
    mode: "incremental"
    enabled: true
```

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
- ⚙️ **配置管理**：查看路径映射、手动触发生成

## 📖 配置说明

### Alist 配置

```yaml
alist:
  url: "http://localhost:5244"  # Alist 服务地址
  token: "your-alist-token"     # Alist API Token
  sign_enabled: false           # 是否启用签名
  timeout: 30                   # 请求超时（秒）
```

### 路径映射

```yaml
mappings:
  - name: "Movies"               # 映射名称
    source: "/media/movies"      # Alist 源路径
    target: "/mnt/strm/movies"   # STRM 目标路径
    mode: "incremental"          # 更新模式：incremental 或 full
    enabled: true                # 是否启用
```

### 定时任务

```yaml
schedule:
  enabled: true
  cron: "0 2 * * *"  # 每天凌晨 2 点执行
```

Cron 表达式说明：
- `0 2 * * *` - 每天凌晨 2 点
- `0 */6 * * *` - 每 6 小时
- `0 0 * * 0` - 每周日凌晨

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

## 🐳 Docker 部署

### 使用 Docker Compose（推荐）

```bash
# 克隆仓库
git clone https://github.com/konghang/openlist-strm.git
cd openlist-strm

# 复制配置文件
cp configs/config.example.yaml config.yaml
vim config.yaml  # 编辑配置

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

### 使用 Docker

```bash
# 构建镜像
docker build -t openlist-strm:latest .

# 运行容器
docker run -d \
  --name openlist-strm \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v /path/to/strm:/mnt/strm \
  openlist-strm:latest
```

详细部署文档请查看：[deployments/README.md](./deployments/README.md)

## 📦 推荐配套工具

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
- [项目需求文档 (PRD)](./PRD.md)
- [配置文件示例](./configs/config.example.yaml)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

- [Alist](https://alist.nn.ci/) - 优秀的文件列表程序
- [tefuirZ/alist-strm](https://github.com/tefuirZ/alist-strm) - 项目灵感来源

---

**🤖 Powered by Go | Made with ❤️**
