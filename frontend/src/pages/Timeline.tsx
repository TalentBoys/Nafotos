import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { fileAPI } from '@/services/api'
import type { File } from '@/types'
import FileGrid from '@/components/FileGrid'
import SelectionBar from '@/components/SelectionBar'
import { format } from 'date-fns'
import { Upload } from 'lucide-react'

export default function Timeline() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(true)
  const [page, setPage] = useState(1)
  const [hasMore, setHasMore] = useState(true)
  const [selectedFileIds, setSelectedFileIds] = useState<number[]>([])

  useEffect(() => {
    loadFiles()
  }, [page])

  const loadFiles = async () => {
    try {
      setLoading(true)
      const response = await fileAPI.getTimeline(page, 50)
      const newFiles = response.data.files || []

      if (page === 1) {
        setFiles(newFiles)
      } else {
        setFiles((prev) => [...prev, ...newFiles])
      }

      setHasMore(newFiles.length === 50)
    } catch (error) {
      console.error('Failed to load files:', error)
    } finally {
      setLoading(false)
    }
  }

  const groupFilesByDate = (files: File[]) => {
    const groups: Record<string, File[]> = {}

    files.forEach((file) => {
      const date = format(new Date(file.taken_at), 'yyyy-MM-dd')
      if (!groups[date]) {
        groups[date] = []
      }
      groups[date].push(file)
    })

    return groups
  }

  const fileGroups = groupFilesByDate(files)

  const handleShare = () => {
    navigate(`/share?fileIds=${selectedFileIds.join(',')}`)
  }

  const handleClearSelection = () => {
    setSelectedFileIds([])
  }

  return (
    <div className="space-y-8">
      {selectedFileIds.length > 0 && (
        <SelectionBar
          selectedCount={selectedFileIds.length}
          onClear={handleClearSelection}
          onShare={handleShare}
        />
      )}

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('timeline.title')}</h1>
          <p className="text-muted-foreground">{t('app.title')}</p>
        </div>
        <button
          onClick={() => navigate('/upload')}
          className="flex items-center gap-2 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors"
        >
          <Upload className="h-5 w-5" />
          Upload Images
        </button>
      </div>

      {loading && page === 1 ? (
        <div className="text-center py-12">{t('timeline.loading')}</div>
      ) : files.length === 0 ? (
        <div className="text-center py-12">
          <div className="text-muted-foreground mb-4">
            {t('timeline.empty')}
          </div>
          <button
            onClick={() => navigate('/upload')}
            className="inline-flex items-center gap-2 px-6 py-3 bg-blue-500 hover:bg-blue-600 text-white rounded-lg font-medium transition-colors"
          >
            <Upload className="h-5 w-5" />
            Upload Your First Images
          </button>
        </div>
      ) : (
        <div className="space-y-8">
          {Object.entries(fileGroups).map(([date, dateFiles]) => (
            <div key={date}>
              <h2 className="text-xl font-semibold mb-4">
                {format(new Date(date), 'MMMM d, yyyy')}
              </h2>
              <FileGrid
                files={dateFiles}
                allFiles={files}
                selectedFileIds={selectedFileIds}
                onSelectionChange={setSelectedFileIds}
              />
            </div>
          ))}

          {hasMore && (
            <div className="text-center">
              <button
                onClick={() => setPage((p) => p + 1)}
                disabled={loading}
                className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
              >
                {loading ? t('timeline.loading') : t('timeline.loadMore')}
              </button>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
