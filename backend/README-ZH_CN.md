# AwesomeSharing 后端

[English Documentation](./README.md)

AwesomeSharing 是一个基于 Go 的照片和文件共享系统后端服务，支持文件管理、相册、分享链接、权限管理等功能。

## 技术栈

- **语言**: Go 1.24.0
- **Web 框架**: Fiber v2
- **数据库**: SQLite
- **图片处理**: imaging, goexif
- **加密**: golang.org/x/crypto

## 项目结构

```
backend/
├── cmd/
│   └── server/          # 服务器入口
│       └── main.go      # 主程序
├── internal/
│   ├── api/             # API 处理器和路由
│   │   ├── routes_v2.go         # V2 版本路由（带认证）
│   │   ├── auth_handlers.go     # 认证相关处理器
│   │   ├── user_handlers.go     # 用户管理处理器
│   │   ├── folder_handlers.go   # 文件夹处理器
│   │   ├── album_handlers.go    # 相册处理器
│   │   ├── share_handlers.go    # 分享处理器
│   │   └── ...
│   ├── services/        # 业务逻辑服务
│   │   ├── auth.go              # 认证服务
│   │   ├── folder.go            # 文件夹服务
│   │   ├── album.go             # 相册服务
│   │   ├── share.go             # 分享服务
│   │   ├── scanner.go           # 文件扫描服务
│   │   ├── thumbnail.go         # 缩略图服务
│   │   └── ...
│   ├── middleware/      # 中间件
│   │   └── auth.go              # 认证中间件
│   ├── models/          # 数据模型
│   │   └── models.go
│   ├── database/        # 数据库初始化和迁移
│   │   ├── database.go
│   │   └── schema_v3.go
│   ├── config/          # 配置管理
│   │   └── config.go
│   └── initialization/  # 初始化逻辑
│       └── init.go
├── pkg/                 # 可重用的包
│   ├── utils/
│   └── exif/
├── config/              # 配置目录（运行时生成）
├── upload/              # 上传目录（运行时生成）
├── go.mod               # Go 模块依赖
└── run-local-v2.sh      # 本地启动脚本

```

## 快速开始

### 前置要求

- Go 1.24.0 或更高版本
- Git

### 安装依赖

```bash
cd backend
go mod download
```

### 本地启动

#### 方式 1: 使用启动脚本（推荐）

```bash
./run-local-v2.sh
```

#### 方式 2: 直接运行

```bash
# 设置环境变量
export CONFIG_DIR="./config"
export UPLOAD_DIR="./upload"
export PORT="8080"
export ALLOWED_ORIGIN="http://localhost:3000"

# 启动服务器
go run cmd/server/main.go
```

### 环境变量配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | `8080` | 服务器端口 |
| `CONFIG_DIR` | `/config` | 配置目录路径（存储数据库和缩略图） |
| `UPLOAD_DIR` | `/upload` | 上传目录路径 |
| `ALLOWED_ORIGIN` | `*` | CORS 允许的源（生产环境建议设置具体域名） |
| `DISABLE_FILE_VALIDATION` | `false` | 禁用文件验证（设置为 `true` 禁用） |

### 首次启动

服务器首次启动时会自动：

1. 创建 SQLite 数据库（`config/awesome-sharing.db`）
2. 初始化数据库表结构
3. 创建默认管理员账号：
   - 用户名: `admin`
   - 密码: `admin`
4. 启动后台文件扫描器（每 30 分钟扫描一次）
5. 启动文件验证服务（每 6 小时清理无效文件）

### 验证服务运行

访问健康检查端点：

```bash
curl http://localhost:8080/api/health
```

预期响应：

```json
{"status":"ok"}
```

## API 接口文档

### 公开接口（无需认证）

#### 健康检查

```
GET /api/health
```

#### 公开设置

```
GET /api/settings/public
```

#### 访问分享链接

```
GET /api/s/:id
```

### 认证接口

#### 登录

```
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin"
}
```

#### 注册（可选限制）

```
POST /api/auth/register
Content-Type: application/json

{
  "username": "newuser",
  "password": "password123",
  "email": "user@example.com"
}
```

#### 登出

```
POST /api/auth/logout
Authorization: Bearer <token>
```

#### 获取当前用户信息

```
GET /api/auth/me
Authorization: Bearer <token>
```

#### 修改密码

```
POST /api/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "oldpass",
  "new_password": "newpass"
}
```

### 用户管理接口（管理员）

```
GET    /api/users                      # 列出用户
GET    /api/users/search               # 搜索用户
GET    /api/users/stats                # 用户统计
POST   /api/users                      # 创建用户
GET    /api/users/:id                  # 获取用户详情
PUT    /api/users/:id                  # 更新用户
DELETE /api/users/:id                  # 删除用户
PUT    /api/users/:id/toggle           # 启用/禁用用户
POST   /api/users/:id/reset-password   # 重置密码
GET    /api/users/:id/activity-logs    # 用户活动日志
POST   /api/users/export               # 导出用户
POST   /api/users/bulk/enable-disable  # 批量启用/禁用
POST   /api/users/bulk/delete          # 批量删除
```

### 文件夹管理接口

```
GET    /api/folders                # 列出文件夹
POST   /api/folders                # 创建文件夹（管理员）
GET    /api/folders/:id            # 获取文件夹详情
PUT    /api/folders/:id            # 更新文件夹（管理员）
DELETE /api/folders/:id            # 删除文件夹（管理员）
PUT    /api/folders/:id/toggle     # 启用/禁用文件夹（管理员）
POST   /api/folders/:id/scan       # 扫描文件夹（管理员）
GET    /api/folders/:id/files      # 列出文件夹中的文件
```

### 权限组接口

