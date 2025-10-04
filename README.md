# OpenList-STRM

OpenList-STRM 是一个基于 Go 语言开发的 STRM 文件生成工具，用于将 Alist 网盘中的媒体文件批量生成为 STRM 格式，供 Emby、Jellyfin、Plex 等流媒体服务器使用。

## ✨ 特性

- 🚀 **高性能**：Go 语言实现，并发处理，性能优于 Python 实现
- 💾 **免挂载**：无需挂载网盘，通过 STRM 文件直接播放
- 📦 **节省空间**：本地只存储小体积的 STRM 文件
- ⏰ **自动同步**：支持定时任务，自动更新媒体库
- 🔄 **增量更新**：智能增量同步，只处理新增和修改的文件
- 🎯 **简单易用**：单二进制文件部署，配置简单

## 📋 当前版本

**v1.0.0-MVP** - 核心功能实现

已实现功能：
- ✅ Alist API 集成
- ✅ STRM 文件生成
- ✅ 增量/全量更新模式
- ✅ 定时任务调度（Cron）
- ✅ SQLite 数据存储
- ✅ 并发处理
- ✅ 日志系统

待实现功能（后续版本）：
- ⏳ Web UI 管理界面
- ⏳ RESTful API 接口
- ⏳ 元数据下载
- ⏳ Docker 部署

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
#./bin/openlist-strm --config config.yaml
```

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

## 🛠️ 构建

```bash
# 安装依赖
go mod download

# 编译
make build

# 或者手动编译
go build -o bin/openlist-strm ./cmd/server
```

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
