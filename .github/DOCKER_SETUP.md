# Docker Hub 配置说明

本项目的 Docker 镜像会同时推送到两个平台：
- **GitHub Container Registry (GHCR)**: `ghcr.io/konghanghang/openlist-strm`
- **Docker Hub**: `你的用户名/openlist-strm`

## 配置步骤

### 1. 创建 Docker Hub 账号

访问 [Docker Hub](https://hub.docker.com/) 注册账号。

### 2. 创建仓库（可选）

Docker Hub 会在首次推送时自动创建仓库，但你也可以提前手动创建：
1. 登录 Docker Hub
2. 点击 "Create Repository"
3. 仓库名称设置为：`openlist-strm`
4. 选择 Public（公开）或 Private（私有）

### 3. 生成 Access Token

1. 登录 Docker Hub
2. 点击右上角头像 > **Account Settings**
3. 左侧菜单选择 **Security**
4. 点击 **New Access Token**
5. 填写描述（如 `GitHub Actions`）
6. 权限选择 **Read, Write, Delete**
7. 点击 **Generate**
8. **重要**：立即复制生成的 Token，关闭后将无法再次查看

### 4. 配置 GitHub Secrets

1. 打开 GitHub 仓库
2. 进入 **Settings** > **Secrets and variables** > **Actions**
3. 点击 **New repository secret**
4. 添加以下两个 secrets：

   **Secret 1:**
   - Name: `DOCKER_USERNAME`
   - Value: 你的 Docker Hub 用户名

   **Secret 2:**
   - Name: `DOCKER_TOKEN`
   - Value: 刚才生成的 Access Token

### 5. 验证配置

配置完成后：
1. 提交代码到主分支，触发 Docker Build 工作流
2. 或者打 tag 触发 Release 工作流
3. 在 Actions 页面查看工作流运行状态
4. 成功后可在 Docker Hub 仓库页面看到推送的镜像

## 镜像拉取命令

### 从 Docker Hub 拉取（推荐国内用户）

```bash
# 拉取最新版本
docker pull 你的用户名/openlist-strm:latest

# 拉取特定版本
docker pull 你的用户名/openlist-strm:v1.0.0

# 拉取主分支构建
docker pull 你的用户名/openlist-strm:master
```

### 从 GHCR 拉取

```bash
# 拉取最新版本
docker pull ghcr.io/konghanghang/openlist-strm:latest

# 拉取特定版本
docker pull ghcr.io/konghanghang/openlist-strm:v1.0.0
```

## 注意事项

1. **Token 安全**：
   - 不要将 Token 提交到代码仓库
   - Token 泄露后立即在 Docker Hub 撤销并重新生成

2. **权限设置**：
   - GitHub Actions 需要 `DOCKER_USERNAME` 和 `DOCKER_TOKEN` 两个 secrets
   - 确保 Token 有 Write 权限

3. **仓库命名**：
   - 工作流会自动使用 GitHub 仓库名作为 Docker 镜像名
   - 如需自定义，可修改 `.github/workflows/release.yml` 中的 tags

4. **多架构支持**：
   - 镜像支持 `linux/amd64` 和 `linux/arm64`
   - Docker 会自动选择匹配系统架构的镜像

## 故障排查

### 工作流失败：authentication required

**原因**：未配置 Docker Hub secrets 或配置错误

**解决**：
1. 检查 GitHub Secrets 中是否存在 `DOCKER_USERNAME` 和 `DOCKER_TOKEN`
2. 确认 Token 未过期且有效
3. 重新生成 Token 并更新 Secret

### 镜像推送失败：denied

**原因**：Token 权限不足

**解决**：
1. 重新生成 Token，确保选择 **Read, Write, Delete** 权限
2. 更新 GitHub Secret

### 国内拉取慢

**解决方案**：
1. 使用 Docker Hub 镜像（比 GHCR 快）
2. 配置 Docker 镜像加速器：
   ```bash
   # 编辑 /etc/docker/daemon.json
   {
     "registry-mirrors": [
       "https://docker.mirrors.ustc.edu.cn"
     ]
   }
   ```
3. 重启 Docker：`sudo systemctl restart docker`
