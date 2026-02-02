import { useState, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
import { useNavigate } from 'react-router-dom'
import { folderService } from '@/services/folders'
import type { Folder } from '@/services/folders'
import type { File } from '@/types'
import { FolderOpen, ChevronRight } from 'lucide-react'
import FileGrid from '@/components/FileGrid'
import SelectionBar from '@/components/SelectionBar'

export default function Folders() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const [folders, setFolders] = useState<Folder[]>([])
  const [selectedFolder, setSelectedFolder] = useState<Folder | null>(null)
  const [files, setFiles] = useState<File[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingFiles, setLoadingFiles] = useState(false)
  const [selectedFileIds, setSelectedFileIds] = useState<number[]>([])

  useEffect(() => {
    loadFolders()
  }, [])

  const loadFolders = async () => {
    try {
      setLoading(true)
      const response = await folderService.listFolders()
      setFolders(response.folders || [])
    } catch (error) {
      console.error('Failed to load folders:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadFolderFiles = async (folderId: number) => {
    try {
      setLoadingFiles(true)
      const response = await folderService.listFilesInFolder(folderId)
      setFiles(response.files || [])
    } catch (error) {
      console.error('Failed to load folder files:', error)
      setFiles([])
    } finally {
      setLoadingFiles(false)
    }
  }

  const handleFolderClick = (folder: Folder) => {
    setSelectedFolder(folder)
    setSelectedFileIds([])
    loadFolderFiles(folder.id)
  }

  const handleShare = () => {
    navigate(`/share?fileIds=${selectedFileIds.join(',')}`)
  }

  const handleClearSelection = () => {
    setSelectedFileIds([])
  }

  if (loading) {
    return <div className="text-center py-12">{t('timeline.loading')}</div>
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

      <div>
        <h1 className="text-3xl font-bold mb-2">{t('folderBrowser.title')}</h1>
        <p className="text-muted-foreground">{t('folderBrowser.description')}</p>
      </div>

      {folders.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          {t('folderBrowser.noFolders')}
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
          {/* Folders List */}
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow p-4">
              <h2 className="text-lg font-semibold mb-4">{t('folderManagement.folderList')}</h2>
              <div className="space-y-2">
                {folders
                  .filter((folder) => folder.enabled)
                  .map((folder) => (
                    <button
                      key={folder.id}
                      onClick={() => handleFolderClick(folder)}
                      className={`w-full text-left p-3 rounded-lg transition ${
                        selectedFolder?.id === folder.id
                          ? 'bg-blue-50 border-2 border-blue-500'
                          : 'bg-gray-50 hover:bg-gray-100 border-2 border-transparent'
                      }`}
                    >
                      <div className="flex items-center space-x-2">
                        <FolderOpen size={20} className="text-blue-500 flex-shrink-0" />
                        <div className="flex-1 min-w-0">
                          <div className="font-medium truncate">{folder.name}</div>
                          <div className="text-xs text-gray-500 truncate">
                            {folder.absolute_path}
                          </div>
                        </div>
                        <ChevronRight
                          size={16}
                          className={`flex-shrink-0 transition-transform ${
                            selectedFolder?.id === folder.id ? 'rotate-90' : ''
                          }`}
                        />
                      </div>
                    </button>
                  ))}
              </div>
            </div>
          </div>

          {/* Files Grid */}
          <div className="lg:col-span-3">
            {selectedFolder ? (
              <div>
                <div className="mb-4">
                  <h2 className="text-2xl font-bold">{selectedFolder.name}</h2>
                  <p className="text-sm text-gray-600 mt-1">
                    {selectedFolder.absolute_path}
                  </p>
                </div>

                {loadingFiles ? (
                  <div className="text-center py-12">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
                  </div>
                ) : files.length === 0 ? (
                  <div className="text-center py-12 text-gray-500">
                    {t('folderBrowser.noFiles')}
                  </div>
                ) : (
                  <FileGrid
                    files={files}
                    selectedFileIds={selectedFileIds}
                    onSelectionChange={setSelectedFileIds}
                  />
                )}
              </div>
            ) : (
              <div className="flex items-center justify-center h-full">
                <div className="text-center text-gray-500">
                  <FolderOpen size={48} className="mx-auto mb-4 text-gray-400" />
                  <p>{t('folderBrowser.selectFolder')}</p>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
