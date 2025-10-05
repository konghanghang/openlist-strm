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
# 构建镜像
docker build -t openlist-strm:latest ..

# 运行容器
docker run -d \
  --name openlist-strm \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/configs/config.yaml:ro \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/logs:/app/logs \
  -v /path/to/your/strm:/mnt/strm \
  -e TZ=Asia/Shanghai \
  openlist-strm:latest
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
| `./logs` | `/app/logs` | 日志文件 |
| `/your/strm/path` | `/mnt/strm` | STRM 输出目录 |

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

- `./data` - 数据库文件
- `./logs` - 日志文件
- `/your/strm/path` - STRM 文件输出

## 健康检查

容器内置健康检查，每 30 秒检查一次 `/api/health` 端点。

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
