# 文件验证和自动清理系统

## 概述

文件验证系统会自动检测并清理已删除文件的数据库记录，确保用户界面不会显示已不存在的文件和损坏的缩略图。

## 问题

当用户直接从文件系统删除图片文件时，数据库中仍保留该文件的记录，导致：
1. Timeline 和文件列表显示已删除的文件
2. 显示损坏的缩略图图标
3. 用户体验差，无法区分有效文件和已删除文件

## 解决方案

### 1. FileValidatorService

新增的文件验证服务 (`backend/internal/services/file_validator.go`) 提供以下功能：

#### 核心功能

1. **实时验证** - `ValidateFiles()`
   - 在返回文件列表前验证每个文件是否存在
   - 自动过滤掉不存在的文件
   - 后台异步清理无效记录

2. **批量清理** - `CleanupAllInvalidFiles()`
   - 扫描整个数据库
   - 检查每个文件记录是否对应真实文件
   - 删除无效记录和关联的缩略图

3. **缩略图清理**
   - 删除文件记录时，自动删除相关的缩略图文件
   - 通过外键级联删除 `file_thumbnails` 表中的记录

#### 缓存机制

使用内存缓存避免重复清理同一个文件：
```go
cleanupCache map[int64]bool
```

### 2. 自动验证集成

所有文件列表 API 现在都会自动验证文件存在性：

- `GET /api/files` - 文件列表
- `GET /api/timeline` - 时间线
- `GET /api/search` - 搜索
- `GET /api/libraries/:id/files` - 文件库列表

每次查询后会调用：
```go
files = h.validator.ValidateFiles(files)
```

### 3. 定期自动清理

服务器启动后会自动运行定期清理任务：

- **启动清理**：服务器启动后 10 秒执行首次清理
- **定期清理**：每 6 小时自动运行一次清理
- **日志记录**：清理结果会记录到服务器日志

启动日志示例：
```
✓ Background file validator scheduled (first cleanup in 10 seconds)
Running initial file validation and cleanup...
✓ Initial cleanup: removed 5 missing files
```

### 4. 手动清理 API

管理员可以通过 API 手动触发清理：

**端点**：
```
POST /api/cleanup
```

**响应**：
```json
{
  "message": "Cleanup completed",
  "deleted": 5
}
```

**使用方法**：
```bash
# 使用 curl
curl -X POST http://localhost:8080/api/cleanup \
  -H "Cookie: session_id=your_session_id"

# 或在前端
fetch('/api/cleanup', { method: 'POST' })
```

## 工作流程

### 文件列表请求流程

```
1. 用户请求 /api/timeline
   ↓
2. 数据库查询返回文件列表
   ↓
3. ValidateFiles() 验证每个文件
   ↓
4. 过滤掉不存在的文件
   ↓
5. 返回有效文件列表给用户
   ↓
6. 后台异步清理无效记录
```

### 清理流程

```
1. 检测文件不存在
   ↓
2. 从 files 表删除记录
   ↓
3. 外键级联删除 file_thumbnails 记录
   ↓
4. 删除文件系统中的缩略图文件
   ↓
5. 更新缓存避免重复清理
   ↓
6. 记录日志
```

## 性能优化

1. **异步清理**：文件验证后，清理操作在后台 goroutine 执行，不阻塞 API 响应
2. **缓存机制**：避免对同一文件重复执行清理操作
3. **批量操作**：定期清理一次性处理所有无效文件，减少数据库操作

## 配置

清理间隔可以在 `main.go` 中调整：

```go
// 当前配置：每 6 小时清理一次
ticker := time.NewTicker(6 * time.Hour)

// 修改为其他间隔，例如每小时：
ticker := time.NewTicker(1 * time.Hour)
```

## 日志

系统会记录以下清理相关日志：

```
// 启动时
✓ Background file validator scheduled (first cleanup in 10 seconds)

// 清理执行
Running initial file validation and cleanup...
Cleaned up missing file record: 123
Cleaned up missing file record: 456
File validation complete: 2 invalid out of 150 total files cleaned up
✓ Initial cleanup: removed 2 missing files

// 定期清理
✓ Periodic cleanup: removed 1 missing files
```

## 与缩略图系统的集成

文件验证系统与缩略图系统完美集成：

1. 删除原图记录时，自动清理所有相关缩略图
2. `file_thumbnails` 表通过外键级联删除
3. 缩略图文件也从文件系统中删除

## 前端集成建议

### 手动触发清理（管理界面）

```typescript
const cleanupFiles = async () => {
  try {
    const response = await fetch('/api/cleanup', { method: 'POST' });
    const result = await response.json();
    alert(`清理完成，删除了 ${result.deleted} 个无效文件`);
  } catch (error) {
    alert('清理失败');
  }
};
```

### 显示清理状态

```typescript
// 在设置页面添加清理按钮
<button onClick={cleanupFiles} className="btn">
  清理已删除文件
</button>
```

## 注意事项

1. **权限要求**：手动清理 API 需要管理员权限
2. **性能影响**：大型数据库的全量清理可能需要一些时间，建议在低峰时段执行
3. **不可逆操作**：清理操作会永久删除数据库记录，请确保文件确实不存在
4. **备份建议**：重要数据请定期备份数据库

## 故障排查

### 文件仍然显示为损坏

1. 检查定期清理是否正常运行（查看日志）
2. 手动执行清理：`POST /api/cleanup`
3. 重启服务器触发启动清理

### 清理速度慢

1. 减少清理间隔会增加 I/O 负载
2. 大型数据库建议保持 6 小时间隔
3. 可以在夜间手动执行全量清理

### 误删除文件

1. 文件验证基于 `os.Stat()`，只有文件真正不存在才会删除
2. 如果文件权限问题导致访问失败，可能会误判
3. 建议确保应用有足够的文件系统权限
