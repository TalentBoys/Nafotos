# AwesomeSharing - 实施总结

## 项目概述

AwesomeSharing 是一个自托管的照片和文件管理系统，支持多目录管理、权限控制、相册组织和文件分享功能。

## 技术栈

### 后端
- **语言**: Go 1.24.0
- **Web 框架**: Fiber v2
- **数据库**: SQLite
- **图片处理**: disintegration/imaging
- **EXIF 提取**: rwcarlsen/goexif
- **密码加密**: bcrypt (golang.org/x/crypto)

### 前端
- **框架**: React 18 + TypeScript
- **样式**: Tailwind CSS + Shadcn UI
- **路由**: React Router v6
- **国际化**: react-i18next
- **构建工具**: Vite

## 核心功能实现 ✅

### 1. 用户管理与权限控制
**位置**: `backend/internal/services/auth.go`, `backend/internal/api/auth_handlers.go`, `backend/internal/api/user_handlers.go`

**已实现功能**:
- 用户注册与登录（支持可配置的注册开关）
- 基于 session 的身份认证（7天有效期）
- 三种用户角色：`server_owner`（超级管理员）、`admin`（管理员）、`user`（普通用户）
- 用户 CRUD 操作（创建、查询、更新、删除）
- 用户启用/禁用功能
- 密码修改与重置
- 批量用户操作（批量启用/禁用、批量删除）
- 用户活动日志（记录用户管理操作）
- 用户统计信息和搜索功能
- 用户导出功能

**数据库表**:
- `users` - 用户信息
- `sessions` - 会话管理
- `user_activity_logs` - 用户活动审计日志

**前端实现**:
- 登录页面：`frontend/src/pages/Login.tsx`
- 用户管理页面：`frontend/src/pages/UserManagement.tsx`
- 认证上下文：`frontend/src/contexts/AuthContext.tsx`
- 路由保护：`frontend/src/components/ProtectedRoute.tsx`

### 2. 文件夹（Folder）系统
**位置**: `backend/internal/services/folder.go`, `backend/internal/api/folder_handlers.go`

**核心特性**:
- 管理员可以创建文件夹，指向文件系统中的任意绝对路径
- 每个文件夹包含：名称、绝对路径、启用状态、创建者
- 文件夹可以被启用/禁用
- 支持手动触发文件夹扫描
- 查询文件夹中的文件列表（带权限过滤）

**数据库表**:
- `folders` - 文件夹元数据
- `file_folder_mappings` - 文件到文件夹的映射关系（存储相对路径）

**设计亮点**:
- 用户可以配置系统中的任意路径，不限于 `/config` 或 `/upload`
- 通过相对路径存储文件位置，即使文件夹移动也能重新关联

**前端实现**:
- 文件夹管理页面：`frontend/src/pages/FolderManagement.tsx`
- 文件夹浏览页面：`frontend/src/pages/Folders.tsx`

### 3. 权限组（Permission Group）系统
**位置**: `backend/internal/services/permission_group.go`, `backend/internal/api/permission_group_handlers.go`

**核心功能**:
- 创建权限组（包含多个文件夹）
- 为权限组分配用户并设置权限（`read` 或 `write`）
- 管理权限组中的文件夹（添加/移除）
- 查询用户的权限组及其访问权限

**权限逻辑**:
- `admin` 和 `server_owner` 角色自动拥有所有权限
- 普通用户通过权限组获得特定文件夹的访问权限
- 支持读权限（只能查看）和写权限（可以修改）

**数据库表**:
- `permission_groups` - 权限组元数据
- `permission_group_folders` - 权限组包含的文件夹
- `permission_group_permissions` - 用户对权限组的权限

**前端实现**:
- 权限组管理页面：`frontend/src/pages/PermissionGroupManagement.tsx`

### 4. 增强型相册（Albums V2）系统
**位置**: `backend/internal/services/album.go`, `backend/internal/api/album_handlers.go`

**核心创新 - 软链接（Soft Link）**:
- 相册项存储为 `(folder_id, relative_path)` 而不是直接存储 `file_id`
- 当文件夹路径变更时，更新文件夹配置并重新扫描即可
- 相册自动通过相对路径重新解析文件，避免相册失效

