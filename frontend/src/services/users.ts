import api from './api'

export interface User {
  id: number
  username: string
  email: string
  role: 'server_owner' | 'admin' | 'user'
  enabled: boolean
  created_at: string
  updated_at: string
  last_login_at?: string
  password_changed_at?: string
}

export interface UserActivityLog {
  id: number
  user_id: number
  performed_by: number
  action: string
  details: string
  ip_address: string
  created_at: string
}

export interface PaginatedUsers {
  users: User[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface PaginatedActivityLogs {
  logs: UserActivityLog[]
  total: number
  page: number
  limit: number
  total_pages: number
}

export interface UserStats {
  total_users: number
  active_users: number
  admins: number
  disabled_users: number
}

export const userAPI = {
  // List with pagination and filters
  listUsers: (params?: {
    page?: number
    limit?: number
    search?: string
    role?: string
  }) => api.get<PaginatedUsers>('/users', { params }),

  // CRUD operations
  getUser: (id: number) =>
    api.get<{ user: User }>(`/users/${id}`),

  createUser: (data: {
    username: string
    password: string
    email: string
    role: string
  }) =>
    api.post<{ user: User }>('/users', data),

  updateUser: (id: number, data: {
    email?: string
    role?: string
    enabled?: boolean
  }) =>
    api.put<{ user: User }>(`/users/${id}`, data),

  deleteUser: (id: number) =>
    api.delete(`/users/${id}`),

  toggleUser: (id: number, enabled: boolean) =>
    api.put(`/users/${id}/toggle`, { enabled }),

  // Password management
  resetPassword: (id: number, newPassword: string) =>
    api.post(`/users/${id}/reset-password`, { new_password: newPassword }),

  // Bulk operations
  bulkEnableDisable: (userIds: number[], enabled: boolean) =>
    api.post('/users/bulk/enable-disable', { user_ids: userIds, enabled }),

  bulkDelete: (userIds: number[]) =>
    api.post('/users/bulk/delete', { user_ids: userIds }),

  // Activity logs
  getUserActivityLogs: (userId: number, page?: number, limit?: number) =>
    api.get<PaginatedActivityLogs>(
      `/users/${userId}/activity-logs`,
      { params: { page, limit } }
    ),

  // Export
  exportUsers: () =>
    api.post('/users/export', {}, { responseType: 'blob' }),

  // Statistics
  getUserStats: () =>
    api.get<UserStats>('/users/stats'),
}
