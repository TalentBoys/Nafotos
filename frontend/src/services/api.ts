import axios from 'axios'
import type { File, Album, Tag, MountPoint, PaginatedResponse, User } from '@/types'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  withCredentials: true, // Important: send cookies with requests
})

// Add response interceptor to handle 401 errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Redirect to login if not authenticated
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const fileAPI = {
  getFiles: (page = 1, limit = 50, type?: string) =>
    api.get<PaginatedResponse<File>>('/files', { params: { page, limit, type } }),

  getFileById: (id: number) =>
    api.get<File>(`/files/${id}`),

  getTimeline: (page = 1, limit = 50) =>
    api.get<PaginatedResponse<File>>('/timeline', { params: { page, limit } }),

  searchFiles: (query: string) =>
    api.get<{ files: File[] }>('/search', { params: { q: query } }),

  getThumbnailUrl: (id: number, size: 'small' | 'medium' | 'large' = 'small') =>
    `/api/files/${id}/thumbnail?size=${size}`,

  getDownloadUrl: (id: number) => `/api/files/${id}/download`,
}

export const albumAPI = {
  getAlbums: () =>
    api.get<{ albums: Album[] }>('/albums'),

  createAlbum: (data: Partial<Album>) =>
    api.post<Album>('/albums', data),
}

export const tagAPI = {
  getTags: () =>
    api.get<{ tags: Tag[] }>('/tags'),

  createTag: (data: Partial<Tag>) =>
    api.post<Tag>('/tags', data),
}

export const mountPointAPI = {
  getMountPoints: () =>
    api.get<{ mount_points: MountPoint[] }>('/mount-points'),
}

export const systemAPI = {
  triggerScan: () =>
    api.post('/scan'),

  healthCheck: () =>
    api.get('/health'),
}

export const userAPI = {
  searchUsers: (query: string, limit = 10) =>
    api.get<{ users: User[], total: number }>('/users/search', { params: { q: query, limit } }),
}

export default api