**已实现功能**:
- 创建/更新/删除相册
- 添加/移除相册项（通过 folder_id + relative_path）
- 查询相册及其所有项
- 设置相册封面
- 解析相册项（重新关联移动后的文件）
- 批量解析所有相册

**数据库表**:
- `albums_v2` - 相册元数据（包含 owner_id）
- `album_items` - 相册项（存储 folder_id + relative_path + 当前 file_id）

**使用流程示例**:
```
1. 文件夹 X 指向路径：/photos/vacation
2. 用户添加 /photos/vacation/beach.jpg 到相册 A
3. 系统存储：(folder_id=X, relative_path="vacation/beach.jpg")
4. 管理员移动文件夹：/photos/vacation → /photos/2024-vacation
5. 管理员更新文件夹 X 的路径为 /photos/2024-vacation
6. 系统重新扫描
7. 相册 A 自动重新解析，beach.jpg 仍然可用！
```

**前端实现**:
- 相册页面：`frontend/src/pages/Albums.tsx`

### 5. 文件分享（Share）系统
**位置**: `backend/internal/services/share.go`, `backend/internal/api/share_handlers.go`

**分享类型**:
- 文件分享（单个文件）
- 相册分享（整个相册）

**访问控制**:
- **Public 公开分享**: 任何人都可以通过链接访问（匿名）
- **Private 私有分享**: 仅指定用户可以访问（需要登录）

**高级特性**:
- 密码保护（可选）
- 过期设置：
  - 按小时（如 24 小时）
  - 按天（如 7 天）
  - 永久（无过期时间）
- 最大访问次数限制
- 访问计数器
- 启用/禁用分享
- 访问日志（记录访问者、IP、UserAgent、访问时间）

**管理功能**:
- 查看自己的所有分享
- 查看分享的访问日志
- 延长分享过期时间
- 批量删除过期分享
- 为私有分享授权/撤销用户权限

**分享链接格式**: `/api/s/:shareId`

**数据库表**:
- `shares` - 分享元数据和设置
- `share_permissions` - 私有分享的用户权限
- `share_access_log` - 访问审计日志

### 6. 域名配置（Domain Config）
**位置**: `backend/internal/services/domain_config.go`, `backend/internal/api/domain_config_handlers.go`

**功能**:
- 配置系统的访问域名/IP（如 `qjkobe.online:1234`）
- 分别配置协议（http/https）、域名、端口
- 用于生成完整的分享链接
- 管理员专属功能

**数据库表**:
- `domain_config` - 域名配置

**前端实现**:
- 域名配置页面：`frontend/src/pages/DomainConfig.tsx`

### 7. 系统设置（System Settings）
**位置**: `backend/internal/services/settings.go`, `backend/internal/api/settings_handlers.go`

**可配置项**:
- 网站名称
- 注册开关（是否允许新用户注册）
- 其他系统级配置（Key-Value 存储）

**数据库表**:
- `system_settings` - 系统配置（Key-Value）

**前端实现**:
- 设置页面：`frontend/src/pages/Settings.tsx`

### 8. 文件扫描与缩略图生成
**位置**: `backend/internal/services/scanner.go`, `backend/internal/services/thumbnail.go`

**扫描服务**:
- 自动扫描所有启用的文件夹
- 提取文件元数据（EXIF、拍摄日期、尺寸等）
- 建立文件与文件夹的映射关系
- 定时扫描（每 30 分钟）
- 支持手动触发扫描

**缩略图生成**:
- 自动为图片生成缩略图（多种尺寸）
- 支持格式：JPEG、PNG、HEIC、TIFF 等
- 缩略图存储在 `config/thumbs/` 目录
- 懒加载生成（首次访问时生成）

**文件验证服务**:
- 定期验证数据库中的文件是否仍存在于文件系统
- 清理失效的文件记录
- 定时运行（每 6 小时）
- 可通过环境变量 `DISABLE_FILE_VALIDATION=true` 禁用

**数据库表**:
- `files` - 文件元数据
- `file_thumbnails` - 缩略图信息

