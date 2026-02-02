import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { folderService } from '@/services/folders'
import type { Folder } from '@/services/folders'
import DirectoryTreePicker from '@/components/DirectoryTreePicker'

export default function FolderManagement() {
  const { t } = useTranslation()
  const [folders, setFolders] = useState<Folder[]>([])
  const [selectedFolder, setSelectedFolder] = useState<Folder | null>(null)
  const [loading, setLoading] = useState(true)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [showEditModal, setShowEditModal] = useState(false)
  const [scanning, setScanning] = useState<{ [key: number]: boolean }>({})

  // Form states
  const [folderName, setFolderName] = useState('')
  const [folderPath, setFolderPath] = useState('')
  const [showTreePicker, setShowTreePicker] = useState(false)

  useEffect(() => {
    loadFolders()
  }, [])

  const loadFolders = async () => {
    try {
      setLoading(true)
      const response = await folderService.listFolders()
      setFolders(response.folders || [])
    } catch (err) {
      console.error('Failed to load folders:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleCreateFolder = async () => {
    if (!folderName.trim() || !folderPath.trim()) return

    try {
      await folderService.createFolder({
        name: folderName,
        absolute_path: folderPath
      })
      setShowCreateModal(false)
      setFolderName('')
      setFolderPath('')
      loadFolders()
    } catch (err: any) {
      console.error('Failed to create folder:', err)
      const errorMsg = err.response?.data?.error || t('folderManagement.createError')
      alert(errorMsg)
    }
  }

  const handleUpdateFolder = async () => {
    if (!selectedFolder || !folderName.trim()) return

    try {
      await folderService.updateFolder(selectedFolder.id, {
        name: folderName
      })
      setShowEditModal(false)
      setFolderName('')
      setSelectedFolder(null)
      loadFolders()
    } catch (err) {
      console.error('Failed to update folder:', err)
      alert(t('folderManagement.updateError'))
    }
  }

  const handleDeleteFolder = async (folderId: number) => {
    if (!confirm(t('folderManagement.confirmDelete'))) return

    try {
      await folderService.deleteFolder(folderId)
      if (selectedFolder?.id === folderId) {
        setSelectedFolder(null)
      }
      loadFolders()
    } catch (err) {
      console.error('Failed to delete folder:', err)
      alert(t('folderManagement.deleteError'))
    }
  }

  const handleToggleFolder = async (folderId: number, enabled: boolean) => {
    try {
      await folderService.toggleFolder(folderId, enabled)
      loadFolders()
    } catch (err) {
      console.error('Failed to toggle folder:', err)
    }
  }

  const handleScanFolder = async (folderId: number) => {
    try {
      setScanning(prev => ({ ...prev, [folderId]: true }))
      await folderService.scanFolder(folderId)
      alert(t('folderManagement.scanStarted'))
    } catch (err: any) {
      console.error('Failed to scan folder:', err)
      const errorMsg = err.response?.data?.error || t('folderManagement.scanError')
      alert(errorMsg)
    } finally {
      setScanning(prev => ({ ...prev, [folderId]: false }))
    }
  }

  const openEditModal = (folder: Folder) => {
    setSelectedFolder(folder)
    setFolderName(folder.name)
    setShowEditModal(true)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">{t('folderManagement.title')}</h1>
        <button
          onClick={() => setShowCreateModal(true)}
          className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg"
        >
          {t('folderManagement.addFolder')}
        </button>
      </div>

      <div className="grid grid-cols-1 gap-6">
        {/* Folders List */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-4 border-b">
            <h2 className="text-lg font-semibold">{t('folderManagement.folderList')}</h2>
          </div>
          <div className="p-4">
            {loading ? (
              <div className="text-center py-8">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
              </div>
            ) : folders.length === 0 ? (
              <div className="text-center py-8">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                </svg>
                <p className="mt-4 text-gray-500">{t('folderManagement.noFolders')}</p>
              </div>
            ) : (
              <div className="space-y-3">
                {folders.map((folder) => (
                  <div
                    key={folder.id}
                    className="p-4 border rounded-lg hover:border-blue-300 transition"
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center space-x-3">
                          <h3 className="text-lg font-medium">{folder.name}</h3>
                          <span className={`text-xs px-2 py-1 rounded ${folder.enabled ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-600'}`}>
                            {folder.enabled ? t('folderManagement.enabled') : t('folderManagement.disabled')}
                          </span>
                        </div>
                        <div className="mt-2 flex items-center space-x-2 text-sm text-gray-600">
                          <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                          </svg>
                          <code className="bg-gray-50 px-2 py-1 rounded">{folder.absolute_path}</code>
                        </div>
                        <div className="mt-2 text-xs text-gray-500">
                          {t('folderManagement.createdAt')}: {new Date(folder.created_at).toLocaleString('zh-CN')}
                        </div>
                      </div>
                      <div className="flex items-center space-x-2 ml-4">
                        <button
                          onClick={() => handleScanFolder(folder.id)}
                          disabled={scanning[folder.id]}
                          className={`p-2 rounded hover:bg-blue-50 text-blue-500 ${scanning[folder.id] ? 'opacity-50 cursor-not-allowed' : ''}`}
                          title={t('folderManagement.scan')}
                        >
                          {scanning[folder.id] ? (
                            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-blue-500"></div>
                          ) : (
                            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                            </svg>
                          )}
                        </button>
                        <button
                          onClick={() => handleToggleFolder(folder.id, !folder.enabled)}
                          className="p-2 rounded hover:bg-gray-100 text-gray-600"
                          title={t('folderManagement.toggle')}
                        >
                          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
                          </svg>
                        </button>
                        <button
                          onClick={() => openEditModal(folder)}
                          className="p-2 rounded hover:bg-yellow-50 text-yellow-600"
                          title={t('folderManagement.edit')}
                        >
                          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                          </svg>
                        </button>
                        <button
                          onClick={() => handleDeleteFolder(folder.id)}
                          className="p-2 rounded hover:bg-red-50 text-red-500"
                          title={t('folderManagement.delete')}
                        >
                          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                          </svg>
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Create Folder Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-semibold mb-4">{t('folderManagement.addFolder')}</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('folderManagement.folderName')}
                </label>
                <input
                  type="text"
                  value={folderName}
                  onChange={(e) => setFolderName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder={t('folderManagement.folderNamePlaceholder')}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('folderManagement.absolutePath')}
                </label>
                <div className="flex gap-2">
                  <input
                    type="text"
                    value={folderPath}
                    onChange={(e) => setFolderPath(e.target.value)}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    placeholder={t('folderManagement.absolutePathPlaceholder')}
                  />
                  <button
                    onClick={() => setShowTreePicker(!showTreePicker)}
                    className="px-4 py-2 bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg flex items-center gap-2"
                  >
                    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                    </svg>
                    Browse
                  </button>
                </div>
                <p className="mt-2 text-sm text-gray-500">
                  {t('folderManagement.absolutePathHelp')}
                </p>
              </div>

              {/* Directory Tree Picker */}
              {showTreePicker && (
                <div className="mt-4">
                  <DirectoryTreePicker
                    initialPath={folderPath}
                    onSelect={(path) => {
                      setFolderPath(path)
                      setShowTreePicker(false)
                    }}
                  />
                </div>
              )}
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => {
                  setShowCreateModal(false)
                  setFolderName('')
                  setFolderPath('')
                  setShowTreePicker(false)
                }}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleCreateFolder}
                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
              >
                {t('common.create')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Folder Modal */}
      {showEditModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md">
            <h2 className="text-xl font-semibold mb-4">{t('folderManagement.editFolder')}</h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('folderManagement.folderName')}
                </label>
                <input
                  type="text"
                  value={folderName}
                  onChange={(e) => setFolderName(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder={t('folderManagement.folderNamePlaceholder')}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  {t('folderManagement.absolutePath')}
                </label>
                <input
                  type="text"
                  value={selectedFolder?.absolute_path || ''}
                  disabled
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
                />
                <p className="mt-2 text-sm text-gray-500">
                  {t('folderManagement.pathReadOnly')}
                </p>
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => {
                  setShowEditModal(false)
                  setFolderName('')
                  setSelectedFolder(null)
                }}
                className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200"
              >
                {t('common.cancel')}
              </button>
              <button
                onClick={handleUpdateFolder}
                className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
              >
                {t('common.save')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
