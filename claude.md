# AwesomeSharing

自托管照片与文件管理系统，支持多目录管理、权限控制、相册组织、文件分享、缩略图生成、EXIF 提取、国际化（中/英）。

## 技术栈

### 前端
- **框架**: React 18 + TypeScript + Vite
- **路由**: React Router DOM 6
- **样式**: Tailwind CSS
- **状态**: React Context API
- **HTTP**: Axios
- **国际化**: i18next

### 后端
- **语言**: Go 1.24
- **框架**: Fiber v2
- **数据库**: SQLite
- **认证**: Session-based (Cookie)
- **图像处理**: disintegration/imaging
- **EXIF**: rwcarlsen/goexif

## 项目结构

```
frontend/src/
├── components/    # UI 组件
├── pages/         # 页面
├── services/      # API 服务
├── contexts/      # 状态管理
├── hooks/         # 自定义 hooks
└── locales/       # 国际化

backend/internal/
├── api/           # HTTP 处理器
├── services/      # 业务逻辑
├── models/        # 数据模型
├── middleware/    # 中间件
└── database/      # 数据库
```

## 提示

可以使用 Tavily MCP 进行联网搜索，获取最新技术文档和解决方案。