```
GET    /api/permission-groups                        # 列出权限组
POST   /api/permission-groups                        # 创建权限组（管理员）
GET    /api/permission-groups/:id                    # 获取权限组详情
PUT    /api/permission-groups/:id                    # 更新权限组（管理员）
DELETE /api/permission-groups/:id                    # 删除权限组（管理员）
GET    /api/permission-groups/:id/folders            # 列出权限组中的文件夹
POST   /api/permission-groups/:id/folders            # 添加文件夹到权限组（管理员）
DELETE /api/permission-groups/:id/folders/:folderId  # 从权限组移除文件夹（管理员）
GET    /api/permission-groups/:id/permissions        # 列出权限
POST   /api/permission-groups/:id/permissions        # 授予权限（管理员）
DELETE /api/permission-groups/:id/permissions/:userId # 撤销权限（管理员）
```

### 相册接口（V2）

```
GET    /api/albums-v2                 # 列出相册
POST   /api/albums-v2                 # 创建相册
GET    /api/albums-v2/:id             # 获取相册详情
PUT    /api/albums-v2/:id             # 更新相册
DELETE /api/albums-v2/:id             # 删除相册
GET    /api/albums-v2/:id/items       # 列出相册项目
POST   /api/albums-v2/:id/items       # 添加项目到相册
DELETE /api/albums-v2/:id/items/:itemId # 从相册移除项目
POST   /api/albums-v2/:id/resolve     # 解析相册项目（管理员）
POST   /api/albums-v2/resolve-all     # 解析所有相册（管理员）
```

### 分享接口

```
GET    /api/shares                         # 列出分享
POST   /api/shares                         # 创建分享
GET    /api/shares/:id                     # 获取分享详情
PUT    /api/shares/:id                     # 更新分享
DELETE /api/shares/:id                     # 删除分享
POST   /api/shares/:id/extend              # 延长分享有效期
GET    /api/shares/:id/access-log          # 获取分享访问日志
POST   /api/shares/:id/permissions         # 授予分享权限（私有分享）
DELETE /api/shares/:id/permissions/:userId # 撤销分享权限
DELETE /api/shares/expired                 # 删除过期分享
```

### 文件接口（向后兼容）

```
GET /api/files                  # 获取文件列表
GET /api/files/:id              # 获取文件详情
GET /api/files/:id/thumbnail    # 获取文件缩略图
GET /api/files/:id/download     # 下载文件
GET /api/timeline               # 时间线视图
GET /api/search                 # 搜索文件
```

### 系统管理接口（管理员）

```
GET  /api/settings              # 获取系统设置
PUT  /api/settings              # 更新系统设置
GET  /api/settings/domain       # 获取域名配置
PUT  /api/settings/domain       # 更新域名配置
GET  /api/domain-config         # 获取域名配置
POST /api/domain-config         # 保存域名配置
POST /api/scan                  # 手动触发扫描
POST /api/cleanup               # 清理已删除的文件
```

### 其他接口

```
GET  /api/tags                  # 获取标签列表
POST /api/tags                  # 创建标签
GET  /api/mount-points          # 获取挂载点
```

## 测试

### 运行所有测试

```bash
go test ./...
```

### 运行特定包的测试

```bash
go test ./internal/services
go test ./internal/api
```

### 运行测试并显示覆盖率

```bash
go test -cover ./...
```

### 生成覆盖率报告

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 开发指南

### 代码规范

- 使用 `gofmt` 格式化代码
- 遵循 Go 标准库的命名约定
- 为导出的函数和类型添加注释

### 添加新的 API 端点

1. 在 `internal/api/` 中创建或更新 handler
2. 在 `internal/services/` 中实现业务逻辑
3. 在 `internal/api/routes_v2.go` 中注册路由
4. 如需数据库更改，更新 `internal/database/schema_v*.go`

### 数据库迁移

数据库迁移会在服务器启动时自动执行。要添加新的迁移：

1. 创建新的 schema 文件：`internal/database/schema_v<new_version>.go`
2. 在 `internal/database/database.go` 中注册新的迁移版本
3. 重启服务器以应用迁移

## 后台服务

服务器启动后会运行以下后台任务：

1. **文件扫描器**: 每 30 分钟扫描配置的文件夹以发现新文件
2. **文件验证器**: 每 6 小时验证数据库中的文件是否仍然存在，清理无效记录
3. **会话清理**: 每 1 小时清理过期的会话

可以通过环境变量 `DISABLE_FILE_VALIDATION=true` 禁用文件验证功能。

## 故障排查

### 数据库锁定错误

如果遇到 "database is locked" 错误：

1. 确保只有一个服务器实例在运行
2. 检查是否有其他程序正在访问数据库文件
3. 等待后台扫描任务完成

### 无法创建缩略图

确保：

1. `config/thumbs` 目录存在且可写
2. 图片文件格式受支持（JPEG, PNG, HEIC, TIFF 等）
3. 系统有足够的磁盘空间

### 权限问题

确保运行服务器的用户对以下目录有读写权限：

- `CONFIG_DIR` (默认: `./config`)
- `UPLOAD_DIR` (默认: `./upload`)

## 安全建议

1. **修改默认管理员密码**: 首次登录后立即修改 `admin` 账号密码
2. **配置 CORS**: 在生产环境中设置 `ALLOWED_ORIGIN` 为具体的域名
3. **使用 HTTPS**: 在生产环境中通过反向代理（如 Nginx）启用 HTTPS
4. **定期备份**: 定期备份 `config/awesome-sharing.db` 数据库文件
5. **文件权限**: 确保数据库和配置文件不可被未授权用户访问

## 许可证

请参考项目根目录的 LICENSE 文件。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

如有问题或建议，请通过 GitHub Issues 联系。