### 9. 认证中间件
**位置**: `backend/internal/middleware/auth.go`

**中间件类型**:
- `AuthMiddleware` - 要求有效的 session（强制登录）
- `OptionalAuthMiddleware` - 可选认证（注入用户信息，但不强制登录）
- `AdminOnlyMiddleware` - 仅允许 admin 和 server_owner 访问
- `AdminOrOwnerMiddleware` - 仅允许 admin 和 server_owner 访问

**Session 来源**:
- Cookie 中的 session_id
- Authorization header 中的 Bearer token

### 10. 标签（Tags）系统
**位置**: `backend/internal/api/handlers.go`（待完善）

**基础功能**:
- 创建标签（带颜色）
- 为文件添加标签
- 查询标签列表

**数据库表**:
- `tags` - 标签元数据
- `file_tags` - 文件与标签的多对多关系

## 数据库架构（Schema V3）

**核心表结构**:
```
users                        # 用户
sessions                     # 会话
user_activity_logs           # 用户活动日志
files                        # 文件元数据
folders                      # 文件夹配置
file_folder_mappings         # 文件-文件夹映射（相对路径）
permission_groups            # 权限组
permission_group_folders     # 权限组-文件夹关联
permission_group_permissions # 权限组-用户权限
albums_v2                    # 相册（V2 版本）
album_items                  # 相册项（软链接）
tags                         # 标签
file_tags                    # 文件-标签关联
shares                       # 分享
share_permissions            # 私有分享权限
share_access_log             # 分享访问日志
system_settings              # 系统设置
domain_config                # 域名配置
file_thumbnails              # 缩略图
```

**数据库特性**:
- 外键约束与级联删除
- 性能优化索引
- 自动时间戳
- 数据完整性保证

## API 路由架构

**公开路由（无需认证）**:
- `GET /api/health` - 健康检查
- `GET /api/settings/public` - 公开设置
- `GET /api/s/:id` - 访问分享链接

**认证路由**:
- `POST /api/auth/login` - 登录
- `POST /api/auth/register` - 注册
- `POST /api/auth/logout` - 登出
- `GET /api/auth/me` - 获取当前用户信息
- `POST /api/auth/change-password` - 修改密码

**受保护路由（需要认证）**:
- `/api/users/*` - 用户管理（仅管理员）
- `/api/folders/*` - 文件夹管理
- `/api/permission-groups/*` - 权限组管理
- `/api/albums-v2/*` - 相册管理（V2）
- `/api/shares/*` - 分享管理
- `/api/settings/*` - 系统设置（仅管理员）
- `/api/domain-config/*` - 域名配置（仅管理员）
- `/api/files/*` - 文件访问（向后兼容）
- `/api/timeline` - 时间线视图
- `/api/search` - 文件搜索
- `/api/scan` - 触发扫描
- `/api/cleanup` - 清理失效文件
- `/api/tags/*` - 标签管理

## 项目结构

