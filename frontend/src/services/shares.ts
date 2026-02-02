import axios from 'axios'

export interface Share {
  id: string
  share_type: 'file' | 'album'
  resource_id: number
  owner_id: number
  access_type: 'public' | 'private'
  requires_auth: boolean
  has_password?: boolean // Frontend helper field to indicate if password is set
  expires_at?: string
  max_views?: number
  view_count: number
  enabled: boolean
  created_at: string
}

export interface CreateShareRequest {
  share_type: 'file' | 'album'
  resource_id: number
  access_type?: 'public' | 'private'
  password?: string
  requires_auth?: boolean
  expires_in?: number // Hours
  max_views?: number
}

export interface UpdateShareRequest {
  enabled?: boolean
  max_views?: number
  password?: string
  requires_auth?: boolean
  expires_in?: number // Hours from now, 0 or negative to remove expiration
}

export interface CreateShareResponse {
  share: Share
  url: string
}

export const shareAPI = {
  // List all shares created by current user
  listShares: () => axios.get<{ shares: Share[]; total: number }>('/api/shares'),

  // Get a specific share
  getShare: (id: string) => axios.get<{ share: Share }>(`/api/shares/${id}`),

  // Create a new share
  createShare: (data: CreateShareRequest) =>
    axios.post<CreateShareResponse>('/api/shares', data),

  // Update share settings
  updateShare: (id: string, data: UpdateShareRequest) =>
    axios.put<{ share: Share }>(`/api/shares/${id}`, data),

  // Delete a share
  deleteShare: (id: string) => axios.delete(`/api/shares/${id}`),

  // Extend share expiration
  extendShare: (id: string, hours: number) =>
    axios.post<{ share: Share }>(`/api/shares/${id}/extend`, { hours }),

  // Get share access log
  getAccessLog: (id: string, limit = 100) =>
    axios.get(`/api/shares/${id}/access-log`, { params: { limit } }),

  // Public share access (no auth required)
  accessShare: (id: string, password?: string) =>
    axios.get<{ share: Share }>(`/api/s/${id}`, {
      params: password ? { password } : undefined
    }),

  // Grant permission to a user for private share
  grantPermission: (shareId: string, userId: number) =>
    axios.post(`/api/shares/${shareId}/permissions`, { user_id: userId }),

  // Revoke permission
  revokePermission: (shareId: string, userId: number) =>
    axios.delete(`/api/shares/${shareId}/permissions/${userId}`),

  // Delete all expired shares
  deleteExpiredShares: () => axios.delete('/api/shares/expired'),
}
