# nafotos 前端

为 NAS 设计的现代化照片管理平台前端，基于 React、TypeScript 和 Vite 构建。

[English Documentation](./README.md)

## 功能特性

- 照片和相册管理，支持上传/下载
- 按日期浏览照片的时间轴视图
- 用户认证和授权
- 多语言支持 (i18n)
- 权限组管理
- 多租户域名配置
- 基于 Tailwind CSS 的响应式设计

## 技术栈

- **React 18** - UI 框架
- **TypeScript** - 类型安全开发
- **Vite** - 快速构建工具和开发服务器
- **React Router** - 客户端路由
- **Axios** - HTTP 客户端
- **i18next** - 国际化
- **Tailwind CSS** - 实用优先的 CSS 框架
- **Lucide React** - 图标库

## 前置要求

- Node.js >= 16.0.0
- npm 或 yarn 包管理器

## 快速开始

### 安装依赖

```bash
# 安装依赖
npm install
```

### 开发模式

```bash
# 启动开发服务器,访问 http://localhost:3000
npm run dev
```

开发服务器包含:
- 热模块替换 (HMR)
- API 代理到后端 `http://localhost:8080`

### 构建

```bash
# 生产环境构建
npm run build
```

生产构建将输出到 `dist` 目录。

### 预览生产构建

```bash
# 本地预览生产构建
npm run preview
```

### 代码检查

```bash
# 运行 ESLint
npm run lint
```

## 项目结构

```
frontend/
├── public/              # 静态资源
├── src/
│   ├── components/      # 可复用的 React 组件
│   │   ├── FileGrid.tsx
│   │   ├── Layout.tsx
│   │   ├── Modal.tsx
│   │   └── ProtectedRoute.tsx
│   ├── contexts/        # React 上下文 (认证等)
│   ├── hooks/           # 自定义 React Hooks
│   ├── locales/         # i18n 翻译文件
│   ├── pages/           # 页面组件
│   │   ├── Albums.tsx
│   │   ├── DomainConfig.tsx
│   │   ├── FileDetail.tsx
│   │   ├── FolderManagement.tsx
│   │   ├── Folders.tsx
│   │   ├── Login.tsx
│   │   ├── PermissionGroupManagement.tsx
│   │   ├── Settings.tsx
│   │   ├── Timeline.tsx
│   │   └── UserManagement.tsx
│   ├── services/        # API 服务模块
│   ├── types/           # TypeScript 类型定义
│   ├── utils/           # 工具函数
│   ├── App.tsx          # 主应用组件
│   ├── main.tsx         # 应用入口点
│   └── i18n.ts          # i18n 配置
├── index.html           # HTML 模板
├── vite.config.ts       # Vite 配置
├── tailwind.config.js   # Tailwind CSS 配置
├── tsconfig.json        # TypeScript 配置
└── package.json         # 项目依赖

```

## 配置

### 环境变量

应用使用 Vite 的环境变量系统。在根目录创建 `.env` 文件:

```env
VITE_API_URL=http://localhost:8080
```

### API 代理

开发服务器配置了 API 请求代理到后端:

- 前端: `http://localhost:3000`
- 后端 API: `http://localhost:8080`
- 代理规则: `/api/*` → `http://localhost:8080/api/*`

## 可用脚本

- `npm run dev` - 启动开发服务器
- `npm run build` - 生产环境构建
- `npm run preview` - 预览生产构建
- `npm run lint` - 运行 ESLint

## 许可证

MIT
