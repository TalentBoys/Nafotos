# 缩略图系统设计文档

## 概述

本文档描述了 AwesomeSharing 的缩略图系统架构，解决了原图和缩略图重复显示的问题。

## 问题

之前的系统存在以下问题：
1. 缩略图文件被扫描器当作独立文件索引到数据库
2. Timeline 和文件列表显示了重复的图片（原图和缩略图都显示）
3. 数据库中没有原图和缩略图的关联关系

## 解决方案

### 1. 数据库结构

#### files 表新增字段
- `is_thumbnail` (BOOLEAN): 标识文件是否为缩略图，默认值为 0
- `parent_file_id` (INTEGER): 缩略图对应的原图 ID，外键关联到 files 表

#### 新表：file_thumbnails
存储缩略图的元数据信息：
```sql
CREATE TABLE file_thumbnails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,           -- 原图 ID
    size_type TEXT NOT NULL,            -- 缩略图类型: small, medium, large
    width INTEGER NOT NULL,             -- 缩略图宽度
    height INTEGER NOT NULL,            -- 缩略图高度
    file_size INTEGER NOT NULL,         -- 文件大小
    path TEXT NOT NULL,                 -- 缩略图路径
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    UNIQUE(file_id, size_type)
);
```

### 2. 缩略图尺寸定义

系统支持三种缩略图尺寸：

- **small**: 300x300 - 用于列表、时间线视图
- **medium**: 800x800 - 用于详情页预览（文件 > 10MB）
- **large**: 1920x1920 - 用于详情页预览（文件 > 50MB）

### 3. 文件扫描器改进

`scanner.go` 现在会：
- 跳过 `thumbs` 目录，不索引其中的缩略图文件
- 只索引原始媒体文件
- 所有新索引的文件 `is_thumbnail` 默认为 0

### 4. 缩略图服务改进

`ThumbnailService` 支持：
- 按需生成指定尺寸的缩略图
- API 调用：`GetThumbnail(originalPath, fileID, sizeType)`
- 缩略图文件命名格式：`{fileID}_{hash}_{size}.jpg`

### 5. API 改进

#### 缩略图 API
```
GET /api/files/:id/thumbnail?size=small|medium|large
```

参数：
- `size`: 可选，默认 `small`，可选值：`small`, `medium`, `large`

#### 文件列表查询
所有文件列表 API 现在自动过滤缩略图：
- `GET /api/files` - 文件列表
- `GET /api/timeline` - 时间线
- `GET /api/search` - 搜索
- `GET /api/libraries/:id/files` - 文件库文件列表

查询条件：`WHERE (is_thumbnail IS NULL OR is_thumbnail = 0)`

### 6. 前端集成建议

#### Timeline 视图
```typescript
// 使用 small 缩略图
<img src={`/api/files/${file.id}/thumbnail?size=small`} />
```

#### 详情页
```typescript
// 根据文件大小决定
const thumbnailSize = file.size > 50_000_000 ? 'large'
  : file.size > 10_000_000 ? 'medium'
  : 'small';

<img src={`/api/files/${file.id}/thumbnail?size=${thumbnailSize}`} />

// 提供"查看原图"按钮
<button onClick={() => window.open(`/api/files/${file.id}/download`)}>
  查看原图
</button>
```

## 迁移步骤

### 1. 清理现有数据库

运行清理脚本删除已索引的缩略图：
```bash
sqlite3 data/awesomesharing.db < scripts/cleanup_thumbnails.sql
```

### 2. 重启服务

数据库迁移会自动运行：
```bash
./backend
```

### 3. 重新扫描（可选）

如果需要重新索引所有文件：
```bash
curl -X POST http://localhost:8080/api/scan
```

## 注意事项

1. **向后兼容**：使用 `(is_thumbnail IS NULL OR is_thumbnail = 0)` 确保现有未迁移的数据也能正常工作
2. **缩略图缓存**：缩略图生成后会缓存在 `thumbs` 目录，重复请求直接返回
3. **自动清理**：删除原图时，相关缩略图会通过外键级联删除
4. **性能优化**：为 `is_thumbnail` 字段创建了索引，提高查询性能

## 未来改进

1. 支持视频缩略图
2. 智能预加载：根据网络状况自动选择缩略图尺寸
3. WebP 格式支持，进一步减小文件大小
4. CDN 集成，加速缩略图加载
