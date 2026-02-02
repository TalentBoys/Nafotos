import api from './api'

// Types
export interface Album {
  id: number
  name: string
  description: string
  owner_id: number
  cover_file_id?: number
  created_at: string
  updated_at: string
}

export interface AlbumFolder {
  id: number
  album_id: number
  folder_id: number
  path_prefix: string
  added_at: string
}

export interface FolderConfig {
  folder_id: number
  path_prefix: string
}

export interface AlbumItem {
  id: number
  album_id: number
  folder_id: number
  relative_path: string
  file_id?: number
  added_at: string
}

export interface CreateAlbumRequest {
  name: string
  description?: string
  folders?: FolderConfig[]
}

export interface AddFoldersRequest {
  folders: FolderConfig[]
}

// Album Service
export const albumService = {
  // List all albums
  listAlbums: () =>
    api.get<{ albums: Album[], total: number }>('/albums-v2'),

  // Get album details
  getAlbum: (id: number) =>
    api.get<{ album: Album }>(`/albums-v2/${id}`),

  // Create album (with optional folder configurations)
  createAlbum: (data: CreateAlbumRequest) =>
    api.post<{ album: Album }>('/albums-v2', data),

  // Update album
  updateAlbum: (id: number, data: Partial<Album>) =>
    api.put<{ album: Album }>(`/albums-v2/${id}`, data),

  // Delete album
  deleteAlbum: (id: number) =>
    api.delete<{ message: string }>(`/albums-v2/${id}`),

  // List album items (files) - now returns files directly with optional sort parameter
  listAlbumItems: (id: number, sort: string = 'taken_at DESC') =>
    api.get<{ files: any[], total: number }>(`/albums-v2/${id}/items`, {
      params: { sort }
    }),

  // List folder configurations
  listAlbumFolders: (id: number) =>
    api.get<{ folders: AlbumFolder[], total: number }>(`/albums-v2/${id}/folders`),

  // Add folder configurations
  addAlbumFolders: (id: number, data: AddFoldersRequest) =>
    api.post<{ message: string }>(`/albums-v2/${id}/folders`, data),

  // Remove folder configuration
  removeAlbumFolder: (albumId: number, folderId: number, pathPrefix: string = '') =>
    api.delete<{ message: string }>(`/albums-v2/${albumId}/folders/${folderId}`, {
      params: { path_prefix: pathPrefix }
    }),
}
