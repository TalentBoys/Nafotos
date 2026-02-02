import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { albumService, type Album, type FolderConfig } from '@/services/albums'
import { folderService, type Folder } from '@/services/folders'
import DirectoryTreePickerMulti, { type SelectedFolder } from '@/components/DirectoryTreePickerMulti'
import EditAlbumDialog from '@/components/EditAlbumDialog'
import {
  Image,
  Plus,
  Edit2,
  Trash2
} from 'lucide-react'

export default function Albums() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [albums, setAlbums] = useState<Album[]>([])
  const [loading, setLoading] = useState(true)
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showEditDialog, setShowEditDialog] = useState(false)
  const [selectedAlbum, setSelectedAlbum] = useState<Album | null>(null)
  const [availableFolders, setAvailableFolders] = useState<Folder[]>([])

  // Form states for create dialog
  const [formData, setFormData] = useState({
    name: '',
    description: ''
  })
  const [selectedPaths, setSelectedPaths] = useState<SelectedFolder[]>([])

  useEffect(() => {
    loadAlbums()
    loadFolders()
  }, [])

  const loadAlbums = async () => {
    try {
      const response = await albumService.listAlbums()
      setAlbums(response.data.albums || [])
    } catch (error) {
      console.error('Failed to load albums:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadFolders = async () => {
    try {
      const response = await folderService.listFolders()
      setAvailableFolders(response.folders || [])
    } catch (error) {
      console.error('Failed to load folders:', error)
    }
  }

  const handleCreateAlbum = async () => {
    if (!formData.name) return

    try {
      const folders = selectedPaths.map(path => ({
        folder_id: path.folderId,
        path_prefix: path.relativePath
      }))

      await albumService.createAlbum({
        ...formData,
        folders
      })
      setShowCreateDialog(false)
      setFormData({ name: '', description: '' })
      setSelectedPaths([])
      loadAlbums()
      alert(t('album.createSuccess'))
    } catch (error) {
      console.error('Failed to create album:', error)
      alert(t('album.createFailed'))
    }
  }

  const handleDeleteAlbum = async (album: Album) => {
    if (!confirm(t('album.confirmDelete', { name: album.name }))) return

    try {
      await albumService.deleteAlbum(album.id)
      loadAlbums()
    } catch (error) {
      console.error('Failed to delete album:', error)
      alert(t('album.deleteFailed'))
    }
  }

  const openEditDialog = (album: Album) => {
    setSelectedAlbum(album)
    setShowEditDialog(true)
  }

  const handleEditSuccess = () => {
    loadAlbums()
  }

  const handleFolderSelection = (selections: SelectedFolder[]) => {
    setSelectedPaths(selections)
  }

  if (loading) {
    return <div className="text-center py-12">{t('common.loading')}</div>
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">{t('album.management')}</h1>
          <p className="text-muted-foreground">{t('album.managementDescription')}</p>
        </div>
        <button
          onClick={() => setShowCreateDialog(true)}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 flex items-center gap-2"
        >
          <Plus size={18} />
          {t('album.create')}
        </button>
      </div>

      {albums.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          {t('album.noAlbums')}
        </div>
      ) : (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {albums.map((album) => (
            <div
              key={album.id}
              onClick={() => navigate(`/albums/${album.id}`)}
              className="border rounded-lg overflow-hidden hover:shadow-lg transition-shadow cursor-pointer"
            >
              <div className="aspect-square bg-secondary flex items-center justify-center">
                <Image className="text-muted-foreground" size={48} />
              </div>
              <div className="p-4">
                <h3 className="font-semibold truncate">{album.name}</h3>
                {album.description && (
                  <p className="text-sm text-muted-foreground truncate">
                    {album.description}
                  </p>
                )}
                <div className="flex gap-2 mt-3">
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      openEditDialog(album)
                    }}
                    className="flex-1 px-2 py-1 text-xs bg-gray-500 text-white rounded hover:bg-gray-600"
                    title={t('common.edit')}
                  >
                    <Edit2 size={14} className="inline" />
                  </button>
                  <button
                    onClick={(e) => {
                      e.stopPropagation()
                      handleDeleteAlbum(album)
                    }}
                    className="flex-1 px-2 py-1 text-xs bg-red-500 text-white rounded hover:bg-red-600"
                    title={t('common.delete')}
                  >
                    <Trash2 size={14} className="inline" />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Album Dialog */}
      {showCreateDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">{t('album.createAlbum')}</h2>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">{t('album.name')} *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
                  placeholder={t('album.namePlaceholder')}
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">{t('album.description')}</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
                  placeholder={t('album.descriptionPlaceholder')}
                  rows={3}
                />
              </div>

              <div>
                <label className="block text-sm font-medium mb-2">{t('album.folderConfiguration')}</label>
                <DirectoryTreePickerMulti
                  onSelect={handleFolderSelection}
                  initialSelections={selectedPaths}
                />
              </div>
            </div>

            <div className="flex gap-3 mt-6">
              <button
                onClick={handleCreateAlbum}
                disabled={!formData.name}
                className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {t('common.create')}
              </button>
              <button
                onClick={() => {
                  setShowCreateDialog(false)
                  setFormData({ name: '', description: '' })
                  setSelectedPaths([])
                }}
                className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                {t('common.cancel')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Album Dialog */}
      {selectedAlbum && (
        <EditAlbumDialog
          album={selectedAlbum}
          open={showEditDialog}
          onClose={() => {
            setShowEditDialog(false)
            setSelectedAlbum(null)
          }}
          onSuccess={handleEditSuccess}
        />
      )}

    </div>
  )
}
