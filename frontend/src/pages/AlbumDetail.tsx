import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useParams, useNavigate } from 'react-router-dom'
import { albumService, type Album } from '@/services/albums'
import type { File } from '@/types'
import FileGrid from '@/components/FileGrid'
import SelectionBar from '@/components/SelectionBar'
import EditAlbumDialog from '@/components/EditAlbumDialog'
import { ArrowLeft, Settings } from 'lucide-react'

export default function AlbumDetail() {
  const { t } = useTranslation()
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [album, setAlbum] = useState<Album | null>(null)
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(true)
  const [sortOrder, setSortOrder] = useState<string>('taken_at DESC')
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [selectedFileIds, setSelectedFileIds] = useState<number[]>([])

  useEffect(() => {
    if (id) {
      loadAlbumAndFiles()
    }
  }, [id, sortOrder])

  const loadAlbumAndFiles = async () => {
    if (!id) return

    try {
      setLoading(true)
      const albumId = parseInt(id, 10)

      // Load album details and files in parallel
      const [albumResponse, filesResponse] = await Promise.all([
        albumService.getAlbum(albumId),
        albumService.listAlbumItems(albumId, sortOrder)
      ])

      setAlbum(albumResponse.data.album)
      setFiles(filesResponse.data.files || [])
    } catch (error) {
      console.error('Failed to load album:', error)
      alert(t('album.loadFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleSortChange = (newSortOrder: string) => {
    setSortOrder(newSortOrder)
  }

  const handleEditSuccess = () => {
    loadAlbumAndFiles()
  }

  const handleShare = () => {
    navigate(`/share?fileIds=${selectedFileIds.join(',')}`)
  }

  const handleClearSelection = () => {
    setSelectedFileIds([])
  }

  if (loading) {
    return <div className="text-center py-12">{t('common.loading')}</div>
  }

  if (!album) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground mb-4">{t('album.notFound')}</p>
        <button
          onClick={() => navigate('/albums')}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90"
        >
          {t('common.back')}
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {selectedFileIds.length > 0 && (
        <SelectionBar
          selectedCount={selectedFileIds.length}
          onClear={handleClearSelection}
          onShare={handleShare}
        />
      )}

      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <button
            onClick={() => navigate('/albums')}
            className="flex items-center gap-2 text-muted-foreground hover:text-foreground mb-4"
          >
            <ArrowLeft size={20} />
            {t('common.back')}
          </button>
          <h1 className="text-3xl font-bold mb-2">{album.name}</h1>
          {album.description && (
            <p className="text-muted-foreground">{album.description}</p>
          )}
          <p className="text-sm text-muted-foreground mt-2">
            {t('album.fileCount', { count: files.length })}
          </p>
        </div>

        <button
          onClick={() => setShowEditDialog(true)}
          className="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-secondary"
        >
          <Settings size={18} />
          {t('common.edit')}
        </button>
      </div>

      {/* Sort Controls */}
      <div className="flex items-center gap-4">
        <label className="text-sm font-medium">{t('album.sortBy')}:</label>
        <select
          value={sortOrder}
          onChange={(e) => handleSortChange(e.target.value)}
          className="px-3 py-1.5 border rounded-lg dark:bg-gray-800 dark:border-gray-700"
        >
          <option value="taken_at DESC">{t('album.sortByDateDesc')}</option>
          <option value="taken_at ASC">{t('album.sortByDateAsc')}</option>
          <option value="filename ASC">{t('album.sortByNameAsc')}</option>
          <option value="filename DESC">{t('album.sortByNameDesc')}</option>
        </select>
      </div>

      {/* File Grid */}
      {files.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-muted-foreground">
            {t('album.noFiles')}
          </p>
        </div>
      ) : (
        <FileGrid
          files={files}
          selectedFileIds={selectedFileIds}
          onSelectionChange={setSelectedFileIds}
        />
      )}

      {/* Edit Album Dialog */}
      {album && (
        <EditAlbumDialog
          album={album}
          open={showEditDialog}
          onClose={() => setShowEditDialog(false)}
          onSuccess={handleEditSuccess}
        />
      )}
    </div>
  )
}