```
AwesomeSharing/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # 服务器入口
│   ├── internal/
│   │   ├── api/
│   │   │   ├── routes_v2.go         # 路由配置（V2）
│   │   │   ├── auth_handlers.go     # 认证处理器
│   │   │   ├── user_handlers.go     # 用户管理处理器
│   │   │   ├── folder_handlers.go   # 文件夹处理器
│   │   │   ├── permission_group_handlers.go  # 权限组处理器
│   │   │   ├── album_handlers.go    # 相册处理器
│   │   │   ├── share_handlers.go    # 分享处理器
│   │   │   ├── settings_handlers.go # 设置处理器
│   │   │   ├── domain_config_handlers.go # 域名配置处理器
│   │   │   └── handlers.go          # 其他处理器
│   │   ├── services/
│   │   │   ├── auth.go              # 认证服务
│   │   │   ├── folder.go            # 文件夹服务
│   │   │   ├── permission_group.go  # 权限组服务
│   │   │   ├── album.go             # 相册服务
│   │   │   ├── share.go             # 分享服务
│   │   │   ├── settings.go          # 设置服务
│   │   │   ├── domain_config.go     # 域名配置服务
│   │   │   ├── scanner.go           # 文件扫描服务
│   │   │   ├── thumbnail.go         # 缩略图服务
│   │   │   └── file_validator.go    # 文件验证服务
│   │   ├── middleware/
│   │   │   └── auth.go              # 认证中间件
│   │   ├── models/
│   │   │   └── models.go            # 数据模型
│   │   ├── database/
│   │   │   ├── database.go          # 数据库初始化
│   │   │   └── schema_v3.go         # 数据库架构 V3
│   │   ├── config/
│   │   │   └── config.go            # 配置管理
│   │   └── initialization/
│   │       └── init.go              # 系统初始化
│   ├── pkg/
│   │   ├── utils/                   # 工具函数
│   │   └── exif/                    # EXIF 工具
│   ├── go.mod
│   └── run-local-v2.sh              # 本地启动脚本
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── Layout.tsx           # 布局组件
│   │   │   ├── ProtectedRoute.tsx   # 路由保护
│   │   │   ├── Modal.tsx            # 模态框
│   │   │   └── FileGrid.tsx         # 文件网格
│   │   ├── pages/
│   │   │   ├── Login.tsx            # 登录页
│   │   │   ├── UserManagement.tsx   # 用户管理
│   │   │   ├── FolderManagement.tsx # 文件夹管理
│   │   │   ├── Folders.tsx          # 文件夹浏览
│   │   │   ├── PermissionGroupManagement.tsx # 权限组管理
│   │   │   ├── Albums.tsx           # 相册
│   │   │   ├── Timeline.tsx         # 时间线
│   │   │   ├── Settings.tsx         # 设置
│   │   │   ├── DomainConfig.tsx     # 域名配置
│   │   │   └── FileDetail.tsx       # 文件详情
│   │   ├── contexts/
│   │   │   └── AuthContext.tsx      # 认证上下文
│   │   ├── locales/                 # 国际化翻译
│   │   ├── types/                   # TypeScript 类型
│   │   └── App.tsx
│   ├── package.json
│   └── vite.config.ts
├── config/                          # 配置目录（运行时生成）
│   ├── awesome-sharing.db           # SQLite 数据库
│   └── thumbs/                      # 缩略图目录
├── upload/                          # 上传目录（运行时生成）
├── .env.example
└── README.md
```

## 首次启动流程

1. 服务器启动时自动执行：
   - 创建 SQLite 数据库文件
   - 初始化数据库表结构（Schema V3）
   - 创建默认管理员账户：
     - 用户名: `admin`
     - 密码: `admin`
     - 角色: `server_owner`
   - 启动后台文件扫描服务（每 30 分钟）
   - 启动文件验证服务（每 6 小时）

2. 管理员登录后：
   - 修改默认密码（安全建议）
   - 创建文件夹（指向文件系统路径）
   - 创建权限组并分配文件夹
   - 创建普通用户并授予权限
   - 配置域名（用于分享链接）
   - 配置系统设置

3. 普通用户使用流程：
   - 登录系统
   - 浏览自己有权限的文件夹
   - 创建相册并添加文件
   - 创建分享链接（公开或私有）
   - 管理自己的分享

## 关键设计决策

### 1. Folder（文件夹）替代 Library（文件库）
- **原设计**: 每个 Library 可以包含多个路径
- **当前设计**: 每个 Folder 对应一个绝对路径
- **优势**: 更简单直接，每个路径独立管理

### 2. 软链接（Soft Link）解决文件移动问题
传统系统在文件移动后相册会失效，本系统通过 `(folder_id, relative_path)` 实现软链接：
- 相册项不直接绑定文件 ID
- 文件夹路径变更后重新扫描即可恢复关联
- 大幅提升了文件管理的灵活性

### 3. 三层权限控制
- **角色层**: server_owner > admin > user
- **权限组层**: 通过权限组控制文件夹访问
- **分享层**: 公开分享（匿名）vs 私有分享（指定用户）

### 4. Session 认证（非 JWT）
- 更简单的实现
- 更容易撤销（删除 session 记录即可）
- 更适合小型自托管应用

