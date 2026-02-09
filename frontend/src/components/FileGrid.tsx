import { useState, useEffect, useCallback, useMemo } from 'react'
import { createPortal } from 'react-dom'
import { useTranslation } from 'react-i18next'
import { RowsPhotoAlbum, RenderPhotoContext } from 'react-photo-album'
import 'react-photo-album/rows.css'

import Lightbox from 'yet-another-react-lightbox'
import Video from 'yet-another-react-lightbox/plugins/video'
import Thumbnails from 'yet-another-react-lightbox/plugins/thumbnails'
import Zoom from 'yet-another-react-lightbox/plugins/zoom'
import 'yet-another-react-lightbox/styles.css'
import 'yet-another-react-lightbox/plugins/thumbnails.css'

import type { File } from '@/types'
import { fileAPI } from '@/services/api'
import { Play, Download, Info, Share2, X } from 'lucide-react'
import { format } from 'date-fns'
import { filesToPhotos, filesToSlides, ExtendedPhoto } from '@/utils/photoAdapter'
import ShareDialog from '@/components/ShareDialog'

const cn = (...classes: (string | boolean | undefined)[]) => {
  return classes.filter(Boolean).join(' ')
}

function formatFileSize(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(2)} ${units[unitIndex]}`
}

// 侧边栏组件
function LightboxSidebar({
  file,
  show,
  onClose
}: {
  file: File | null
  show: boolean
  onClose: () => void
}) {
  const { t } = useTranslation()

  if (!file || !show) return null

  return createPortal(
    <div
      className="fixed top-0 right-0 bottom-0 w-80 bg-background border-l border-border overflow-y-auto"
      style={{ zIndex: 10000, pointerEvents: 'auto' }}
    >
      <div className="p-6 space-y-6">
        <div className="flex items-center gap-3">
          <button
            type="button"
            onClick={(e) => {
              e.preventDefault()
              e.stopPropagation()
              onClose()
            }}
            className="p-2 rounded hover:bg-secondary cursor-pointer"
            style={{ pointerEvents: 'auto' }}
          >
            <X size={20} />
          </button>
          <h2 className="text-lg font-semibold">{t('file.details')}</h2>
        </div>

        <div>
          <h1 className="text-xl font-bold mb-2 break-words">{file.filename}</h1>
        </div>

        <div className="space-y-4">
          <div className="border-b border-border pb-3">
            <p className="text-sm text-muted-foreground">{t('file.type')}</p>
            <p className="font-medium capitalize">{file.file_type}</p>
          </div>

          <div className="border-b border-border pb-3">
            <p className="text-sm text-muted-foreground">{t('file.size')}</p>
            <p className="font-medium">{formatFileSize(file.size)}</p>
          </div>

          {file.width > 0 && file.height > 0 && (
            <div className="border-b border-border pb-3">
              <p className="text-sm text-muted-foreground">{t('file.dimensions')}</p>
              <p className="font-medium">{file.width} × {file.height}</p>
            </div>
          )}

          <div className="border-b border-border pb-3">
            <p className="text-sm text-muted-foreground">{t('file.takenAt')}</p>
            <p className="font-medium">
              {format(new Date(file.taken_at), 'PPpp')}
            </p>
          </div>

          {file.absolute_path && (
            <div className="border-b border-border pb-3">
              <p className="text-sm text-muted-foreground">{t('file.location')}</p>
              <p className="font-medium text-sm break-all">{file.absolute_path}</p>
            </div>
          )}
        </div>
      </div>
    </div>,
    document.body
  )
}

interface FileGridProps {
  files: File[]
  allFiles?: File[]  // 全局文件列表，用于 lightbox 导航
  selectedFileIds?: number[]
  onSelectionChange?: (selectedIds: number[]) => void
  selectionMode?: boolean
}

export default function FileGrid({
  files,
  allFiles,
  selectedFileIds = [],
  onSelectionChange,
  selectionMode = false,
}: FileGridProps) {
  const [lightboxIndex, setLightboxIndex] = useState(-1)
  const [currentLightboxIndex, setCurrentLightboxIndex] = useState(0)
  const [hoveredFileId, setHoveredFileId] = useState<number | null>(null)
  const [lastSelectedId, setLastSelectedId] = useState<number | null>(null)
  const [isShiftPressed, setIsShiftPressed] = useState(false)
  const [showSidebar, setShowSidebar] = useState(false)
  const [showShareDialog, setShowShareDialog] = useState(false)
  const [shareFile, setShareFile] = useState<File | null>(null)

  // 用于 lightbox 的文件列表（优先使用 allFiles）
  const lightboxFiles = allFiles || files

  // 当前在 lightbox 中显示的文件
  const currentFile = lightboxIndex >= 0 ? lightboxFiles[currentLightboxIndex] : null

  // 转换数据
  const photos = useMemo(() => filesToPhotos(files), [files])
  const slides = useMemo(() => filesToSlides(lightboxFiles), [lightboxFiles])

  // 监听 Shift 键状态
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Shift') setIsShiftPressed(true)
    }
    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.key === 'Shift') setIsShiftPressed(false)
    }

    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('keyup', handleKeyUp)

    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener('keyup', handleKeyUp)
    }
  }, [])

  const isSelected = useCallback(
    (fileId: number) => selectedFileIds.includes(fileId),
    [selectedFileIds]
  )

  const toggleSelection = useCallback(
    (fileId: number, shiftKey: boolean = false) => {
      if (!onSelectionChange) return

      // Shift + Click 范围选择
      if (shiftKey && lastSelectedId !== null) {
        const fileIds = files.map((f) => f.id)
        const lastIndex = fileIds.indexOf(lastSelectedId)
        const currentIndex = fileIds.indexOf(fileId)

        if (lastIndex !== -1 && currentIndex !== -1) {
          const start = Math.min(lastIndex, currentIndex)
          const end = Math.max(lastIndex, currentIndex)
          const rangeIds = fileIds.slice(start, end + 1)
          const newSelection = Array.from(new Set([...selectedFileIds, ...rangeIds]))
          onSelectionChange(newSelection)
          return
        }
      }

      // 普通切换
      const newSelection = selectedFileIds.includes(fileId)
        ? selectedFileIds.filter((id) => id !== fileId)
        : [...selectedFileIds, fileId]

      setLastSelectedId(fileId)
      onSelectionChange(newSelection)
    },
    [files, lastSelectedId, onSelectionChange, selectedFileIds]
  )

  // 计算 Shift 预览范围
  const getRangePreview = useCallback(
    (hoveredId: number): number[] => {
      if (!isShiftPressed || lastSelectedId === null || selectedFileIds.length === 0) {
        return []
      }

      const fileIds = files.map((f) => f.id)
      const lastIndex = fileIds.indexOf(lastSelectedId)
      const hoveredIndex = fileIds.indexOf(hoveredId)

      if (lastIndex === -1 || hoveredIndex === -1) return []

      const start = Math.min(lastIndex, hoveredIndex)
      const end = Math.max(lastIndex, hoveredIndex)
      return fileIds.slice(start, end + 1)
    },
    [files, isShiftPressed, lastSelectedId, selectedFileIds.length]
  )

  const isInPreviewRange = useCallback(
    (fileId: number): boolean => {
      if (!hoveredFileId) return false
      const previewRange = getRangePreview(hoveredFileId)
      return previewRange.includes(fileId) && !isSelected(fileId)
    },
    [getRangePreview, hoveredFileId, isSelected]
  )

  // 自定义渲染图片
  const renderPhoto = useCallback(
    (_props: { onClick?: React.MouseEventHandler }, context: RenderPhotoContext<ExtendedPhoto>) => {
      const { photo, width, height, index } = context
      const file = photo.file
      const selected = isSelected(file.id)
      const inPreviewRange = isInPreviewRange(file.id)
      const isHovered = hoveredFileId === file.id

      const handlePhotoClick = (e: React.MouseEvent) => {
        // 如果在选择模式或已有选中项，处理选择
        if (selectionMode || selectedFileIds.length > 0) {
          e.preventDefault()
          toggleSelection(file.id, e.shiftKey)
          return
        }
        // 否则打开 lightbox，计算在 lightboxFiles 中的索引
        const globalIndex = lightboxFiles.findIndex(f => f.id === file.id)
        setLightboxIndex(globalIndex >= 0 ? globalIndex : index)
      }

      return (
        <div
          style={{ width, height }}
          className={cn(
            'group relative rounded-lg overflow-hidden transition-all cursor-pointer',
            selected ? 'bg-gray-300 dark:bg-gray-700' : 'bg-secondary'
          )}
          onClick={handlePhotoClick}
          onMouseEnter={() => setHoveredFileId(file.id)}
          onMouseLeave={() => setHoveredFileId(null)}
        >
          {/* 缩放容器 */}
          <div
            className={cn(
              'absolute inset-0 transition-transform duration-200 rounded-lg overflow-hidden',
              selected ? 'scale-[0.85]' : 'scale-100'
            )}
          >
            <img
              src={fileAPI.getThumbnailUrl(file.id)}
              alt={file.filename}
              className="w-full h-full object-contain"
              loading="lazy"
            />

            {/* 视频标识 */}
            {file.file_type === 'video' && (
              <div className="absolute inset-0 flex items-center justify-center bg-black/20">
                <Play className="text-white" size={32} />
              </div>
            )}

            {/* 范围预览遮罩 */}
            {inPreviewRange && (
              <div className="absolute inset-0 bg-amber-400/40 pointer-events-none transition-opacity" />
            )}

            {/* 悬停时顶部渐变 */}
            {isHovered && !inPreviewRange && (
              <div className="absolute top-0 left-0 right-0 h-20 bg-gradient-to-b from-black/25 to-transparent transition-opacity" />
            )}

            {/* 底部渐变带文件名 */}
            <div className="absolute bottom-0 left-0 right-0 h-20 bg-gradient-to-t from-black/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
              <div className="absolute bottom-0 left-0 right-0 p-2">
                <p className="text-white text-xs truncate">{file.filename}</p>
              </div>
            </div>
          </div>

          {/* 选择复选框 */}
          {(isHovered || selectedFileIds.length > 0 || selectionMode) && (
            <button
              type="button"
              className={cn(
                'absolute top-2 left-2 p-1 rounded-full flex items-center justify-center cursor-pointer transition-all z-10',
                selected
                  ? 'bg-gray-200 dark:bg-gray-800'
                  : 'bg-white/90 hover:bg-white'
              )}
              onClick={(e) => {
                e.preventDefault()
                e.stopPropagation()
                toggleSelection(file.id, e.shiftKey)
              }}
            >
              {selected ? (
                <svg
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  className="text-blue-600"
                  fill="currentColor"
                >
                  <path d="M12 2C6.5 2 2 6.5 2 12S6.5 22 12 22 22 17.5 22 12 17.5 2 12 2M10 17L5 12L6.41 10.59L10 14.17L17.59 6.58L19 8L10 17Z" />
                </svg>
              ) : (
                <div className="w-6 h-6 rounded-full border-2 border-white" />
              )}
            </button>
          )}
        </div>
      )
    },
    [hoveredFileId, isInPreviewRange, isSelected, lightboxFiles, selectedFileIds.length, selectionMode, toggleSelection]
  )

  return (
    <div className="w-full">
      {/* 照片网格 */}
      <RowsPhotoAlbum
        photos={photos}
        targetRowHeight={225}
        rowConstraints={{ maxPhotos: 6, minPhotos: 1, singleRowMaxHeight: 300 }}
        spacing={18}
        render={{
          photo: renderPhoto,
        }}
      />

      {/* Lightbox */}
      <Lightbox
        open={lightboxIndex >= 0}
        close={() => {
          setLightboxIndex(-1)
          setShowSidebar(false)
        }}
        index={lightboxIndex}
        on={{
          view: ({ index }) => setCurrentLightboxIndex(index),
        }}
        slides={slides}
        plugins={[Video, Thumbnails, Zoom]}
        thumbnails={{
          position: 'bottom',
          width: 120,
          height: 80,
        }}
        carousel={{
          finite: true,
          preload: 2,
        }}
        toolbar={{
          buttons: [
            <div key="custom-toolbar" className="flex items-center gap-1">
              <button
                type="button"
                onClick={() => {
                  if (currentFile) {
                    setShareFile(currentFile)
                    setShowShareDialog(true)
                  }
                }}
                className="yarl__button"
              >
                <Share2 size={24} />
              </button>
              <button
                type="button"
                onClick={() => setShowSidebar(!showSidebar)}
                className="yarl__button"
              >
                <Info size={24} />
              </button>
              <button
                type="button"
                onClick={() => {
                  if (currentFile) {
                    const link = document.createElement('a')
                    link.href = fileAPI.getDownloadUrl(currentFile.id)
                    link.download = currentFile.filename
                    document.body.appendChild(link)
                    link.click()
                    document.body.removeChild(link)
                  }
                }}
                className="yarl__button"
              >
                <Download size={24} />
              </button>
            </div>,
            'close',
          ],
        }}
        className={showSidebar ? 'sidebar-open' : ''}
        styles={{
          container: {
            backgroundColor: 'black',
          },
        }}
      />

      {/* Sidebar - 独立于 Lightbox */}
      {lightboxIndex >= 0 && (
        <LightboxSidebar
          file={currentFile}
          show={showSidebar}
          onClose={() => setShowSidebar(false)}
        />
      )}

      {/* Share Dialog */}
      {showShareDialog && shareFile && (
        <ShareDialog
          fileId={shareFile.id}
          fileName={shareFile.filename}
          onClose={() => {
            setShowShareDialog(false)
            setShareFile(null)
          }}
        />
      )}
    </div>
  )
}
