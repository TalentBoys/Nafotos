import api from './api'
import { DirectoryInfo } from './folders'

export interface UploadResponse {
  message: string
  uploaded: string[]
  uploaded_count: number
  failed: Array<{
    filename: string
    error: string
  }>
  failed_count: number
  total: number
}

export const uploadService = {
  // Upload files to a target directory
  uploadFiles: async (files: File[], targetPath: string): Promise<UploadResponse> => {
    const formData = new FormData()
    formData.append('target_path', targetPath)

    files.forEach(file => {
      formData.append('files', file)
    })

    const response = await api.post('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
    return response.data
  },

  // Browse upload target (configured folders and subdirectories)
  browseUploadTarget: async (path: string, folderId?: number): Promise<{ path: string; directories: DirectoryInfo[] }> => {
    const response = await api.post('/upload/browse', { path, folder_id: folderId })
    return response.data
  },

  // Create a new directory
  createDirectory: async (parentPath: string, directoryName: string): Promise<{ message: string; path: string }> => {
    const response = await api.post('/upload/create-directory', {
      parent_path: parentPath,
      directory_name: directoryName
    })
    return response.data
  }
}
