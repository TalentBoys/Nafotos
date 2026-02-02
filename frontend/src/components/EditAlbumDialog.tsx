import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { albumService, type Album, type AlbumFolder } from '@/services/albums'
import { folderService, type Folder } from '@/services/folders'
import DirectoryTreePickerMulti, { type SelectedFolder } from '@/components/DirectoryTreePickerMulti'

interface EditAlbumDialogProps {
  album: Album
  open: boolean
  onClose: () => void
  onSuccess: () => void
}

export default function EditAlbumDialog({ album, open, onClose, onSuccess }: EditAlbumDialogProps) {
  const { t } = useTranslation()
  const [availableFolders, setAvailableFolders] = useState<Folder[]>([])
  const [formData, setFormData] = useState({
    name: album.name,
    description: album.description || ''
  })
  const [selectedPaths, setSelectedPaths] = useState<SelectedFolder[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (open) {
      loadData()
    }
  }, [open, album.id])

  const loadData = async () => {
    try {
      // Load folders and album folders in parallel
      const [foldersResponse, albumFoldersResponse] = await Promise.all([
        folderService.listFolders(),
        albumService.listAlbumFolders(album.id)
      ])

      const folders = foldersResponse.folders || []
      setAvailableFolders(folders)

      const albumFolders = albumFoldersResponse.data.folders || []
      const initialSelections: SelectedFolder[] = albumFolders.map(folder => {
        const folderInfo = folders.find(f => f.id === folder.folder_id)
        const fullPath = folder.path_prefix
          ? `${folderInfo?.absolute_path}/${folder.path_prefix}`
          : folderInfo?.absolute_path || ''

        return {
          folderId: folder.folder_id,
          folderName: folderInfo?.name || `Folder #${folder.folder_id}`,
          folderPath: folderInfo?.absolute_path || '',
          relativePath: folder.path_prefix,
          fullPath
        }
      })

      setSelectedPaths(initialSelections)
      setFormData({
        name: album.name,
        description: album.description || ''
      })
    } catch (error) {
      console.error('Failed to load data:', error)
    }
  }

  const handleUpdateAlbum = async () => {
    if (!formData.name) return

    try {
      setLoading(true)

      // Update album metadata
      await albumService.updateAlbum(album.id, {
        name: formData.name,
        description: formData.description
      })

      // Get current album folders
      const currentFoldersResponse = await albumService.listAlbumFolders(album.id)
      const currentFolders = currentFoldersResponse.data.folders || []

      // Remove folders that are no longer selected
      for (const folder of currentFolders) {
        const stillSelected = selectedPaths.some(
          s => s.folderId === folder.folder_id && s.relativePath === folder.path_prefix
        )
        if (!stillSelected) {
          await albumService.removeAlbumFolder(album.id, folder.folder_id, folder.path_prefix)
        }
      }

      // Add new folders
      const newFolders = selectedPaths.filter(selected => {
        return !currentFolders.some(
          f => f.folder_id === selected.folderId && f.path_prefix === selected.relativePath
        )
      })

      if (newFolders.length > 0) {
        const foldersToAdd = newFolders.map(path => ({
          folder_id: path.folderId,
          path_prefix: path.relativePath
        }))
        await albumService.addAlbumFolders(album.id, { folders: foldersToAdd })
      }

      alert(t('album.updateSuccess'))
      onSuccess()
      onClose()
    } catch (error) {
      console.error('Failed to update album:', error)
      alert(t('album.updateFailed'))
    } finally {
      setLoading(false)
    }
  }

  const handleFolderSelection = (selections: SelectedFolder[]) => {
    setSelectedPaths(selections)
  }

  const handleClose = () => {
    setFormData({
      name: album.name,
      description: album.description || ''
    })
    setSelectedPaths([])
    onClose()
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <h2 className="text-xl font-bold mb-4">{t('album.editAlbum')}</h2>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">{t('album.name')} *</label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
            />
          </div>

          <div>
            <label className="block text-sm font-medium mb-1">{t('album.description')}</label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              className="w-full px-3 py-2 border rounded-lg dark:bg-gray-700 dark:border-gray-600"
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
            onClick={handleUpdateAlbum}
            disabled={!formData.name || loading}
            className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50"
          >
            {loading ? t('common.loading') : t('common.save')}
          </button>
          <button
            onClick={handleClose}
            disabled={loading}
            className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50"
          >
            {t('common.cancel')}
          </button>
        </div>
      </div>
    </div>
  )
}
