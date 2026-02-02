# 支持的文件格式

AwesomeSharing 支持多种图片和视频格式。

## 图片格式

系统支持以下图片格式：

| 格式 | 扩展名 | 描述 | EXIF 支持 | 缩略图生成 |
|------|--------|------|-----------|-----------|
| JPEG | `.jpg`, `.jpeg` | 最常见的照片格式 | ✅ | ✅ |
| PNG | `.png` | 无损压缩，支持透明度 | ✅ | ✅ |
| GIF | `.gif` | 动图格式 | ❌ | ✅ |
| BMP | `.bmp` | Windows 位图格式 | ❌ | ✅ |
| WebP | `.webp` | Google 开发的现代格式 | ✅ | ✅ |
| HEIC | `.heic` | iOS 默认照片格式 | ✅ | ✅ |
| HEIF | `.heif` | 高效图像格式 | ✅ | ✅ |
| **TIFF** | **`.tif`, `.tiff`** | **专业摄影和扫描格式** | ✅ | ✅ |

### TIFF 格式说明

TIFF (Tagged Image File Format) 是一种灵活的图像格式，广泛用于：
- 专业摄影
- 扫描文档
- 医学影像
- 印刷出版

**特点**：
- 支持无损压缩
- 可以包含多页
- 支持高位深度（16-bit, 32-bit）
- 文件通常较大
- 完整的 EXIF 元数据支持

**注意事项**：
- TIFF 文件通常很大（10MB - 100MB+）
- 系统会自动生成缩略图以提高浏览速度
- 详情页默认显示缩略图（大于 10MB）
- 用户可点击"查看原图"按钮查看完整质量

## 视频格式

系统支持以下视频格式：

| 格式 | 扩展名 | 描述 | 缩略图生成 |
|------|--------|------|-----------|
| MP4 | `.mp4` | 最常见的视频格式 | 🔜 计划中 |
| MOV | `.mov` | Apple QuickTime 格式 | 🔜 计划中 |
| AVI | `.avi` | Windows 视频格式 | 🔜 计划中 |
| MKV | `.mkv` | Matroska 多媒体容器 | 🔜 计划中 |
| WebM | `.webm` | Web 视频格式 | 🔜 计划中 |
| M4V | `.m4v` | iTunes 视频格式 | 🔜 计划中 |

## 技术实现

### 图片解码

系统使用以下 Go 图像解码库：

```go
import (
    _ "image/jpeg"  // JPEG 支持
    _ "image/png"   // PNG 支持
    _ "image/gif"   // GIF 支持
    _ "golang.org/x/image/tiff"  // TIFF 支持
    _ "golang.org/x/image/bmp"   // BMP 支持
    _ "golang.org/x/image/webp"  // WebP 支持
)
```

### 缩略图生成

所有图片格式都使用 `github.com/disintegration/imaging` 库生成缩略图：
- 自动保持纵横比
- 高质量 Lanczos 重采样
- 输出为 JPEG 格式（85% 质量）

### EXIF 数据提取

使用 `github.com/rwcarlsen/goexif/exif` 库提取 EXIF 数据：
- 拍摄时间
- 相机信息
- GPS 位置（计划中）
- 图片尺寸

## 添加新格式

如需支持新的图片格式，需要：

1. **更新扫描器** (`internal/services/scanner.go`)
   ```go
   imageExts := []string{
       ".jpg", ".jpeg", ".png", ".gif",
       ".bmp", ".webp", ".heic", ".heif",
       ".tif", ".tiff",
       ".new_format"  // 添加新格式
   }
   ```

2. **导入解码器** (`internal/services/thumbnail.go`)
   ```go
   import (
       _ "package/for/new/format"  // 新格式解码器
   )
   ```

3. **测试**
   - 放置测试文件到监控目录
   - 触发扫描：`POST /api/scan`
   - 验证文件被正确索引
   - 检查缩略图生成

## 性能考虑

### 大文件处理

对于大型图片文件（例如 TIFF、高分辨率 JPEG）：

1. **扫描阶段**
   - 只提取元数据和尺寸
   - 不加载完整图片到内存

2. **缩略图生成**
   - 按需生成（首次访问时）
   - 缓存到磁盘
   - 支持三种尺寸（small, medium, large）

3. **前端显示**
   - Timeline：小缩略图 (300x300)
   - 详情页（< 10MB）：原图
   - 详情页（> 10MB）：大缩略图 (1920x1920) + "查看原图"按钮

### 内存使用

缩略图生成时的内存使用估算：

| 原图分辨率 | 原图大小 | 内存占用 | 生成时间 |
|-----------|---------|---------|---------|
| 6000x4000 (24MP) | ~20MB JPEG | ~100MB | ~1s |
| 6000x4000 (24MP) | ~80MB TIFF | ~100MB | ~3s |
| 8000x6000 (48MP) | ~150MB TIFF | ~200MB | ~5s |

**注意**：TIFF 文件解码比 JPEG 慢 2-3 倍，但生成的缩略图相同。

## 常见问题

### Q: 为什么我的 TIFF 文件加载很慢？

A: TIFF 文件通常很大（50-100MB），首次访问需要生成缩略图。后续访问会使用缓存的缩略图，速度会快很多。

### Q: 可以直接查看 TIFF 原图吗？

A: 可以！在详情页点击"查看原图"按钮即可加载完整的 TIFF 文件。

### Q: TIFF 缩略图质量如何？

A: 缩略图使用高质量 Lanczos 重采样算法生成，即使从大型 TIFF 文件生成，质量也非常好。

### Q: 系统支持多页 TIFF 吗？

A: 目前只读取第一页。多页 TIFF 支持在规划中。

### Q: 支持 RAW 格式吗（.cr2, .nef, .arw）？

A: 目前不支持。RAW 格式支持在考虑中，可能会在未来版本添加。

## 浏览器兼容性

### 原生支持格式

现代浏览器原生支持：
- JPEG, PNG, GIF, WebP

### 需要转换的格式

以下格式在浏览器中显示时会使用生成的 JPEG 缩略图：
- TIFF（浏览器不原生支持）
- BMP（部分浏览器支持）
- HEIC/HEIF（大多数浏览器不支持）

这确保了所有格式都能在任何浏览器中正常显示。