### 5. 完整的分享功能
- 过期控制避免永久链接泄露
- 访问次数限制防止滥用
- 访问日志用于审计
- 密码保护增强安全性
- 私有分享实现团队协作

## 环境变量配置

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 服务器端口 |
| `CONFIG_DIR` | `/config` | 配置目录路径（存储数据库和缩略图） |
| `UPLOAD_DIR` | `/upload` | 上传目录路径 |
| `ALLOWED_ORIGIN` | `*` | CORS 允许的来源（生产环境建议设置为具体域名） |
| `DISABLE_FILE_VALIDATION` | `false` | 禁用文件验证（设为 `true` 禁用） |

## 后台服务

服务器启动后，以下后台任务会自动运行：

1. **文件扫描器**: 每 30 分钟扫描所有启用的文件夹，发现新文件
2. **文件验证器**: 每 6 小时验证数据库中的文件是否存在，清理失效记录
3. **Session 清理**: 每 1 小时清理过期的 session

## 已完成的实现状态

✅ **后端完成度: ~95%**
- 所有核心服务已实现
- 所有 API 处理器已实现
- 数据库架构完整
- 中间件完整
- 后台任务完整

✅ **前端完成度: ~80%**
- 认证系统完整（登录、路由保护、上下文）
- 用户管理页面完整
- 文件夹管理页面完整
- 权限组管理页面完整
- 相册页面基础功能完整
- 域名配置页面完整
- 设置页面完整

⚠️ **待完善的功能**:
- 分享管理前端页面（当前仅后端 API）
- 文件详情页的增强功能
- 相册项的拖拽排序
- 批量文件操作
- 高级搜索过滤
- 用户头像上传

## 安全建议

1. **修改默认密码**: 首次登录后立即修改 `admin` 账户密码
2. **配置 CORS**: 生产环境设置 `ALLOWED_ORIGIN` 为具体域名
3. **启用 HTTPS**: 通过反向代理（如 Nginx）启用 HTTPS
4. **定期备份**: 定期备份 `config/awesome-sharing.db` 数据库文件
5. **文件权限**: 确保数据库和配置文件不被未授权用户访问

## 技术特点

- **bcrypt** 密码哈希（安全且慢，防暴力破解）
- **Session 认证**（简单且易撤销）
- **SQLite** 数据库（无需独立数据库服务器）
- **Fiber** 框架（快速、Express 风格）
- **React + TypeScript**（类型安全的前端）
- **响应式设计**（支持桌面和移动端）
- **国际化支持**（中文/英文）

## 文档

- 完整的后端 API 文档：`backend/README.md`
- 后端中文文档：`backend/README-ZH_CN.md`
- 前端文档：`frontend/README.md`
- 前端中文文档：`frontend/README-ZH_CN.md`

## 使用示例

### 管理员首次配置
```bash
1. 启动服务器
2. 使用 admin/admin 登录
3. 修改管理员密码
4. 创建文件夹（如：/photos/vacation）
5. 创建权限组（如：家庭相册）
6. 将文件夹添加到权限组
7. 创建普通用户
8. 为用户授予权限组的访问权限
9. 配置域名（用于生成分享链接）
```

### 普通用户使用
```bash
1. 登录系统
2. 浏览有权限的文件夹
3. 创建相册，添加喜欢的照片
4. 创建分享链接（设置过期时间、密码等）
5. 将分享链接发给朋友
6. 查看访问日志了解谁访问了分享
```

### 文件移动场景
```bash
1. 管理员在文件系统移动文件夹：/photos/2023 → /photos/archived/2023
2. 管理员更新 Folder 配置：修改路径为 /photos/archived/2023
3. 触发扫描
4. 所有相册自动重新解析，照片关联恢复
5. 用户无感知，相册正常使用
```

## 项目状态

🎉 **核心功能已全部实现并可用！**

当前系统已经可以：
- 完整的用户管理和权限控制
- 多文件夹管理和扫描
- 权限组和细粒度访问控制
- 相册创建和管理（带软链接）
- 分享链接生成和管理
- 域名配置和系统设置
- 自动缩略图生成
- 文件验证和清理
- 完整的审计日志

## License

MIT License
