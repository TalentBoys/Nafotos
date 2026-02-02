export interface File {
  id: number
  filename: string
  file_type: 'image' | 'video'
  size: number
  width: number
  height: number
  taken_at: string
  created_at: string
  updated_at: string
  absolute_path?: string  // Computed from folder + relative_path
  thumbnail_url?: string
}

export interface Album {
  id: number
  name: string
  description: string
  cover_file_id: number
  created_at: string
  updated_at: string
}

export interface Tag {
  id: number
  name: string
  color: string
  created_at: string
}

export interface MountPoint {
  id: number
  path: string
  name: string
  enabled: boolean
  created_at: string
}

export interface PaginatedResponse<T> {
  files?: T[]
  page: number
  limit: number
}

export interface User {
  id: number
  username: string
  email: string
  role: string
  enabled: boolean
  created_at: string
  updated_at: string
}

export interface Folder {
  id: number
  name: string
  absolute_path: string
  enabled: boolean
  created_by: number
  created_at: string
  updated_at: string
}
