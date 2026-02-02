import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import FileUploadZone, { UploadProgress } from '@/components/FileUploadZone'
import UploadTargetSelector from '@/components/UploadTargetSelector'
import { uploadService } from '@/services/upload'
import { Upload as UploadIcon, FolderOpen } from 'lucide-react'
import { folderService } from '@/services/folders'
import { useAuth } from '@/contexts/AuthContext'
import { useNavigate } from 'react-router-dom'

export default function Upload() {
  const { t } = useTranslation()
  const { user } = useAuth()
  const navigate = useNavigate()
  const [selectedFiles, setSelectedFiles] = useState<File[]>([])
  const [targetPath, setTargetPath] = useState('')
  const [showTargetSelector, setShowTargetSelector] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [uploadedFiles, setUploadedFiles] = useState<string[]>([])
  const [failedFiles, setFailedFiles] = useState<Array<{ filename: string; error: string }>>([])
  const [noFolders, setNoFolders] = useState(false)
  const [foldersLoading, setFoldersLoading] = useState(true)
  const [foldersError, setFoldersError] = useState<string | null>(null)

  // On mount, check if there are configured folders.
  useEffect(() => {
    const checkFolders = async () => {
      try {
        setFoldersLoading(true)
        setFoldersError(null)
        const resp = await folderService.listFolders()
        const count = (resp.total ?? resp.folders?.length ?? 0)
        setNoFolders(count === 0)
      } catch (e) {
        setFoldersError('Unable to load folders. Please try again.')
        setNoFolders(false)
      }
      setFoldersLoading(false)
    }

    checkFolders()
    // Only run on first render or when role changes
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user?.role])

  const handleFilesSelected = (newFiles: File[]) => {
    setSelectedFiles(prev => [...prev, ...newFiles])
  }

  const handleRemoveFile = (index: number) => {
    setSelectedFiles(prev => prev.filter((_, i) => i !== index))
  }

  const handleSelectTarget = (path: string) => {
    setTargetPath(path)
    setShowTargetSelector(false)
  }

  const handleUpload = async () => {
    if (selectedFiles.length === 0) {
      return
    }

    if (!targetPath) {
      return
    }

    try {
      setUploading(true)
      setUploadedFiles([])
      setFailedFiles([])

      const response = await uploadService.uploadFiles(selectedFiles, targetPath)

      setUploadedFiles(response.uploaded)
      setFailedFiles(response.failed)

      // Remove successfully uploaded files from the list
      if (response.uploaded.length > 0) {
        const uploadedNames = new Set(response.uploaded)
        setSelectedFiles(prev => prev.filter(file => !uploadedNames.has(file.name)))
      }
    } catch (err: any) {
      console.error('Upload failed:', err)
      setFailedFiles([{
        filename: 'Upload Error',
        error: err.response?.data?.error || 'Upload failed. Please try again.'
      }])
    } finally {
      setUploading(false)
    }
  }

  return (
    <div className="container mx-auto px-4 py-8 max-w-5xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2 flex items-center gap-2">
          <UploadIcon className="h-8 w-8" />
          Upload Images
        </h1>
        <p className="text-gray-600">
          Upload your photos to the server and organize them in folders
        </p>
      </div>

      <div className="space-y-6">
        {noFolders && (
          <div className="p-4 bg-blue-50 border border-blue-200 rounded">
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-blue-800 font-medium mb-1">
                  No folders configured yet
                </p>
                <p className="text-xs text-blue-700">
                  To upload files, pick a destination folder. If you’re an admin, create your first folder in Folder Management.
                </p>
              </div>
              {(user?.role === 'admin' || user?.role === 'server_owner') && (
                <button
                  onClick={() => navigate('/folder-management')}
                  className="ml-4 px-3 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded text-sm flex-shrink-0"
                >
                  Add Folder
                </button>
              )}
            </div>
          </div>
        )}

        {/* File Selection */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">1. Select Files</h2>
          <FileUploadZone
            selectedFiles={selectedFiles}
            onFilesSelected={handleFilesSelected}
            onRemoveFile={handleRemoveFile}
          />
        </div>

        {/* Target Directory Selection */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">2. Select Target Directory</h2>

          {!showTargetSelector ? (
            <div>
              {targetPath ? (
                <div className="mb-4 p-4 bg-green-50 border border-green-200 rounded-lg">
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="text-sm font-medium text-green-800 mb-1">
                        Target Directory Selected
                      </p>
                      <p className="text-sm text-green-700 font-mono break-all">
                        {targetPath}
                      </p>
                    </div>
                    <button
                      onClick={() => {
                        if (!foldersError) setShowTargetSelector(true)
                      }}
                      disabled={!!foldersError}
                      className="ml-4 text-sm text-blue-600 hover:text-blue-700 underline flex-shrink-0 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      Change
                    </button>
                  </div>
                </div>
              ) : (
                <div className="text-center py-8 border-2 border-dashed border-gray-300 rounded-lg">
                  <FolderOpen className="h-12 w-12 text-gray-400 mx-auto mb-3" />
                  <p className="text-gray-600 mb-4">No target directory selected</p>
                  <button
                    onClick={() => {
                      if (!noFolders && !foldersError) setShowTargetSelector(true)
                    }}
                    disabled={noFolders || !!foldersError || foldersLoading}
                    className="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    Select Directory
                  </button>
                  {foldersLoading && (
                    <p className="mt-3 text-xs text-gray-500">Checking available folders…</p>
                  )}
                  {foldersError && (
                    <div className="mt-3 text-xs text-red-600">
                      {foldersError}
                      <button
                        onClick={() => {
                          setFoldersError(null)
                          setFoldersLoading(true)
                          ;(async () => {
                            try {
                              const resp = await folderService.listFolders()
                              const count = (resp.total ?? resp.folders?.length ?? 0)
                              setNoFolders(count === 0)
                              setFoldersError(null)
                            } catch (e) {
                              setFoldersError('Unable to load folders. Please try again.')
                            } finally {
                              setFoldersLoading(false)
                            }
                          })()
                        }}
                        className="ml-2 underline text-blue-600 hover:text-blue-700"
                      >
                        Retry
                      </button>
                    </div>
                  )}
                  {noFolders && (
                    <p className="mt-3 text-xs text-gray-500">
                      No folders available yet. {(user?.role === 'admin' || user?.role === 'server_owner') ? (
                        <button
                          onClick={() => navigate('/folder-management')}
                          className="ml-1 underline text-blue-600 hover:text-blue-700"
                        >
                          add a folder
                        </button>
                      ) : (
                        'Please ask an administrator to add one.'
                      )}
                    </p>
                  )}
                </div>
              )}
            </div>
          ) : (
            <div className="space-y-4">
              <UploadTargetSelector
                initialPath={targetPath}
                onSelect={handleSelectTarget}
              />
              <button
                onClick={() => setShowTargetSelector(false)}
                className="text-sm text-gray-600 hover:text-gray-700 underline"
              >
                Cancel
              </button>
            </div>
          )}
        </div>

        {/* Upload Button */}
        <div className="bg-white rounded-lg shadow p-6">
          <h2 className="text-xl font-semibold mb-4">3. Upload</h2>

          {/* Warning message when uploading */}
          {uploading && (
            <div className="mb-4 p-4 bg-yellow-50 border-l-4 border-yellow-400 rounded">
              <div className="flex items-start">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-yellow-700 font-medium">
                    Upload in progress - Do not close this page!
                  </p>
                  <p className="text-xs text-yellow-600 mt-1">
                    Closing the page will interrupt the upload process.
                  </p>
                </div>
              </div>
            </div>
          )}

          <div className="flex items-center justify-between">
            <div className="text-sm text-gray-600">
              {selectedFiles.length > 0 ? (
                <div>
                  <p className="mb-1">
                    Ready to upload <span className="font-semibold">{selectedFiles.length}</span> file(s)
                  </p>
                  {targetPath && (
                    <p className="text-xs text-gray-500">
                      Target: <span className="font-mono bg-gray-100 px-2 py-0.5 rounded">{targetPath}</span>
                    </p>
                  )}
                </div>
              ) : (
                <p>Select files and target directory to begin upload</p>
              )}
            </div>

            <button
              onClick={handleUpload}
              disabled={selectedFiles.length === 0 || !targetPath || uploading}
              className="px-6 py-3 bg-green-500 hover:bg-green-600 text-white rounded-lg font-medium disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 transition-colors"
            >
              {uploading ? (
                <>
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                  Uploading...
                </>
              ) : (
                <>
                  <UploadIcon className="h-5 w-5" />
                  Start Upload
                </>
              )}
            </button>
          </div>
        </div>

        {/* Upload Progress/Results */}
        <UploadProgress
          uploaded={uploadedFiles}
          failed={failedFiles}
          uploading={uploading}
        />
      </div>
    </div>
  )
}
