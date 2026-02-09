import type { Photo } from 'react-photo-album'
import type { Slide, SlideImage } from 'yet-another-react-lightbox'
import type { File } from '@/types'
import { fileAPI } from '@/services/api'

// 扩展 Photo 类型，包含原始文件数据
export interface ExtendedPhoto extends Photo {
  file: File
}

// 浏览器不支持直接显示的图片格式
const UNSUPPORTED_IMAGE_FORMATS = ['.tif', '.tiff', '.heic', '.heif', '.raw', '.cr2', '.nef', '.arw', '.dng']

// 10MB 阈值
const TEN_MB = 10 * 1024 * 1024

/**
 * 检查文件扩展名是否为浏览器不支持的格式
 */
function isUnsupportedFormat(filename: string): boolean {
  const ext = filename.toLowerCase().substring(filename.lastIndexOf('.'))
  return UNSUPPORTED_IMAGE_FORMATS.includes(ext)
}

/**
 * 将 File 数组转换为 react-photo-album 需要的 Photo 数组
 */
export function filesToPhotos(files: File[]): ExtendedPhoto[] {
  return files.map((file) => ({
    src: fileAPI.getThumbnailUrl(file.id, 'medium'),
    // 如果宽高缺失，使用 4:3 默认比例避免布局异常
    width: file.width || 4,
    height: file.height || 3,
    alt: file.filename,
    key: `photo-${file.id}`,
    file,
  }))
}

// 扩展的图片 Slide 类型，包含原图 URL 用于渐进式加载
export interface ExtendedSlide extends SlideImage {
  originalSrc?: string  // 原图 URL（大文件时用于静默加载）
  fileId?: number
}

/**
 * 将 File 数组转换为 yet-another-react-lightbox 需要的 Slide 数组
 * 返回类型为 (ExtendedSlide | Slide)[] 因为视频使用不同的 slide 类型
 */
export function filesToSlides(files: File[]): (ExtendedSlide | Slide)[] {
  return files.map((file) => {
    if (file.file_type === 'video') {
      return {
        type: 'video' as const,
        width: file.width || 1280,
        height: file.height || 720,
        poster: fileAPI.getThumbnailUrl(file.id, 'large'),
        sources: [
          {
            src: fileAPI.getDownloadUrl(file.id),
            type: 'video/mp4',
          },
        ],
        fileId: file.id,
      }
    }

    // 对于浏览器不支持的格式，只使用缩略图
    if (isUnsupportedFormat(file.filename)) {
      return {
        src: fileAPI.getThumbnailUrl(file.id, 'large'),
        width: file.width || 1,
        height: file.height || 1,
        alt: file.filename,
        fileId: file.id,
      }
    }

    // 大文件：先显示缩略图，提供原图 URL 用于静默加载
    const isLargeFile = file.size > TEN_MB
    if (isLargeFile) {
      return {
        src: fileAPI.getThumbnailUrl(file.id, 'large'),
        originalSrc: fileAPI.getDownloadUrl(file.id),
        width: file.width || 1,
        height: file.height || 1,
        alt: file.filename,
        fileId: file.id,
      }
    }

    // 普通文件：直接显示原图
    return {
      src: fileAPI.getDownloadUrl(file.id),
      width: file.width || 1,
      height: file.height || 1,
      alt: file.filename,
      fileId: file.id,
    }
  })
}
