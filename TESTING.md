# 测试计划

本文档记录 OpenList-STRM 项目的测试计划和进度。

## 测试策略

### 测试类型
- **单元测试**: 测试独立的函数和方法
- **集成测试**: 测试模块间的交互
- **HTTP 测试**: 使用 `httptest` 测试 API 端点

### 测试覆盖目标
- 核心业务逻辑: 80%+
- API 处理器: 70%+
- 工具函数: 60%+

## 测试清单

### 高优先级 - 核心业务逻辑

#### 1. Config 模块 (`internal/config`) ✅
**文件**: `internal/config/loader_test.go`

- [x] 测试配置文件加载
  - [x] 加载有效的 YAML 配置
  - [x] 处理文件不存在的情况
  - [x] 处理无效的 YAML 格式
  - [x] 处理缺失必填字段
- [x] 测试配置验证
  - [x] 验证必填字段（Alist URL, Token）
  - [x] 验证端口范围
  - [x] 验证路径映射配置
  - [x] 验证 mapping mode
- [x] 测试默认值
  - [x] Server 默认配置
  - [x] STRM 默认并发数
  - [x] 日志默认级别
- [x] 测试 GetAddr() 方法

**测试结果**: ✅ 全部通过 (7 tests, 13 subtests)
**实际时间**: 已完成

---

#### 2. STRM Generator 模块 (`internal/strm`) ✅
**文件**: `internal/strm/generator_test.go`

- [x] 测试 STRM 文件生成
  - [x] 生成单个 STRM 文件
  - [x] 生成多个 STRM 文件（全量模式）
  - [x] 保持目录结构
  - [x] 处理中文文件名
  - [x] 增量模式（跳过已存在文件）
- [x] 测试并发生成
  - [x] 并发数限制
  - [x] 上下文取消
- [x] 测试错误处理
  - [x] Alist API 错误
- [x] 测试工具函数
  - [x] changeExtension
  - [x] cleanDirectory

**代码改进**:
- 引入 `AlistClient` 接口支持 mock 测试
- 修改 `generateSTRMFile` 返回 (created bool, error) 以正确统计跳过的文件

**测试结果**: ✅ 全部通过 (10 tests, 4 subtests)
**实际时间**: 已完成

---

#### 3. Alist Client 模块 (`internal/alist`)
**文件**: `internal/alist/client_test.go`

- [ ] 测试 Ping 方法
  - [ ] 成功连接
  - [ ] 连接超时
  - [ ] 认证失败
- [ ] 测试 ListFilesRecursive 方法
  - [ ] 列出单层目录
  - [ ] 递归列出子目录
  - [ ] 过滤文件扩展名
  - [ ] 处理空目录
  - [ ] 处理 API 错误响应
- [ ] 测试 GetFileURL 方法
  - [ ] 生成直链 URL
  - [ ] 处理签名（如果启用）
  - [ ] URL 编码
- [ ] 测试错误处理和重试
  - [ ] HTTP 错误响应
  - [ ] 网络超时
  - [ ] JSON 解析错误

**预计时间**: 60-90 分钟

---

#### 4. Storage 模块 (`internal/storage`)
**文件**: `internal/storage/sqlite_test.go`

- [ ] 测试数据库初始化
  - [ ] 创建新数据库
  - [ ] 自动迁移表结构
  - [ ] 连接已存在的数据库
- [ ] 测试 Task CRUD
  - [ ] CreateTask
  - [ ] GetTaskByID
  - [ ] ListTasks (分页)
  - [ ] UpdateTaskStatus
- [ ] 测试 File CRUD
  - [ ] SaveFile
  - [ ] GetFileByPath
  - [ ] ListFilesByConfig
  - [ ] DeleteFile
- [ ] 测试事务和并发
  - [ ] 并发写入
  - [ ] 事务回滚
- [ ] 测试 Close 方法

**预计时间**: 60-75 分钟

---

### 中优先级 - 业务逻辑

#### 5. API Handlers 模块 (`internal/api`)
**文件**: `internal/api/handlers_test.go`

- [ ] 测试 handleHealth
  - [ ] 返回健康状态
