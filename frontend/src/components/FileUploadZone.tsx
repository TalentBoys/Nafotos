import { useState, useRef, DragEvent } from 'react'
import { Upload, X, FileImage, CheckCircle, XCircle } from 'lucide-react'

interface FileUploadZoneProps {
  onFilesSelected: (files: File[]) => void
  selectedFiles: File[]
  onRemoveFile: (index: number) => void
  maxFiles?: number
}

export default function FileUploadZone({
  onFilesSelected,
  selectedFiles,
  onRemoveFile,
  maxFiles = 100
}: FileUploadZoneProps) {
  const [isDragging, setIsDragging] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  // Supported image formats
  const acceptedFormats = [
    '.jpg', '.jpeg', '.png', '.gif', '.bmp',
    '.webp', '.heic', '.heif', '.tif', '.tiff'
  ]
  const acceptString = acceptedFormats.join(',')

  const handleDragEnter = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(true)
  }

  const handleDragLeave = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)
  }

  const handleDragOver = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
  }

  const handleDrop = (e: DragEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDragging(false)

    const files = Array.from(e.dataTransfer.files)
    const imageFiles = files.filter(file => {
      const ext = '.' + file.name.split('.').pop()?.toLowerCase()
      return acceptedFormats.includes(ext)
    })

    if (imageFiles.length > 0) {
      onFilesSelected(imageFiles.slice(0, maxFiles - selectedFiles.length))
    }
  }

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || [])
    if (files.length > 0) {
      onFilesSelected(files.slice(0, maxFiles - selectedFiles.length))
    }
    // Reset input
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const handleClick = () => {
    fileInputRef.current?.click()
  }

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
  }

  return (
    <div className="space-y-4">
      {/* Drop Zone */}
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
          isDragging
            ? 'border-blue-500 bg-blue-50'
            : 'border-gray-300 hover:border-gray-400 bg-white'
        }`}
        onDragEnter={handleDragEnter}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={handleClick}
      >
        <input
          ref={fileInputRef}
          type="file"
          multiple
          accept={acceptString}
          onChange={handleFileInputChange}
          className="hidden"
        />

        <div className="flex flex-col items-center space-y-3">
          <div className="w-16 h-16 bg-gray-100 rounded-full flex items-center justify-center">
            <Upload className="h-8 w-8 text-gray-500" />
          </div>
          <div>
            <p className="text-lg font-medium text-gray-700">
              Drop images here or click to browse
            </p>
            <p className="text-sm text-gray-500 mt-1">
              Supports: JPG, PNG, GIF, BMP, WebP, HEIC, HEIF, TIFF
            </p>
            <p className="text-xs text-gray-400 mt-1">
              Max {maxFiles} files at once
            </p>
          </div>
        </div>
      </div>

      {/* Selected Files List */}
      {selectedFiles.length > 0 && (
        <div className="bg-gray-50 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-medium text-gray-700">
              Selected Files ({selectedFiles.length})
            </h3>
            {selectedFiles.length >= maxFiles && (
              <span className="text-xs text-orange-600 bg-orange-50 px-2 py-1 rounded">
                Maximum files reached
              </span>
            )}
          </div>
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {selectedFiles.map((file, index) => (
              <div
                key={index}
                className="flex items-center justify-between bg-white p-3 rounded border border-gray-200"
              >
                <div className="flex items-center space-x-3 flex-1 min-w-0">
                  <FileImage className="h-5 w-5 text-blue-500 flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-700 truncate">
                      {file.name}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatFileSize(file.size)}
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => onRemoveFile(index)}
                  className="ml-3 p-1 hover:bg-gray-100 rounded flex-shrink-0"
                  title="Remove file"
                >
                  <X className="h-4 w-4 text-gray-500" />
                </button>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

// Upload Progress Component
interface UploadProgressProps {
  uploaded: string[]
  failed: Array<{ filename: string; error: string }>
  uploading: boolean
}

export function UploadProgress({ uploaded, failed, uploading }: UploadProgressProps) {
  // Ensure arrays are not null
  const uploadedFiles = uploaded || []
  const failedFiles = failed || []

  if (!uploading && uploadedFiles.length === 0 && failedFiles.length === 0) {
    return null
  }

  const total = uploadedFiles.length + failedFiles.length
  const hasCompleted = !uploading && total > 0

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h3 className="text-xl font-semibold mb-4">Upload Status</h3>

      {/* Uploading indicator */}
      {uploading && (
        <div className="flex items-center space-x-3 text-blue-600 mb-4 p-4 bg-blue-50 rounded-lg">
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
          <div>
            <p className="font-medium">Uploading files...</p>
            <p className="text-sm text-blue-500">Please wait while your files are being uploaded</p>
          </div>
        </div>
      )}

      {/* Completion summary */}
      {hasCompleted && (
        <div className={`mb-4 p-4 rounded-lg ${
          failedFiles.length === 0
            ? 'bg-green-50 border border-green-200'
            : 'bg-orange-50 border border-orange-200'
        }`}>
          <div className="flex items-start">
            <div className="flex-shrink-0">
              {failedFiles.length === 0 ? (
                <CheckCircle className="h-6 w-6 text-green-500" />
              ) : (
                <svg className="h-6 w-6 text-orange-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
              )}
            </div>
            <div className="ml-3 flex-1">
              <h4 className={`text-sm font-medium ${
                failedFiles.length === 0 ? 'text-green-800' : 'text-orange-800'
              }`}>
                {failedFiles.length === 0
                  ? 'Upload Completed Successfully!'
                  : 'Upload Completed with Issues'
                }
              </h4>
              <div className={`mt-2 text-sm ${
                failedFiles.length === 0 ? 'text-green-700' : 'text-orange-700'
              }`}>
                <p>
                  <span className="font-semibold">{uploadedFiles.length}</span> of <span className="font-semibold">{total}</span> files uploaded successfully
                </p>
                {failedFiles.length > 0 && (
                  <p className="mt-1">
                    <span className="font-semibold">{failedFiles.length}</span> file(s) failed to upload
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Successfully uploaded files */}
      {uploadedFiles.length > 0 && (
        <div className="mb-4">
          <h4 className="text-sm font-medium text-green-600 flex items-center gap-2 mb-2">
            <CheckCircle className="h-4 w-4" />
            Successfully Uploaded ({uploadedFiles.length})
          </h4>
          <div className="bg-gray-50 rounded border border-gray-200 p-3 max-h-40 overflow-y-auto">
            <div className="space-y-1">
              {uploadedFiles.map((filename, index) => (
                <div key={index} className="text-sm text-gray-700 flex items-center gap-2">
                  <span className="text-green-500">✓</span>
                  {filename}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Failed files */}
      {failedFiles.length > 0 && (
        <div>
          <h4 className="text-sm font-medium text-red-600 flex items-center gap-2 mb-2">
            <XCircle className="h-4 w-4" />
            Failed ({failedFiles.length})
          </h4>
          <div className="bg-red-50 rounded border border-red-200 p-3 max-h-40 overflow-y-auto">
            <div className="space-y-2">
              {failedFiles.map((item, index) => (
                <div key={index} className="text-sm">
                  <div className="font-medium text-red-800 flex items-center gap-2">
                    <span className="text-red-500">✗</span>
                    {item.filename}
                  </div>
                  <div className="text-xs text-red-600 ml-5 mt-0.5">
                    {item.error}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
