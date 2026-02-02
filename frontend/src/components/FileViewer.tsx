import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useTranslation } from 'react-i18next'
import { Download, X, Info, Share2 } from 'lucide-react'
import { format } from 'date-fns'
import { fileAPI } from '@/services/api'
import type { File } from '@/types'
import ShareDialog from '@/components/ShareDialog'

interface FileViewerProps {
  file: File
  mode?: 'detail' | 'share'
  onClose?: () => void
  getImageUrl?: (fileId: number) => string
  getDownloadUrl?: (fileId: number) => string
}

export default function FileViewer({
  file,
  mode = 'detail',
  onClose,
  getImageUrl,
  getDownloadUrl
}: FileViewerProps) {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const [showOriginal, setShowOriginal] = useState(false)
  const [showSidebar, setShowSidebar] = useState(false)
  const [showShareDialog, setShowShareDialog] = useState(false)

  // Define 10MB threshold
  const TEN_MB = 10 * 1024 * 1024
  const isLargeFile = file.size > TEN_MB

  // Determine what to display
  const getImageSource = () => {
    if (mode === 'share' && getImageUrl) {
      // For share mode, use provided URL
      return getImageUrl(file.id)
    }

    if (showOriginal || !isLargeFile) {
      // Show original image for small files or when user clicks "View Original"
      return fileAPI.getDownloadUrl(file.id)
    } else {
      // Show large thumbnail for large files
      return fileAPI.getThumbnailUrl(file.id, 'large')
    }
  }

  const getVideoSource = () => {
    if (mode === 'share' && getImageUrl) {
      return getImageUrl(file.id)
    }
    return fileAPI.getDownloadUrl(file.id)
  }

  const handleClose = () => {
    if (onClose) {
      onClose()
    } else {
      navigate(-1)
    }
  }

  const handleDownload = () => {
    const downloadUrl = mode === 'share' && getDownloadUrl
      ? getDownloadUrl(file.id)
      : fileAPI.getDownloadUrl(file.id)

    // Create a temporary anchor element to trigger download
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = file.filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <div className="fixed inset-0 bg-black flex">
      {/* Main content area - fullscreen image */}
      <div className="flex-1 relative flex items-center justify-center">
        {/* Close button */}
        <button
          onClick={handleClose}
          className="absolute top-4 left-4 z-10 p-2 rounded-lg bg-black/50 text-white hover:bg-black/70 transition-colors"
          title={t('common.back')}
        >
          <X size={24} />
        </button>

        {/* Action buttons in top right */}
        <div className="absolute top-4 right-4 z-10 flex items-center gap-2">
          {/* Share button - only show in detail mode */}
          {mode === 'detail' && (
            <button
              onClick={() => setShowShareDialog(true)}
              className="p-2 rounded-lg bg-black/50 text-white hover:bg-black/70 transition-colors"
              title={t('file.share')}
            >
              <Share2 size={20} />
            </button>
          )}

          {/* Info button */}
          <button
            onClick={() => setShowSidebar(!showSidebar)}
            className="p-2 rounded-lg bg-black/50 text-white hover:bg-black/70 transition-colors"
            title={t('file.details')}
          >
            <Info size={20} />
          </button>

          {/* Download button */}
          <button
            onClick={handleDownload}
            className="p-2 rounded-lg bg-black/50 text-white hover:bg-black/70 transition-colors"
            title={t('file.download')}
          >
            <Download size={20} />
          </button>
        </div>

        {/* Image or video content */}
        <div className="w-full h-full flex items-center justify-center p-4 relative">
          {file.file_type === 'image' ? (
            <>
              <img
                src={getImageSource()}
                alt={file.filename}
                className="max-w-full max-h-full object-contain"
              />
              {/* View original button floating over image - only in detail mode for large files */}
              {mode === 'detail' && isLargeFile && !showOriginal && (
                <button
                  onClick={() => setShowOriginal(true)}
                  className="absolute bottom-8 left-1/2 -translate-x-1/2 px-6 py-3 bg-black/70 text-white rounded-lg hover:bg-black/90 transition-colors text-sm font-medium border-2 border-white/30 shadow-lg"
                >
                  {t('file.viewOriginal')}
                </button>
              )}
            </>
          ) : (
            <video
              src={getVideoSource()}
              controls
              className="max-w-full max-h-full"
            />
          )}
        </div>
      </div>

      {/* Sidebar */}
      {showSidebar && (
        <div className="w-80 bg-background border-l border-border overflow-y-auto flex-shrink-0">
          <div className="p-6 space-y-6">
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
                  <p className="text-sm text-muted-foreground">
                    {t('file.dimensions')}
                  </p>
                  <p className="font-medium">
                    {file.width} × {file.height}
                  </p>
                </div>
              )}

              <div className="border-b border-border pb-3">
                <p className="text-sm text-muted-foreground">{t('file.takenAt')}</p>
                <p className="font-medium">
                  {format(new Date(file.taken_at), 'PPpp')}
                </p>
              </div>

              {/* Only show absolute path in detail mode */}
              {mode === 'detail' && file.absolute_path && (
                <div className="border-b border-border pb-3">
                  <p className="text-sm text-muted-foreground">{t('file.location')}</p>
                  <p className="font-medium text-sm break-all">{file.absolute_path}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Share Dialog - only in detail mode */}
      {mode === 'detail' && showShareDialog && file && (
        <ShareDialog
          fileId={file.id}
          fileName={file.filename}
          onClose={() => setShowShareDialog(false)}
        />
      )}
    </div>
  )
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