- [ ] 测试 handleGenerate
  - [ ] 生成所有映射
  - [ ] 生成指定路径
  - [ ] 验证 mode 参数
  - [ ] 参数验证
- [ ] 测试 handleGetTask
  - [ ] 获取存在的任务
  - [ ] 任务不存在返回 404
- [ ] 测试 handleListTasks
  - [ ] 返回任务列表
  - [ ] 分页功能
- [ ] 测试 handleGetConfigs
  - [ ] 返回配置列表
- [ ] 测试 handleWebhook
  - [ ] 有效的 webhook 请求
  - [ ] 路径匹配
  - [ ] 路径不匹配
  - [ ] 缺少必填字段

**预计时间**: 60-90 分钟

---

#### 6. API Middleware 模块 (`internal/api`)
**文件**: `internal/api/middleware_test.go`

- [ ] 测试 tokenAuthMiddleware
  - [ ] 有效 Token
  - [ ] 无效 Token
  - [ ] 缺少 Token
  - [ ] Token 格式错误
  - [ ] Bearer 前缀验证

**预计时间**: 20-30 分钟

---

#### 7. Scheduler 模块 (`internal/scheduler`)
**文件**: `internal/scheduler/scheduler_test.go`

- [ ] 测试 New 创建调度器
- [ ] 测试 Start/Stop
  - [ ] 启动定时任务
  - [ ] 停止定时任务
- [ ] 测试 RunMapping
  - [ ] 执行单个映射
  - [ ] 增量模式
  - [ ] 全量模式
  - [ ] 记录任务状态
- [ ] 测试 RunAll
  - [ ] 执行所有启用的映射
  - [ ] 跳过禁用的映射
- [ ] 测试 RunMappingByName
  - [ ] 按名称查找并执行
  - [ ] 名称不存在的情况

**预计时间**: 45-60 分钟

---

### 低优先级 - 辅助功能

#### 8. Logger 模块 (`internal/logger`)
**文件**: `internal/logger/logger_test.go`

- [ ] 测试 Init
  - [ ] 初始化日志配置
  - [ ] 创建日志文件
  - [ ] 设置日志级别
- [ ] 测试日志输出
  - [ ] Info 级别
  - [ ] Error 级别
  - [ ] Debug 级别
- [ ] 测试日志轮转
- [ ] 测试 Close

**预计时间**: 20-30 分钟

---

## 测试工具和依赖

### 标准库
- `testing` - Go 标准测试框架
- `net/http/httptest` - HTTP 测试
- `io/ioutil` - 临时文件/目录

### 第三方库（可选）
- `github.com/stretchr/testify` - 断言库（推荐）
- `github.com/DATA-DOG/go-sqlmock` - SQL Mock（可选）

## 测试命令

```bash
# 运行所有测试
go test ./...

# 运行指定包的测试
go test ./internal/config

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 运行测试（详细输出）
go test -v ./...

# 运行指定的测试函数
go test -v -run TestLoadConfig ./internal/config
```

## 进度跟踪

- [x] Config 模块 (4/4) ✅
- [x] STRM Generator 模块 (4/4) ✅
- [ ] Alist Client 模块 (0/4)
- [ ] Storage 模块 (0/5)
- [ ] API Handlers 模块 (0/7)
- [ ] API Middleware 模块 (0/1)
- [ ] Scheduler 模块 (0/5)
- [ ] Logger 模块 (0/4)

**总体进度**: 8/34 (24%)
**核心模块覆盖**: 2/4 (50%)

---

## 测试编写规范

### 文件命名
- 测试文件名: `{package}_test.go`
- 与被测试文件同目录

### 测试函数命名
```go
func TestFunctionName(t *testing.T)           // 单个测试
func TestFunctionName_Scenario(t *testing.T)  // 特定场景
```

### 表驱动测试
```go
func TestExample(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case1", "input1", "output1", false},
        {"case2", "input2", "output2", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Mock 和 Stub
- 使用接口抽象外部依赖
- 创建 mock 实现用于测试
- 避免真实的网络/文件系统操作

---

最后更新: 2025-10-05
